package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"ptop/internal/utils"
)

type Balance struct {
	ID           string          `gorm:"primaryKey;size:21"`
	ClientID     string          `gorm:"size:21;not null"`
	Client       Client          `gorm:"foreignKey:ClientID" json:"-"`
	AssetID      string          `gorm:"size:21;not null"`
	Asset        Asset           `gorm:"foreignKey:AssetID" json:"-"`
	Amount       decimal.Decimal `gorm:"type:decimal(32,8);not null"`
	AmountEscrow decimal.Decimal `gorm:"type:decimal(32,8);not null"`
	CreatedAt    time.Time       `gorm:"autoCreateTime"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime"`
}

func (b *Balance) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == "" {
		b.ID, err = utils.GenerateNanoID()
	}
	return
}
