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
	ID                    string              `gorm:"primaryKey;size:21" json:"id"`
	OfferID               string              `gorm:"size:21;not null" json:"offerID"`
	Offer                 Offer               `gorm:"foreignKey:OfferID" json:"-"`
	BuyerID               string              `gorm:"size:21;not null" json:"buyerID"`
	Buyer                 Client              `gorm:"foreignKey:BuyerID" json:"-"`
	SellerID              string              `gorm:"size:21;not null" json:"sellerID"`
	Seller                Client              `gorm:"foreignKey:SellerID" json:"-"`
	AuthorID              string              `gorm:"size:21;not null" json:"authorID"`
	Author                Client              `gorm:"foreignKey:AuthorID" json:"-"`
	OfferOwnerID          string              `gorm:"size:21;not null" json:"offerOwnerID"`
	OfferOwner            Client              `gorm:"foreignKey:OfferOwnerID" json:"-"`
	FromAssetID           string              `gorm:"size:21;not null" json:"fromAssetID"`
	FromAsset             Asset               `gorm:"foreignKey:FromAssetID" json:"-"`
	ToAssetID             string              `gorm:"size:21;not null" json:"toAssetID"`
	ToAsset               Asset               `gorm:"foreignKey:ToAssetID" json:"-"`
	Amount                decimal.Decimal     `gorm:"type:decimal(32,8);not null" json:"amount"`
	Price                 decimal.Decimal     `gorm:"type:decimal(32,8);not null" json:"price"`
	ClientPaymentMethodID string              `gorm:"size:21" json:"clientPaymentMethodID"`
	ClientPaymentMethod   ClientPaymentMethod `gorm:"foreignKey:ClientPaymentMethodID" json:"-"`
	Status                OrderStatus         `gorm:"type:varchar(20);not null" json:"status"`
	IsEscrow              bool                `gorm:"not null;default:false" json:"isEscrow"`
	ExpiresAt             time.Time           `gorm:"not null" json:"expiresAt"`
	ReleasedAt            *time.Time          `json:"releasedAt"`
	CreatedAt             time.Time           `json:"createdAt"`
	UpdatedAt             time.Time           `json:"updatedAt"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == "" {
		o.ID, err = utils.GenerateNanoID()
	}
	return
}
