package models

import (
	"time"

	"ptop/internal/utils"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type TransactionInStatus string

const (
	TransactionInStatusPending    TransactionInStatus = "pending"
	TransactionInStatusProcessing TransactionInStatus = "processing"
	TransactionInStatusConfirmed  TransactionInStatus = "confirmed"
	TransactionInStatusFailed     TransactionInStatus = "failed"
)

type TransactionIn struct {
	ID        string              `gorm:"primaryKey;size:21" json:"id"`
	ClientID  string              `gorm:"size:21;not null"`
	Client    Client              `gorm:"foreignKey:ClientID" json:"-"`
	WalletID  string              `gorm:"size:21;not null"`
	Wallet    Wallet              `gorm:"foreignKey:WalletID" json:"-"`
	AssetID   string              `gorm:"size:21;not null"`
	Asset     Asset               `gorm:"foreignKey:AssetID" json:"-"`
	AssetName string              `gorm:"->;column:asset_name" json:"assetName"`
	Amount    decimal.Decimal     `gorm:"type:decimal(32,8);not null" json:"amount"`
	Status    TransactionInStatus `gorm:"type:varchar(20);not null" json:"status"`
	Data      datatypes.JSON      `gorm:"type:json" swaggertype:"object"`
	CreatedAt time.Time           `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time           `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (t *TransactionIn) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == "" {
		t.ID, err = utils.GenerateNanoID()
	}
	return
}
