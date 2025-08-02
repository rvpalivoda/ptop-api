package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"ptop/internal/models"
)

func NewDB(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Postgres: %w", err)
	}

	if err := db.AutoMigrate(
		&models.Client{},
		&models.Token{},
		&models.Country{},
		&models.PaymentMethod{},
		&models.Order{},
		&models.ClientPaymentMethod{},
		&models.Asset{},
		&models.Offer{},
		&models.Wallet{},
		&models.TransactionIn{},
		&models.TransactionOut{},
		&models.TransactionInternal{},
		&models.Balance{},
		&models.Escrow{},
	// &models.Product{}, и т.д.
	); err != nil {
		return nil, fmt.Errorf("auto migrate failed: %w", err)
	}

	// Отдельная миграция для Order, чтобы избежать ссылки на несуществующие таблицы
	if err := db.AutoMigrate(&models.Order{}); err != nil {
		return nil, fmt.Errorf("auto migrate failed: %w", err)
	}

	return db, nil
}
