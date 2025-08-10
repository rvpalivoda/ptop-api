package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ptop/internal/btcwatcher"
	"ptop/internal/models"
)

func TestDebugDeposit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file:test_debug?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	if err := db.AutoMigrate(&models.Client{}, &models.Asset{}, &models.Wallet{}, &models.TransactionIn{}, &models.Balance{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	client := models.Client{Username: "u"}
	if err := db.Create(&client).Error; err != nil {
		t.Fatalf("create client: %v", err)
	}
	asset := models.Asset{Name: "BTC"}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("create asset: %v", err)
	}
	wallet := models.Wallet{ClientID: client.ID, AssetID: asset.ID, Value: "addr", DerivationIndex: 1}
	if err := db.Create(&wallet).Error; err != nil {
		t.Fatalf("create wallet: %v", err)
	}
	bal := models.Balance{ClientID: client.ID, AssetID: asset.ID, Amount: decimal.Zero, AmountEscrow: decimal.Zero}
	if err := db.Create(&bal).Error; err != nil {
		t.Fatalf("create balance: %v", err)
	}
	w, err := btcwatcher.New(db, "", "", "", nil, true)
	if err != nil {
		t.Fatalf("watcher: %v", err)
	}
	if err := w.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	watchers := map[string]DebugDepositor{"BTC": w}
	r := gin.Default()
	r.POST("/debug/deposit", DebugDeposit(db, watchers))

	body, _ := json.Marshal(map[string]string{"wallet_id": wallet.ID, "amount": "1"})
	req := httptest.NewRequest(http.MethodPost, "/debug/deposit", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Code != http.StatusNoContent {
		t.Fatalf("status %d", resp.Code)
	}
	time.Sleep(50 * time.Millisecond)
	var tx models.TransactionIn
	if err := db.First(&tx).Error; err != nil {
		t.Fatalf("tx: %v", err)
	}
	if !tx.Amount.Equal(decimal.RequireFromString("1")) {
		t.Fatalf("amount %s", tx.Amount)
	}
	if err := db.First(&bal).Error; err != nil {
		t.Fatalf("balance: %v", err)
	}
	if !bal.Amount.Equal(decimal.RequireFromString("1")) {
		t.Fatalf("balance amount %s", bal.Amount)
	}
}
