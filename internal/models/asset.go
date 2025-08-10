package models

import (
	"ptop/internal/utils"

	"gorm.io/gorm"
)

const (
	AssetTypeFiat   = "fiat"
	AssetTypeCrypto = "crypto"
)

type Asset struct {
	ID            string `gorm:"primaryKey;size:21" json:"id"`
	Name          string `gorm:"type:varchar(255);unique;not null" json:"name"`
	Description   string `gorm:"type:varchar(1500);" json:"description"`
	Type          string `gorm:"type:varchar(10);not null" json:"type"`
	IsActive      bool   `gorm:"not null;default:false" json:"isActive"`
	IsConvertible bool   `gorm:"not null;default:false" json:"isConvertible"`
	Xpub          string `gorm:"type:varchar(255)" json:"-"`
}

func (a *Asset) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == "" {
		a.ID, err = utils.GenerateNanoID()
	}
	return
}
