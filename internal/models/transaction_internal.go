package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"ptop/internal/utils"
)

type TransactionInternalStatus string

const (
	TransactionInternalStatusProcessing TransactionInternalStatus = "processing"
	TransactionInternalStatusConfirmed  TransactionInternalStatus = "confirmed"
	TransactionInternalStatusFailed     TransactionInternalStatus = "failed"
)

type TransactionInternal struct {
	ID           string                    `gorm:"primaryKey;size:21"`
	AssetID      string                    `gorm:"size:21;not null"`
	Asset        Asset                     `gorm:"foreignKey:AssetID" json:"-"`
	Amount       decimal.Decimal           `gorm:"type:decimal(32,8);not null"`
	OrderInfo    string                    `gorm:"type:text"`
	FromClientID string                    `gorm:"size:21"`
	FromClient   Client                    `gorm:"foreignKey:FromClientID" json:"-"`
	ToClientID   string                    `gorm:"size:21"`
	ToClient     Client                    `gorm:"foreignKey:ToClientID" json:"-"`
	Status       TransactionInternalStatus `gorm:"type:varchar(20);not null"`
	Data         datatypes.JSON            `gorm:"type:json"`
	CreatedAt    time.Time                 `gorm:"autoCreateTime"`
	UpdatedAt    time.Time                 `gorm:"autoUpdateTime"`
}

func (t *TransactionInternal) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == "" {
		t.ID, err = utils.GenerateNanoID()
	}
	return
}
