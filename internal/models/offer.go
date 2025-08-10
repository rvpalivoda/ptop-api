package models

import (
	"time"

	"ptop/internal/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Offer struct {
	ID                     string                `gorm:"primaryKey;size:21" json:"id"`
	MaxAmount              decimal.Decimal       `gorm:"type:decimal(32,8);not null" json:"maxAmount"`
	MinAmount              decimal.Decimal       `gorm:"type:decimal(32,8);not null" json:"minAmount"`
	Amount                 decimal.Decimal       `gorm:"type:decimal(32,8);not null" json:"amount"`
	Price                  decimal.Decimal       `gorm:"type:decimal(32,8);not null" json:"price"`
	FromAssetID            string                `gorm:"size:21;not null" json:"fromAssetID"`
	FromAsset              Asset                 `gorm:"foreignKey:FromAssetID" json:"-"`
	ToAssetID              string                `gorm:"size:21;not null" json:"toAssetID"`
	ToAsset                Asset                 `gorm:"foreignKey:ToAssetID" json:"-"`
	Conditions             string                `gorm:"type:text" json:"conditions"`
	OrderExpirationTimeout int                   `gorm:"not null;default:15" json:"orderExpirationTimeout"`
	TTL                    time.Time             `gorm:"not null" json:"TTL"`
	EnabledAt              *time.Time            `json:"enabledAt"`
	DisabledAt             *time.Time            `json:"disabledAt"`
	IsEnabled              bool                  `gorm:"not null;default:false" json:"isEnabled"`
	ClientID               string                `gorm:"size:21;not null" json:"clientID"`
	Client                 Client                `gorm:"foreignKey:ClientID" json:"-"`
	ClientPaymentMethods   []ClientPaymentMethod `gorm:"many2many:offer_client_payment_methods" json:"-"`
	CreatedAt              time.Time             `json:"createdAt"`
	UpdatedAt              time.Time             `json:"updatedAt"`
}

func (o *Offer) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == "" {
		o.ID, err = utils.GenerateNanoID()
	}
	return
}
