package xmrwatcher

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/omani/go-monero-rpc-client/wallet"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"ptop/internal/models"
)

// Watcher отслеживает входящие транзакции Monero и сохраняет их в базе данных.
type Watcher struct {
	client       wallet.Client
	db           *gorm.DB
	pollInterval time.Duration
}

// New создаёт нового наблюдателя.
func New(db *gorm.DB, rpcURL string, interval time.Duration) *Watcher {
	if interval == 0 {
		interval = time.Minute
	}
	cl := wallet.New(wallet.Config{Address: rpcURL})
	return &Watcher{client: cl, db: db, pollInterval: interval}
}

// Start запускает периодический опрос get_transfers.
func (w *Watcher) Start() {
	go w.check()
	ticker := time.NewTicker(w.pollInterval)
	go func() {
		for range ticker.C {
			w.check()
		}
	}()
}

func (w *Watcher) check() {
	// Получаем все кошельки XMR.
	var wallets []models.Wallet
	if err := w.db.Joins("JOIN assets ON wallets.asset_id = assets.id").Where("assets.name LIKE ?", "XMR%").Find(&wallets).Error; err != nil {
		log.Printf("ошибка базы данных: %v", err)
		return
	}
	if len(wallets) == 0 {
		return
	}
	subMap := make(map[uint64]models.Wallet, len(wallets))
	indices := make([]uint64, 0, len(wallets))
	for _, wal := range wallets {
		subMap[uint64(wal.DerivationIndex)] = wal
		indices = append(indices, uint64(wal.DerivationIndex))
	}

	res, err := w.client.GetTransfers(&wallet.RequestGetTransfers{
		In:             true,
		Pending:        true,
		AccountIndex:   0,
		SubaddrIndices: indices,
	})
	if err != nil {
		log.Printf("get_transfers error: %v", err)
		return
	}

	for _, tr := range res.Pending {
		w.handleTransfer(tr, subMap, models.TransactionInStatusPending)
	}
	for _, tr := range res.In {
		w.handleTransfer(tr, subMap, models.TransactionInStatusConfirmed)
	}
}

func (w *Watcher) handleTransfer(tr *wallet.Transfer, subMap map[uint64]models.Wallet, status models.TransactionInStatus) {
	wal, ok := subMap[tr.SubaddrIndex.Minor]
	if !ok {
		return
	}
	var existing models.TransactionIn
	err := w.db.Where("data ->> 'txid' = ? AND data ->> 'subaddr_index' = ?", tr.TxID, fmt.Sprintf("%d", tr.SubaddrIndex.Minor)).First(&existing).Error
	if err == nil {
		if existing.Status != status && status == models.TransactionInStatusConfirmed {
			data, _ := json.Marshal(map[string]any{
				"txid":          tr.TxID,
				"subaddr_index": tr.SubaddrIndex.Minor,
				"confirmations": tr.Confirmations,
			})
			if err := w.db.Model(&existing).Updates(map[string]any{"status": status, "data": datatypes.JSON(data)}).Error; err != nil {
				log.Printf("не удалось обновить депозит: %v", err)
			}
		}
		return
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("ошибка проверки существующей транзакции: %v", err)
		return
	}
	amount := decimal.NewFromInt(int64(tr.Amount)).Div(decimal.NewFromInt(1e12))
	data, _ := json.Marshal(map[string]any{
		"txid":          tr.TxID,
		"subaddr_index": tr.SubaddrIndex.Minor,
		"confirmations": tr.Confirmations,
	})
	dep := models.TransactionIn{
		ClientID: wal.ClientID,
		WalletID: wal.ID,
		AssetID:  wal.AssetID,
		Amount:   amount,
		Status:   status,
		Data:     datatypes.JSON(data),
	}
	if err := w.db.Create(&dep).Error; err != nil {
		log.Printf("не удалось сохранить депозит: %v", err)
	}
}
