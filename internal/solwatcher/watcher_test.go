package solwatcher

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ptop/internal/models"
)

func TestWatcherDebugDeposit(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:sol_debug?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	if err := db.AutoMigrate(&models.Client{}, &models.Asset{}, &models.Wallet{}, &models.TransactionIn{}, &models.Balance{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	client := models.Client{Username: "u"}
	db.Create(&client)
	asset := models.Asset{Name: "USDC"}
	db.Create(&asset)
	wallet := models.Wallet{ClientID: client.ID, AssetID: asset.ID, Value: "addr", DerivationIndex: 1}
	db.Create(&wallet)
	bal := models.Balance{ClientID: client.ID, AssetID: asset.ID, Amount: decimal.Zero, AmountEscrow: decimal.Zero}
	db.Create(&bal)

	w, err := New(db, "", "", true)
	if err != nil {
		t.Fatalf("watcher: %v", err)
	}
	if err := w.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	w.TriggerDeposit(wallet.ID, decimal.RequireFromString("2"))
	time.Sleep(50 * time.Millisecond)
	var tx models.TransactionIn
	if err := db.First(&tx).Error; err != nil {
		t.Fatalf("tx: %v", err)
	}
	if !tx.Amount.Equal(decimal.RequireFromString("2")) {
		t.Fatalf("amount %s", tx.Amount)
	}
	if err := db.First(&bal).Error; err != nil {
		t.Fatalf("balance: %v", err)
	}
	if !bal.Amount.Equal(decimal.RequireFromString("2")) {
		t.Fatalf("balance amount %s", bal.Amount)
	}
}

func TestWatcherMintRequiredInProd(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:sol_prod?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	if _, err := New(db, "", "", false); err == nil {
		t.Fatalf("expected error")
	}

}
