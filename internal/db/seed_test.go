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
