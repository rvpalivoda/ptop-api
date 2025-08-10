package models

import (
	"time"

	"ptop/internal/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Balance struct {
	ID           string          `gorm:"primaryKey;size:21" json:"id"`
	ClientID     string          `gorm:"size:21;not null" json:"clientID"`
	Client       Client          `gorm:"foreignKey:ClientID" json:"-"`
	AssetID      string          `gorm:"size:21;not null" json:"assetID"`
	Asset        Asset           `gorm:"foreignKey:AssetID" json:"-"`
	Amount       decimal.Decimal `gorm:"type:decimal(32,8);not null" json:"amount"`
	AmountEscrow decimal.Decimal `gorm:"type:decimal(32,8);not null" json:"amountEscrow"`
	CreatedAt    time.Time       `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (b *Balance) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == "" {
		b.ID, err = utils.GenerateNanoID()
	}
	return
}
