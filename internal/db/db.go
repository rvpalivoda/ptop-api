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

	// Автомуграция ваших моделей
	if err := db.AutoMigrate(
		&models.Client{},
		&models.Token{},
		&models.Country{},
		&models.PaymentMethod{},
		&models.ClientPaymentMethod{},
		&models.Asset{},
		// &models.Product{}, и т.д.
	); err != nil {
		return nil, fmt.Errorf("auto migrate failed: %w", err)
	}

	return db, nil
}
