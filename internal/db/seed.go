package db

import (
	"github.com/biter777/countries"
	"gorm.io/gorm"

	"ptop/internal/models"
)

// SeedCountries заполняет таблицу стран перечнем всех стран на английском языке.
func SeedCountries(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Country{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var list []models.Country
	for _, code := range countries.All() {
		list = append(list, models.Country{Name: code.String()})
	}
	return db.Create(&list).Error
}

// SeedPaymentMethods добавляет базовые платёжные методы, если таблица пуста.
func SeedPaymentMethods(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.PaymentMethod{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	methods := []models.PaymentMethod{
		{
			Name:              "Interac",
			MethodGroup:       "bank_transfer",
			Provider:          "Interac",
			TypicalFiatCCY:    "CAD",
			Regions:           []string{"CA"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 5,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintMedium,
		},
		{
			Name:              "SPEI",
			MethodGroup:       "bank_transfer",
			Provider:          "SPEI",
			TypicalFiatCCY:    "MXN",
			Regions:           []string{"MX"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 5,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintMedium,
		},
		{
			Name:              "RTP",
			MethodGroup:       "bank_transfer",
			Provider:          "TCH",
			TypicalFiatCCY:    "USD",
			Regions:           []string{"US"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintHigh,
		},
		{
			Name:              "Orange Money",
			MethodGroup:       "mobile_money",
			Provider:          "Orange",
			TypicalFiatCCY:    "XOF",
			Regions:           []string{"SN", "CI", "ML", "BF", "BJ"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideReceiver,
			KycLevelHint:      models.KycLevelHintLow,
		},
		{
			Name:                  "SEPA",
			MethodGroup:           "bank_transfer",
			Provider:              "EBA",
			TypicalFiatCCY:        "EUR",
			Regions:               []string{"EU"},
			IsRealtime:            false,
			IsReversible:          true,
			SettlementMinutes:     1440,
			ChargebackWindowHours: 720,
			FeeSide:               models.FeeSideShared,
			KycLevelHint:          models.KycLevelHintMedium,
		},
		{
			Name:                  "SWIFT",
			MethodGroup:           "bank_transfer",
			Provider:              "SWIFT",
			TypicalFiatCCY:        "USD",
			Regions:               []string{"GLOBAL"},
			IsRealtime:            false,
			IsReversible:          true,
			SettlementMinutes:     2880,
			ChargebackWindowHours: 720,
			FeeSide:               models.FeeSideShared,
			KycLevelHint:          models.KycLevelHintHigh,
		},
	}
	return db.Create(&methods).Error
}

// SeedAssets добавляет базовые активы, если таблица пуста.
func SeedAssets(db *gorm.DB) error {
	var count int64
	if err := db.Model(&models.Asset{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	assets := []models.Asset{
		// fiat
		{Name: "USD", Type: models.AssetTypeFiat, IsConvertible: true, IsActive: true},
		{Name: "EUR", Type: models.AssetTypeFiat, IsConvertible: true, IsActive: true},
		{Name: "UAH", Type: models.AssetTypeFiat, IsConvertible: true, IsActive: true},
		{Name: "GBP", Type: models.AssetTypeFiat, IsConvertible: true, IsActive: true},
		{Name: "PLN", Type: models.AssetTypeFiat, IsConvertible: true, IsActive: true},
		// crypto
		{Name: "BTC", Type: models.AssetTypeCrypto, IsActive: true},
		{Name: "ETH", Type: models.AssetTypeCrypto, IsActive: true},
		{Name: "USDT", Type: models.AssetTypeCrypto, IsActive: true},
		{Name: "USDC", Type: models.AssetTypeCrypto, IsActive: true},
		{Name: "XMR", Type: models.AssetTypeCrypto, IsActive: true},
	}
	return db.Create(&assets).Error
}
