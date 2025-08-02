package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"ptop/internal/utils"
)

type Offer struct {
	ID                     string          `gorm:"primaryKey;size:21"`
	MaxAmount              decimal.Decimal `gorm:"type:decimal(32,8);not null"`
	MinAmount              decimal.Decimal `gorm:"type:decimal(32,8);not null"`
	Amount                 decimal.Decimal `gorm:"type:decimal(32,8);not null"`
	Price                  decimal.Decimal `gorm:"type:decimal(32,8);not null"`
	FromAssetID            string          `gorm:"size:21;not null"`
	FromAsset              Asset           `gorm:"foreignKey:FromAssetID" json:"-"`
	ToAssetID              string          `gorm:"size:21;not null"`
	ToAsset                Asset           `gorm:"foreignKey:ToAssetID" json:"-"`
	Conditions             string          `gorm:"type:text"`
	OrderExpirationTimeout int             `gorm:"not null;default:15"`
	TTL                    time.Time       `gorm:"not null"`
	EnabledAt              *time.Time
	DisabledAt             *time.Time
	IsEnabled              bool                  `gorm:"not null;default:false"`
	ClientID               string                `gorm:"size:21;not null"`
	Client                 Client                `gorm:"foreignKey:ClientID" json:"-"`
	PaymentMethods         []PaymentMethod       `gorm:"many2many:offer_payment_methods" json:"-"`
	ClientPaymentMethods   []ClientPaymentMethod `gorm:"many2many:offer_client_payment_methods" json:"-"`
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

func (o *Offer) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == "" {
		o.ID, err = utils.GenerateNanoID()
	}
	return
}
