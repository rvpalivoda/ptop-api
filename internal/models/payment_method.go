package models

import (
	"gorm.io/gorm"
	"ptop/internal/utils"
)

type PaymentMethod struct {
	ID   string `gorm:"primaryKey;size:21"`
	Name string `gorm:"type:varchar(255);unique;not null"`
}

func (p *PaymentMethod) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID, err = utils.GenerateNanoID()
	}
	return
}
