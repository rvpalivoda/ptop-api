package xmrwatcher

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ptop/internal/models"
)

func TestWatcherDebugDeposit(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:xmr_debug?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	if err := db.AutoMigrate(&models.Client{}, &models.Asset{}, &models.Wallet{}, &models.TransactionIn{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	client := models.Client{Username: "u"}
	db.Create(&client)
	asset := models.Asset{Name: "XMR"}
	db.Create(&asset)
	wallet := models.Wallet{ClientID: client.ID, AssetID: asset.ID, Value: "subaddr", DerivationIndex: 2}
	db.Create(&wallet)

	w, err := New(db, "", time.Second, true)
	if err != nil {
		t.Fatalf("watcher: %v", err)
	}
	w.Start()
	w.TriggerDeposit(wallet.ID, decimal.RequireFromString("4"))
	time.Sleep(50 * time.Millisecond)
	var tx models.TransactionIn
	if err := db.First(&tx).Error; err != nil {
		t.Fatalf("tx: %v", err)
	}
	if !tx.Amount.Equal(decimal.RequireFromString("4")) {
		t.Fatalf("amount %s", tx.Amount)
	}
}
