package models

import (
	"time"

	"ptop/internal/utils"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Client struct {
	ID           string          `gorm:"primaryKey;size:21" json:"id"`
	Username     string          `gorm:"type:varchar(255);not null;unique" json:"username"`
	PinCode      *string         `gorm:"type:varchar(255)" json:"-"`
	TwoFAEnabled bool            `gorm:"not null;default:false" json:"twoFAEnabled"`
	TOTPSecret   *string         `gorm:"type:varchar(255)" json:"-"`
	Bip39        datatypes.JSON  `gorm:"type:json" json:"bip39" swaggertype:"object" json:"-"`
	Password     *string         `gorm:"type:varchar(255)" json:"-"`
	RegistredAt  time.Time       `gorm:"autoCreateTime"`
	Rating       decimal.Decimal `gorm:"type:decimal(3,2);not null;default:0" json:"rating"`
	OrdersCount  int             `gorm:"not null;default:0" json:"ordersCount"`
}

func (c *Client) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID, err = utils.GenerateNanoID()
	}
	return
}
