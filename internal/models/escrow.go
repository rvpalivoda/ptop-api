package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"ptop/internal/utils"
)

type Escrow struct {
	ID        string          `gorm:"primaryKey;size:21"`
	ClientID  string          `gorm:"size:21;not null"`
	Client    Client          `gorm:"foreignKey:ClientID" json:"-"`
	AssetID   string          `gorm:"size:21;not null"`
	Asset     Asset           `gorm:"foreignKey:AssetID" json:"-"`
	Amount    decimal.Decimal `gorm:"type:decimal(32,8);not null"`
	OfferID   *string         `gorm:"size:21"`
	Offer     Offer           `gorm:"foreignKey:OfferID" json:"-"`
	OrderID   *string         `gorm:"size:21"`
	Order     Order           `gorm:"foreignKey:OrderID" json:"-"`
	CreatedAt time.Time       `gorm:"autoCreateTime"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime"`
}

func (e *Escrow) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == "" {
		e.ID, err = utils.GenerateNanoID()
	}
	return
}
