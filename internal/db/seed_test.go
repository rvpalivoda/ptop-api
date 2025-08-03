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
	if err := gdb.AutoMigrate(&models.Country{}, &models.PaymentMethod{}, &models.Asset{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := SeedCountries(gdb); err != nil {
		t.Fatalf("seed countries: %v", err)
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
	if pmCount != 22 || assetCount != 10 {
		t.Fatalf("expected 22 methods and 10 assets, got %d and %d", pmCount, assetCount)
	}
	var assets []models.Asset
	if err := gdb.Find(&assets).Error; err != nil {
		t.Fatalf("list assets: %v", err)
	}
	want := []string{"USD", "EUR", "UAH", "GBP", "PLN", "BTC", "ETH", "USDT", "USDC", "XMR"}
	names := make(map[string]bool)
	for _, a := range assets {
		names[a.Name] = true
	}
	for _, n := range want {
		if !names[n] {
			t.Fatalf("missing asset %s", n)
		}
	}

	var methods []models.PaymentMethod
	flexepinCountries := append([]string{"Canada", "Australia"}, regionCountries["EU"]...)
	if err := gdb.Preload("Countries").Find(&methods).Error; err != nil {
		t.Fatalf("list methods: %v", err)
	}
	wantCountries := map[string][]string{
		"Interac e-Transfer": {"Canada"},
		"SPEI":               {"Mexico"},
		"RTP (Network)":      {"United States"},
		"FedNow":             {"United States"},
		"NPP / Osko":         {"Australia"},
		"PayNow (FAST)":      {"Singapore"},
		"PromptPay":          {"Thailand"},
		"DuitNow":            {"Malaysia"},
		"FPS (Hong Kong)":    {"Hong Kong (Special Administrative Region of China)"},
		"SBP (Система быстрых платежей)": {"Russian Federation"},
		"Orange Money":           {"Senegal", "Cote d'Ivoire", "Mali", "Burkina Faso", "Benin"},
		"MTN MoMo":               {"Ghana", "Uganda", "South Africa"},
		"TIPS (pan-EU rail)":     regionCountries["EEA"],
		"Alipay":                 {"China"},
		"WeChat Pay":             {"China"},
		"Cash App":               {"United States"},
		"Venmo":                  {"United States"},
		"Ria Money Transfer":     regionCountries["GLOBAL"],
		"WorldRemit cash pickup": regionCountries["GLOBAL"],
		"Flexepin code":          flexepinCountries,
		"SEPA":                   regionCountries["EU"],
		"SWIFT":                  regionCountries["GLOBAL"],
	}
	for _, m := range methods {
		expected, ok := wantCountries[m.Name]
		if !ok {
			t.Fatalf("unexpected method %s", m.Name)
		}
		if len(m.Countries) != len(expected) {
			t.Fatalf("method %s expected %d countries, got %d", m.Name, len(expected), len(m.Countries))
		}
		got := make(map[string]bool)
		for _, c := range m.Countries {
			got[c.Name] = true
		}
		for _, name := range expected {
			if !got[name] {
				t.Fatalf("method %s missing country %s", m.Name, name)
			}
		}
	}
	var mtn models.PaymentMethod
	if err := gdb.Where("name = ?", "MTN MoMo").First(&mtn).Error; err != nil {
		t.Fatalf("get MTN MoMo: %v", err)
	}
	if mtn.TypicalFiatCCY != "GHS/UGX/ZAR" {
		t.Fatalf("expected MTN MoMo TypicalFiatCCY to be GHS/UGX/ZAR, got %s", mtn.TypicalFiatCCY)
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
