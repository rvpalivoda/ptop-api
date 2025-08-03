package db

import (
	"github.com/biter777/countries"
	"gorm.io/gorm"

	"ptop/internal/models"
)

var (
	euCountries = []string{
		countries.AT.String(),
		countries.BE.String(),
		countries.BG.String(),
		countries.HR.String(),
		countries.CY.String(),
		countries.CZ.String(),
		countries.DK.String(),
		countries.EE.String(),
		countries.FI.String(),
		countries.FR.String(),
		countries.DE.String(),
		countries.GR.String(),
		countries.HU.String(),
		countries.IE.String(),
		countries.IT.String(),
		countries.LV.String(),
		countries.LT.String(),
		countries.LU.String(),
		countries.MT.String(),
		countries.NL.String(),
		countries.PL.String(),
		countries.PT.String(),
		countries.RO.String(),
		countries.SK.String(),
		countries.SI.String(),
		countries.ES.String(),
		countries.SE.String(),
	}
	regionCountries = map[string][]string{
		"EU":  euCountries,
		"EEA": append(append([]string{}, euCountries...), countries.IS.String(), countries.LI.String(), countries.NO.String()),
		"GLOBAL": func() []string {
			list := make([]string, 0, countries.Total())
			for _, c := range countries.All() {
				list = append(list, c.String())
			}
			return list
		}(),
	}
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
			Name:              "Interac e-Transfer",
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
			Name:              "RTP (Network)",
			MethodGroup:       "bank_transfer",
			Provider:          "RTP",
			TypicalFiatCCY:    "USD",
			Regions:           []string{"US"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintHigh,
		},
		{
			Name:              "FedNow",
			MethodGroup:       "bank_transfer",
			Provider:          "Federal Reserve",
			TypicalFiatCCY:    "USD",
			Regions:           []string{"US"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintHigh,
		},
		{
			Name:              "NPP / Osko",
			MethodGroup:       "bank_transfer",
			Provider:          "NPP",
			TypicalFiatCCY:    "AUD",
			Regions:           []string{"AU"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintMedium,
		},
		{
			Name:              "PayNow (FAST)",
			MethodGroup:       "bank_transfer",
			Provider:          "FAST",
			TypicalFiatCCY:    "SGD",
			Regions:           []string{"SG"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintMedium,
		},
		{
			Name:              "PromptPay",
			MethodGroup:       "bank_transfer",
			Provider:          "PromptPay",
			TypicalFiatCCY:    "THB",
			Regions:           []string{"TH"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintMedium,
		},
		{
			Name:              "DuitNow",
			MethodGroup:       "bank_transfer",
			Provider:          "DuitNow",
			TypicalFiatCCY:    "MYR",
			Regions:           []string{"MY"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintMedium,
		},
		{
			Name:              "FPS (Hong Kong)",
			MethodGroup:       "bank_transfer",
			Provider:          "FPS",
			TypicalFiatCCY:    "HKD",
			Regions:           []string{"HK"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintMedium,
		},
		{
			Name:              "SBP (Система быстрых платежей)",
			MethodGroup:       "bank_transfer",
			Provider:          "SBP",
			TypicalFiatCCY:    "RUB",
			Regions:           []string{"RU"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintMedium,
		},
		{
			Name:              "Orange Money",
			MethodGroup:       "mobile_money",
			Provider:          "Orange",
			TypicalFiatCCY:    "XOF/XAF",
			Regions:           []string{"SN", "CI", "ML", "BF", "BJ"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideReceiver,
			KycLevelHint:      models.KycLevelHintLow,
		},
		{
			Name:              "MTN MoMo",
			MethodGroup:       "mobile_money",
			Provider:          "MTN",
			TypicalFiatCCY:    "GHS/UGX/ZAR",
			Regions:           []string{"GH", "UG", "ZA"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideReceiver,
			KycLevelHint:      models.KycLevelHintLow,
		},
		{
			Name:              "TIPS (pan-EU rail)",
			MethodGroup:       "instant_pay_net",
			Provider:          "TIPS",
			TypicalFiatCCY:    "EUR",
			Regions:           []string{"EEA"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintMedium,
		},
		{
			Name:              "Alipay",
			MethodGroup:       "e_wallet",
			Provider:          "Ant Financial",
			TypicalFiatCCY:    "CNY",
			Regions:           []string{"CN"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintLow,
		},
		{
			Name:              "WeChat Pay",
			MethodGroup:       "e_wallet",
			Provider:          "Tencent",
			TypicalFiatCCY:    "CNY",
			Regions:           []string{"CN"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintLow,
		},
		{
			Name:              "Cash App",
			MethodGroup:       "e_wallet",
			Provider:          "Block",
			TypicalFiatCCY:    "USD",
			Regions:           []string{"US"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintLow,
		},
		{
			Name:              "Venmo",
			MethodGroup:       "e_wallet",
			Provider:          "PayPal",
			TypicalFiatCCY:    "USD",
			Regions:           []string{"US"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintLow,
		},
		{
			Name:              "Ria Money Transfer",
			MethodGroup:       "money_xfer_svc",
			Provider:          "Ria",
			TypicalFiatCCY:    "multi-currency",
			Regions:           []string{"GLOBAL"},
			IsRealtime:        false,
			IsReversible:      false,
			SettlementMinutes: 180,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintMedium,
		},
		{
			Name:              "WorldRemit cash pickup",
			MethodGroup:       "money_xfer_svc",
			Provider:          "WorldRemit",
			TypicalFiatCCY:    "multi-currency",
			Regions:           []string{"GLOBAL"},
			IsRealtime:        false,
			IsReversible:      false,
			SettlementMinutes: 180,
			FeeSide:           models.FeeSideSender,
			KycLevelHint:      models.KycLevelHintMedium,
		},
		{
			Name:              "Flexepin code",
			MethodGroup:       "prepaid_voucher",
			Provider:          "Flexepin",
			TypicalFiatCCY:    "CAD/AUD/EUR",
			Regions:           []string{"CA", "AU", "EU"},
			IsRealtime:        true,
			IsReversible:      false,
			SettlementMinutes: 1,
			FeeSide:           models.FeeSideSender,
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

	for _, m := range methods {
		method := m
		if err := db.Create(&method).Error; err != nil {
			return err
		}
		for _, region := range method.Regions {
			name := countries.ByName(region).String()
			if name == "Unknown" {

				names, ok := regionCountries[region]
				if !ok {
					continue
				}
				for _, n := range names {
					var country models.Country
					if err := db.Where("name = ?", n).First(&country).Error; err != nil {
						return err
					}
					if err := db.Model(&method).Association("Countries").Append(&country); err != nil {
						return err
					}
				}

				continue
			}
			var country models.Country
			if err := db.Where("name = ?", name).First(&country).Error; err != nil {
				return err
			}
			if err := db.Model(&method).Association("Countries").Append(&country); err != nil {
				return err
			}
		}
	}
	return nil
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
