package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"ptop/internal/utils"
)

type TransactionOutStatus string

const (
	TransactionOutStatusPending    TransactionOutStatus = "pending"
	TransactionOutStatusProcessing TransactionOutStatus = "processing"
	TransactionOutStatusConfirmed  TransactionOutStatus = "confirmed"
	TransactionOutStatusFailed     TransactionOutStatus = "failed"
	TransactionOutStatusCancelled  TransactionOutStatus = "cancelled"
)

type TransactionOut struct {
	ID          string               `gorm:"primaryKey;size:21"`
	ClientID    string               `gorm:"size:21;not null"`
	Client      Client               `gorm:"foreignKey:ClientID" json:"-"`
	AssetID     string               `gorm:"size:21;not null"`
	Asset       Asset                `gorm:"foreignKey:AssetID" json:"-"`
	Amount      decimal.Decimal      `gorm:"type:decimal(32,8);not null"`
	FromAddress string               `gorm:"type:varchar(255)"`
	ToAddress   string               `gorm:"type:varchar(255)"`
	Status      TransactionOutStatus `gorm:"type:varchar(20);not null"`
	Data        datatypes.JSON       `gorm:"type:json"`
	CreatedAt   time.Time            `gorm:"autoCreateTime"`
	UpdatedAt   time.Time            `gorm:"autoUpdateTime"`
}

func (t *TransactionOut) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == "" {
		t.ID, err = utils.GenerateNanoID()
	}
	return
}
