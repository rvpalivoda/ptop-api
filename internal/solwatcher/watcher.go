package solwatcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"

	solana "github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	ws "github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"ptop/internal/models"
)

// Watcher отслеживает переводы USDC в сети Solana и сохраняет депозиты.
type Watcher struct {
	wsClient  *ws.Client
	rpcClient *rpc.Client
	db        *gorm.DB
	mint      solana.PublicKey
	debug     bool
	debugCh   chan debugDeposit
}

type debugDeposit struct {
	walletID string
	amount   decimal.Decimal
}

// New создаёт нового наблюдателя.
func New(db *gorm.DB, rpcURL, mintAddress string, debug bool) (*Watcher, error) {
	w := &Watcher{db: db, debug: debug}
	if mintAddress == "" {
		return nil, fmt.Errorf("mint address required")
	}
	mint, err := solana.PublicKeyFromBase58(mintAddress)
	if err != nil {
		return nil, fmt.Errorf("mint address: %w", err)
	}
	w.mint = mint
	if debug {
		w.debugCh = make(chan debugDeposit)
		return w, nil
	}
	if rpcURL == "" {
		return nil, fmt.Errorf("solana rpc url required")
	}
	ctx := context.Background()
	wsClient, err := ws.Connect(ctx, rpcURL)
	if err != nil {
		return nil, err
	}
	w.wsClient = wsClient
	httpURL := rpcURL
	if strings.HasPrefix(httpURL, "wss://") {
		httpURL = "https://" + strings.TrimPrefix(httpURL, "wss://")
	} else if strings.HasPrefix(httpURL, "ws://") {
		httpURL = "http://" + strings.TrimPrefix(httpURL, "ws://")
	}
	w.rpcClient = rpc.New(httpURL)
	return w, nil
}

// Start запускает подписку на логи программы SPL Token.
func (w *Watcher) Start() error {
	if w.debug {
		go w.debugLoop()
		return nil
	}
	sub, err := w.wsClient.LogsSubscribe(ws.LogsSubscribeFilterAll, rpc.CommitmentFinalized)
	if err != nil {
		return err
	}
	go w.handleLogs(sub)
	return nil
}

func (w *Watcher) handleLogs(sub *ws.LogSubscription) {
	ctx := context.Background()
	for {
		res, err := sub.Recv(ctx)
		if err != nil {
			log.Printf("sol logs recv: %v", err)
			return
		}
		if res.Value.Err != nil {
			continue
		}
		w.processSignature(res.Value.Signature)
	}
}

func (w *Watcher) processSignature(sig solana.Signature) {
	tx, err := w.rpcClient.GetParsedTransaction(context.Background(), sig, nil)
	if err != nil || tx == nil || tx.Transaction == nil {
		return
	}
	for _, inst := range tx.Transaction.Message.Instructions {
		if inst.Program != "spl-token" || inst.Parsed == nil {
			continue
		}
		var parsed rpc.InstructionInfo
		pData, err := json.Marshal(inst.Parsed)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(pData, &parsed); err != nil {
			continue
		}
		mintStr, _ := parsed.Info["mint"].(string)
		if mintStr != w.mint.String() {
			continue
		}
		dest, _ := parsed.Info["destination"].(string)
		if dest == "" {
			continue
		}
		var amountStr string
		if v, ok := parsed.Info["amount"].(string); ok {
			amountStr = v
		} else if ta, ok := parsed.Info["tokenAmount"].(map[string]interface{}); ok {
			if am, ok := ta["amount"].(string); ok {
				amountStr = am
			}
		}
		if amountStr == "" {
			continue
		}
		var wallet models.Wallet
		if err := w.db.Where("value = ?", dest).First(&wallet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			log.Printf("db error: %v", err)
			continue
		}
		var existing models.TransactionIn
		if err := w.db.Where("data ->> 'signature' = ?", sig.String()).First(&existing).Error; err == nil {
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("existing check error: %v", err)
			continue
		}
		amtBig, ok := new(big.Int).SetString(amountStr, 10)
		if !ok {
			continue
		}
		amount := decimal.NewFromBigInt(amtBig, -6)
		data, _ := json.Marshal(map[string]any{"signature": sig.String(), "slot": tx.Slot})
		dep := models.TransactionIn{
			ClientID: wallet.ClientID,
			WalletID: wallet.ID,
			AssetID:  wallet.AssetID,
			Amount:   amount,
			Status:   models.TransactionInStatusConfirmed,
			Data:     datatypes.JSON(data),
		}
		if err := w.db.Create(&dep).Error; err != nil {
			log.Printf("failed to save deposit: %v", err)
		}
	}
}

func (w *Watcher) debugLoop() {
	for dep := range w.debugCh {
		w.createDebugDeposit(dep.walletID, dep.amount)
	}
}

// TriggerDeposit отправляет фейковый депозит.
func (w *Watcher) TriggerDeposit(walletID string, amount decimal.Decimal) {
	if w.debug {
		w.debugCh <- debugDeposit{walletID: walletID, amount: amount}
	}
}

func (w *Watcher) createDebugDeposit(walletID string, amount decimal.Decimal) {
	var wal models.Wallet
	if err := w.db.First(&wal, "id = ?", walletID).Error; err != nil {
		log.Printf("ошибка поиска кошелька: %v", err)
		return
	}
	data, _ := json.Marshal(map[string]any{"debug": true})
	dep := models.TransactionIn{
		ClientID: wal.ClientID,
		WalletID: wal.ID,
		AssetID:  wal.AssetID,
		Amount:   amount,
		Status:   models.TransactionInStatusConfirmed,
		Data:     datatypes.JSON(data),
	}
	if err := w.db.Create(&dep).Error; err != nil {
		log.Printf("не удалось сохранить депозит: %v", err)
	}
}

// TriggerSignature используется только для тестов, чтобы обработать сигнатуру.
func (w *Watcher) TriggerSignature(sig solana.Signature) {
	if !w.debug {
		w.processSignature(sig)
	}
}
