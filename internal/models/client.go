package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
	"ptop/internal/utils"
)

type Client struct {
        ID           string         `gorm:"primaryKey;size:21"`
        Username     string         `gorm:"type:varchar(255);not null;unique"`
        PinCode      *string        `gorm:"type:varchar(255)"`
        TwoFAEnabled bool           `gorm:"not null;default:false"`
        TOTPSecret   *string        `gorm:"type:varchar(255)"`
        Bip39        datatypes.JSON `gorm:"type:json"`
        Password     *string        `gorm:"type:varchar(255)"`
        RegistredAt  time.Time      `gorm:"autoCreateTime"`
}

func (c *Client) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID, err = utils.GenerateNanoID()
	}
	return
}
