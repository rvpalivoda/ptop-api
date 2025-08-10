package models

import (
	"time"

	"ptop/internal/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Escrow struct {
	ID        string          `gorm:"primaryKey;size:21" json:"id"`
	ClientID  string          `gorm:"size:21;not null" json:"clientID"`
	Client    Client          `gorm:"foreignKey:ClientID" json:"-"`
	AssetID   string          `gorm:"size:21;not null" json:"assetID"`
	Asset     Asset           `gorm:"foreignKey:AssetID" json:"-"`
	Amount    decimal.Decimal `gorm:"type:decimal(32,8);not null" json:"amount"`
	OfferID   *string         `gorm:"size:21" json:"offerID"`
	Offer     Offer           `gorm:"foreignKey:OfferID" json:"-"`
	OrderID   *string         `gorm:"size:21" json:"orderID"`
	Order     Order           `gorm:"foreignKey:OrderID" json:"-"`
	CreatedAt time.Time       `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (e *Escrow) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == "" {
		e.ID, err = utils.GenerateNanoID()
	}
	return
}
