package models

import (
	"gorm.io/gorm"
	"ptop/internal/utils"
)

type Country struct {
	ID   string `gorm:"primaryKey;size:21"`
	Name string `gorm:"type:varchar(255);unique;not null"`
}

func (c *Country) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID, err = utils.GenerateNanoID()
	}
	return
}
