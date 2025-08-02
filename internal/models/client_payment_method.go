package models

import (
	"gorm.io/gorm"
	"ptop/internal/utils"
)

type ClientPaymentMethod struct {
	ID              string `gorm:"primaryKey;size:21"`
	ClientID        string `gorm:"size:21;not null;uniqueIndex:idx_client_name"`
	CountryID       string `gorm:"size:21;not null"`
	PaymentMethodID string `gorm:"size:21;not null"`
	City            string `gorm:"type:text"`
	PostCode        string `gorm:"type:text"`
	Name            string `gorm:"type:varchar(255);not null;uniqueIndex:idx_client_name"`
}

func (cpm *ClientPaymentMethod) BeforeCreate(tx *gorm.DB) (err error) {
	if cpm.ID == "" {
		cpm.ID, err = utils.GenerateNanoID()
	}
	return
}
