package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"ptop/internal/utils"
)

type OrderStatus string

const (
	OrderStatusWaitPayment OrderStatus = "WAIT_PAYMENT"
	OrderStatusPaid        OrderStatus = "PAID"
	OrderStatusReleased    OrderStatus = "RELEASED"
	OrderStatusCancelled   OrderStatus = "CANCELLED"
	OrderStatusDispute     OrderStatus = "DISPUTE"
)

type Order struct {
	ID                    string              `gorm:"primaryKey;size:21"`
	OfferID               string              `gorm:"size:21;not null"`
	Offer                 Offer               `gorm:"foreignKey:OfferID" json:"-"`
	BuyerID               string              `gorm:"size:21;not null"`
	Buyer                 Client              `gorm:"foreignKey:BuyerID" json:"-"`
	SellerID              string              `gorm:"size:21;not null"`
	Seller                Client              `gorm:"foreignKey:SellerID" json:"-"`
	AuthorID              string              `gorm:"size:21;not null" json:"authorID"`
	Author                Client              `gorm:"foreignKey:AuthorID" json:"-"`
	OfferOwnerID          string              `gorm:"size:21;not null" json:"offerOwnerID"`
	OfferOwner            Client              `gorm:"foreignKey:OfferOwnerID" json:"-"`
	FromAssetID           string              `gorm:"size:21;not null"`
	FromAsset             Asset               `gorm:"foreignKey:FromAssetID" json:"-"`
	ToAssetID             string              `gorm:"size:21;not null"`
	ToAsset               Asset               `gorm:"foreignKey:ToAssetID" json:"-"`
	Amount                decimal.Decimal     `gorm:"type:decimal(32,8);not null"`
	Price                 decimal.Decimal     `gorm:"type:decimal(32,8);not null"`
	ClientPaymentMethodID string              `gorm:"size:21"`
	ClientPaymentMethod   ClientPaymentMethod `gorm:"foreignKey:ClientPaymentMethodID" json:"-"`
	Status                OrderStatus         `gorm:"type:varchar(20);not null"`
	IsEscrow              bool                `gorm:"not null;default:false"`
	ExpiresAt             time.Time           `gorm:"not null"`
	ReleasedAt            *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == "" {
		o.ID, err = utils.GenerateNanoID()
	}
	return
}
