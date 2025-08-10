package ethwatcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"ptop/internal/models"
)

var transferSigHash = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

// tokenInfo содержит информацию о токене ERC20.
type tokenInfo struct {
	assetID  string
	decimals int32
}

// Watcher отслеживает блоки Ethereum и события Transfer выбранных токенов.
type Watcher struct {
	client  *ethclient.Client
	db      *gorm.DB
	tokens  map[common.Address]*tokenInfo
	debug   bool
	debugCh chan debugDeposit
}

type debugDeposit struct {
	walletID string
	amount   decimal.Decimal
}

// New создаёт новый наблюдатель.
func New(db *gorm.DB, rpcURL string, debug bool) (*Watcher, error) {
	tokens := make(map[common.Address]*tokenInfo)

	var usdtAsset models.Asset
	if err := db.Where("name = ?", "USDT").First(&usdtAsset).Error; err == nil {
		tokens[common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")] = &tokenInfo{assetID: usdtAsset.ID, decimals: 6}
	}
	var usdcAsset models.Asset
	if err := db.Where("name = ?", "USDC").First(&usdcAsset).Error; err == nil {
		tokens[common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9EB0cE3606eB48")] = &tokenInfo{assetID: usdcAsset.ID, decimals: 6}
	}

	w := &Watcher{db: db, tokens: tokens, debug: debug}
	if debug {
		w.debugCh = make(chan debugDeposit)
		return w, nil
	}
	if rpcURL == "" {
		return nil, fmt.Errorf("eth rpc url required")
	}
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}
	w.client = client
	return w, nil
}

// Start запускает подписку на новые блоки и события Transfer.
func (w *Watcher) Start() error {
	if w.debug {
		go w.debugLoop()
		return nil
	}
	heads := make(chan *types.Header)
	sub, err := w.client.SubscribeNewHead(context.Background(), heads)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case err := <-sub.Err():
				log.Printf("ошибка подписки на блоки: %v", err)
				return
			case head := <-heads:
				go w.handleBlock(head.Hash(), head.Number.Uint64())
			}
		}
	}()

	if len(w.tokens) > 0 {
		addresses := make([]common.Address, 0, len(w.tokens))
		for addr := range w.tokens {
			addresses = append(addresses, addr)
		}
		query := ethereum.FilterQuery{
			Addresses: addresses,
			Topics:    [][]common.Hash{{transferSigHash}},
		}
		logsCh := make(chan types.Log)
		logSub, err := w.client.SubscribeFilterLogs(context.Background(), query, logsCh)
		if err != nil {
			return err
		}
		go w.handleLogs(logSub, logsCh)
	}

	return nil
}

func (w *Watcher) debugLoop() {
	for dep := range w.debugCh {
		w.createDebugDeposit(dep.walletID, dep.amount)
	}
}

// TriggerDeposit отправляет фейковый депозит для тестов.
func (w *Watcher) TriggerDeposit(walletID string, amount decimal.Decimal) {
	if w.debug {
		w.debugCh <- debugDeposit{walletID: walletID, amount: amount}
	}
}

func (w *Watcher) createDebugDeposit(walletID string, amount decimal.Decimal) {
	var wallet models.Wallet
	if err := w.db.First(&wallet, "id = ?", walletID).Error; err != nil {
		log.Printf("ошибка поиска кошелька: %v", err)
		return
	}
	data, _ := json.Marshal(map[string]any{"debug": true})
	dep := models.TransactionIn{
		ClientID: wallet.ClientID,
		WalletID: wallet.ID,
		AssetID:  wallet.AssetID,
		Amount:   amount,
		Status:   models.TransactionInStatusConfirmed,
		Data:     datatypes.JSON(data),
	}
	if err := w.db.Create(&dep).Error; err != nil {
		log.Printf("не удалось сохранить депозит: %v", err)
		return
	}
	if err := w.db.Model(&models.Balance{}).
		Where("client_id = ? AND asset_id = ?", wallet.ClientID, wallet.AssetID).
		Update("amount", gorm.Expr("amount + ?", amount)).Error; err != nil {
		log.Printf("не удалось обновить баланс: %v", err)
	}
}

func (w *Watcher) handleBlock(hash common.Hash, number uint64) {
	block, err := w.client.BlockByHash(context.Background(), hash)
	if err != nil {
		log.Printf("не удалось получить блок %s: %v", hash.Hex(), err)
		return
	}
	for _, tx := range block.Transactions() {
		w.processTx(tx, number)
	}
}

func (w *Watcher) processTx(tx *types.Transaction, blockNumber uint64) {
	to := tx.To()
	if to == nil || tx.Value().Sign() == 0 {
		return
	}
	var wallet models.Wallet
	if err := w.db.Where("value = ?", to.Hex()).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
		log.Printf("ошибка базы данных: %v", err)
		return
	}
	var existing models.TransactionIn
	if err := w.db.Where("data ->> 'tx_hash' = ?", tx.Hash().Hex()).First(&existing).Error; err == nil {
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("ошибка проверки существующей транзакции: %v", err)
		return
	}
	amount := decimal.NewFromBigInt(tx.Value(), -18)
	data, _ := json.Marshal(map[string]any{
		"tx_hash":      tx.Hash().Hex(),
		"block_number": blockNumber,
	})
	dep := models.TransactionIn{
		ClientID: wallet.ClientID,
		WalletID: wallet.ID,
		AssetID:  wallet.AssetID,
		Amount:   amount,
		Status:   "confirmed",
		Data:     datatypes.JSON(data),
	}
	if err := w.db.Create(&dep).Error; err != nil {
		log.Printf("не удалось сохранить депозит: %v", err)
		return
	}
	if err := w.db.Model(&models.Balance{}).
		Where("client_id = ? AND asset_id = ?", wallet.ClientID, wallet.AssetID).
		Update("amount", gorm.Expr("amount + ?", amount)).Error; err != nil {
		log.Printf("не удалось обновить баланс: %v", err)
	}
}

func (w *Watcher) handleLogs(sub ethereum.Subscription, logsCh <-chan types.Log) {
	for {
		select {
		case err := <-sub.Err():
			log.Printf("ошибка подписки на логи: %v", err)
			return
		case vLog := <-logsCh:
			w.processLog(vLog)
		}
	}
}

func (w *Watcher) processLog(vLog types.Log) {
	info, ok := w.tokens[vLog.Address]
	if !ok || len(vLog.Topics) < 3 {
		return
	}
	to := common.HexToAddress(vLog.Topics[2].Hex())
	var wallet models.Wallet
	if err := w.db.Where("value = ? AND asset_id = ?", to.Hex(), info.assetID).First(&wallet).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
		log.Printf("ошибка базы данных: %v", err)
		return
	}
	var existing models.TransactionIn
	if err := w.db.Where("data ->> 'tx_hash' = ? AND data ->> 'log_index' = ?", vLog.TxHash.Hex(), fmt.Sprintf("%d", vLog.Index)).First(&existing).Error; err == nil {
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("ошибка проверки существующей транзакции: %v", err)
		return
	}
	value := new(big.Int).SetBytes(vLog.Data)
	amount := decimal.NewFromBigInt(value, -info.decimals)
	data, _ := json.Marshal(map[string]any{
		"tx_hash":      vLog.TxHash.Hex(),
		"log_index":    vLog.Index,
		"block_number": vLog.BlockNumber,
		"token":        vLog.Address.Hex(),
	})
	dep := models.TransactionIn{
		ClientID: wallet.ClientID,
		WalletID: wallet.ID,
		AssetID:  wallet.AssetID,
		Amount:   amount,
		Status:   "confirmed",
		Data:     datatypes.JSON(data),
	}
	if err := w.db.Create(&dep).Error; err != nil {
		log.Printf("не удалось сохранить депозит: %v", err)
		return
	}
	if err := w.db.Model(&models.Balance{}).
		Where("client_id = ? AND asset_id = ?", wallet.ClientID, wallet.AssetID).
		Update("amount", gorm.Expr("amount + ?", amount)).Error; err != nil {
		log.Printf("не удалось обновить баланс: %v", err)
	}
}
