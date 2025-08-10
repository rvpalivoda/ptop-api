package models

import (
	"time"

	"ptop/internal/utils"

	"gorm.io/gorm"
)

type Wallet struct {
	ID              string     `gorm:"primaryKey;size:21" json:"id"`
	ClientID        string     `gorm:"size:21;not null;uniqueIndex:idx_wallet_client_asset_active" json:"clientID"`
	Client          Client     `gorm:"foreignKey:ClientID" json:"-"`
	AssetID         string     `gorm:"size:21;not null;uniqueIndex:idx_wallet_client_asset_active" json:"assetID"`
	Asset           Asset      `gorm:"foreignKey:AssetID" json:"-"`
	Value           string     `gorm:"type:varchar(255);not null" json:"value"`
	DerivationIndex uint32     `gorm:"not null" json:"index"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	IsEnabled       bool       `gorm:"not null;default:true;uniqueIndex:idx_wallet_client_asset_active" json:"-"`
	EnabledAt       time.Time  `gorm:"autoCreateTime" json:"enabledAt"`
	DisabledAt      *time.Time `json:"-"`
}

func (w *Wallet) BeforeCreate(tx *gorm.DB) (err error) {
	if w.ID == "" {
		w.ID, err = utils.GenerateNanoID()
	}
	return
}
