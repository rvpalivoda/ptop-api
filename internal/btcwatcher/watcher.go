package btcwatcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"ptop/internal/models"
)

// Watcher следит за блоками Bitcoin и записывает подтвержденные депозиты.
type Watcher struct {
	client *rpcclient.Client
	db     *gorm.DB
	params *chaincfg.Params
}

// New создаёт нового наблюдателя.
func New(db *gorm.DB, cfg *rpcclient.ConnConfig, params *chaincfg.Params) (*Watcher, error) {
	if params == nil {
		params = &chaincfg.MainNetParams
	}
	ntfnHandlers := rpcclient.NotificationHandlers{}
	w := &Watcher{db: db, params: params}
	ntfnHandlers.OnBlockConnected = func(hash *chainhash.Hash, height int32, t time.Time) {
		go w.handleBlock(hash)
	}
	client, err := rpcclient.New(cfg, &ntfnHandlers)
	if err != nil {
		return nil, err
	}
	w.client = client
	return w, nil
}

// Start запускает подписку на новые блоки.
func (w *Watcher) Start() error {
	if err := w.client.NotifyBlocks(); err != nil {
		return fmt.Errorf("notify blocks: %w", err)
	}
	return nil
}

// handleBlock обрабатывает подключенный блок.
func (w *Watcher) handleBlock(hash *chainhash.Hash) {
	block, err := w.client.GetBlock(hash)
	if err != nil {
		log.Printf("не удалось получить блок %s: %v", hash, err)
		return
	}
	for _, tx := range block.Transactions {
		w.processTx(tx, hash.String())
	}
}

func (w *Watcher) processTx(tx *wire.MsgTx, blockHash string) {
	txid := tx.TxHash().String()
	for i, out := range tx.TxOut {
		_, addrs, _, err := txscript.ExtractPkScriptAddrs(out.PkScript, w.params)
		if err != nil || len(addrs) == 0 {
			continue
		}
		for _, addr := range addrs {
			var wallet models.Wallet
			if err := w.db.Where("value = ?", addr.EncodeAddress()).First(&wallet).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				log.Printf("ошибка базы данных: %v", err)
				continue
			}
			var existing models.TransactionIn
			if err := w.db.Where("data ->> 'txid' = ? AND data ->> 'vout' = ?", txid, fmt.Sprintf("%d", i)).First(&existing).Error; err == nil {
				continue
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("ошибка проверки существующей транзакции: %v", err)
				continue
			}
			amount := decimal.NewFromInt(out.Value).Div(decimal.NewFromInt(1e8))
			data, _ := json.Marshal(map[string]any{
				"txid":       txid,
				"vout":       i,
				"block_hash": blockHash,
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
			}
		}
	}
}
