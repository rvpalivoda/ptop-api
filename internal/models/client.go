package models

import (
	"time"

	"ptop/internal/utils"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Client struct {
	ID           string         `gorm:"primaryKey;size:21" json:"id"`
	Username     string         `gorm:"type:varchar(255);not null;unique" json:"username"`
	PinCode      *string        `gorm:"type:varchar(255)" json:"pinCode"`
	TwoFAEnabled bool           `gorm:"not null;default:false" json:"twoFAEnabled"`
	TOTPSecret   *string        `gorm:"type:varchar(255)" json:"TOTPSecret"`
        Bip39        datatypes.JSON `gorm:"type:json" json:"bip39" swaggertype:"object"`
	Password     *string        `gorm:"type:varchar(255)"`
	RegistredAt  time.Time      `gorm:"autoCreateTime"`
}

func (c *Client) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID, err = utils.GenerateNanoID()
	}
	return
}
