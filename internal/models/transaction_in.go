package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"ptop/internal/utils"
)

type TransactionInStatus string

const (
	TransactionInStatusPending    TransactionInStatus = "pending"
	TransactionInStatusProcessing TransactionInStatus = "processing"
	TransactionInStatusConfirmed  TransactionInStatus = "confirmed"
	TransactionInStatusFailed     TransactionInStatus = "failed"
)

type TransactionIn struct {
	ID        string              `gorm:"primaryKey;size:21"`
	ClientID  string              `gorm:"size:21;not null"`
	Client    Client              `gorm:"foreignKey:ClientID" json:"-"`
	WalletID  string              `gorm:"size:21;not null"`
	Wallet    Wallet              `gorm:"foreignKey:WalletID" json:"-"`
        AssetID   string              `gorm:"size:21;not null"`
        Asset     Asset               `gorm:"foreignKey:AssetID" json:"-"`
        AssetName string              `gorm:"->;column:asset_name" json:"assetName"`
	Amount    decimal.Decimal     `gorm:"type:decimal(32,8);not null"`
	Status    TransactionInStatus `gorm:"type:varchar(20);not null"`
        Data      datatypes.JSON      `gorm:"type:json" swaggertype:"object"`
	CreatedAt time.Time           `gorm:"autoCreateTime"`
	UpdatedAt time.Time           `gorm:"autoUpdateTime"`
}

func (t *TransactionIn) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == "" {
		t.ID, err = utils.GenerateNanoID()
	}
	return
}
