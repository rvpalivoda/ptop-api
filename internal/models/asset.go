package models

import (
	"gorm.io/gorm"
	"ptop/internal/utils"
)

const (
	AssetTypeFiat   = "fiat"
	AssetTypeCrypto = "crypto"
)

type Asset struct {
	ID            string `gorm:"primaryKey;size:21"`
	Name          string `gorm:"type:varchar(255);unique;not null"`
	Type          string `gorm:"type:varchar(10);not null"`
	IsActive      bool   `gorm:"not null;default:false"`
	IsConvertible bool   `gorm:"not null;default:false"`
	Xpub          string `gorm:"type:varchar(255)" json:"-"`
}

func (a *Asset) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == "" {
		a.ID, err = utils.GenerateNanoID()
	}
	return
}
