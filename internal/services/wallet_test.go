package services

import (
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ptop/internal/models"
)

func TestGetAddressFake(t *testing.T) {
	t.Setenv("DEBUG_FAKE_NETWORK", "true")
	db, err := gorm.Open(sqlite.Open("file:test_fake?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	if err := db.AutoMigrate(&models.Asset{}, &models.Wallet{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	asset := models.Asset{Name: "BTC_fake", Type: models.AssetTypeCrypto, IsActive: true}
	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("asset: %v", err)
	}
	addr, idx, err := GetAddress(db, "clientA", asset.ID)
	if err != nil {
		t.Fatalf("get address: %v", err)
	}
	exp := fmt.Sprintf("fake:%s:%s:%d", asset.ID, "clientA", 0)
	if addr != exp || idx != 0 {
		t.Fatalf("expected %s idx 0, got %s idx %d", exp, addr, idx)
	}
	w := models.Wallet{ClientID: "clientA", AssetID: asset.ID, Value: addr, DerivationIndex: idx, IsEnabled: true}
	if err := db.Create(&w).Error; err != nil {
		t.Fatalf("wallet: %v", err)
	}
	addr2, idx2, err := GetAddress(db, "clientA", asset.ID)
	if err != nil {
		t.Fatalf("second address: %v", err)
	}
	exp2 := fmt.Sprintf("fake:%s:%s:%d", asset.ID, "clientA", 1)
	if addr2 != exp2 || idx2 != 1 {
		t.Fatalf("expected %s idx 1, got %s idx %d", exp2, addr2, idx2)
	}
}
