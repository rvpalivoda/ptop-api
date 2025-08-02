package db

import (
	"testing"

	"github.com/biter777/countries"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ptop/internal/models"
)

func TestSeedCountries(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := gdb.AutoMigrate(&models.Country{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := SeedCountries(gdb); err != nil {
		t.Fatalf("seed: %v", err)
	}
	var count int64
	if err := gdb.Model(&models.Country{}).Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if int(count) != countries.Total() {
		t.Fatalf("expected %d countries, got %d", countries.Total(), count)
	}
	if err := SeedCountries(gdb); err != nil {
		t.Fatalf("reseeding: %v", err)
	}
	var count2 int64
	gdb.Model(&models.Country{}).Count(&count2)
	if count2 != count {
		t.Fatalf("expected no duplicates after reseed; got %d vs %d", count2, count)
	}
}

func TestSeedPaymentMethodsAndAssets(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := gdb.AutoMigrate(&models.PaymentMethod{}, &models.Asset{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := SeedPaymentMethods(gdb); err != nil {
		t.Fatalf("seed payment methods: %v", err)
	}
	if err := SeedAssets(gdb); err != nil {
		t.Fatalf("seed assets: %v", err)
	}
	var pmCount, assetCount int64
	gdb.Model(&models.PaymentMethod{}).Count(&pmCount)
	gdb.Model(&models.Asset{}).Count(&assetCount)
	if pmCount != 2 || assetCount != 2 {
		t.Fatalf("expected 2 methods and 2 assets, got %d and %d", pmCount, assetCount)
	}
	if err := SeedPaymentMethods(gdb); err != nil {
		t.Fatalf("reseeding methods: %v", err)
	}
	if err := SeedAssets(gdb); err != nil {
		t.Fatalf("reseeding assets: %v", err)
	}
	var pmCount2, assetCount2 int64
	gdb.Model(&models.PaymentMethod{}).Count(&pmCount2)
	gdb.Model(&models.Asset{}).Count(&assetCount2)
	if pmCount2 != pmCount || assetCount2 != assetCount {
		t.Fatalf("expected no duplicates after reseed")
	}
}
