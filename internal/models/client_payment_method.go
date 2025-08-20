package models

import (
	"gorm.io/gorm"
	"ptop/internal/utils"
)

type ClientPaymentMethod struct {
	ID                  string        `gorm:"primaryKey;size:21" json:"id"`
	ClientID            string        `gorm:"size:21;not null;uniqueIndex:idx_client_name" json:"clientID"`
	CountryID           string        `gorm:"size:21;not null" json:"countryID"`
	PaymentMethodID     string        `gorm:"size:21;not null" json:"paymentMethodID"`
	City                string        `gorm:"type:text" json:"city"`
	PostCode            string        `gorm:"type:text" json:"postCode"`
	DetailedInformation string        `gorm:"type:text" json:"detailedInformation"`
	Name                string        `gorm:"type:varchar(255);not null;uniqueIndex:idx_client_name" json:"name"`
	Country             Country       `gorm:"foreignKey:CountryID" json:"country"`
	PaymentMethod       PaymentMethod `gorm:"foreignKey:PaymentMethodID" json:"paymentMethod"`
}

func (cpm *ClientPaymentMethod) BeforeCreate(tx *gorm.DB) (err error) {
	if cpm.ID == "" {
		cpm.ID, err = utils.GenerateNanoID()
	}
	return
}
