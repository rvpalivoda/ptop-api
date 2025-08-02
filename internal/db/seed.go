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
		{Name: "Bank"},
		{Name: "Cash"},
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
		{Name: "USD", Type: "fiat", IsConvertible: true},
		{Name: "BTC", Type: "crypto"},
	}
	return db.Create(&assets).Error
}
