package models

import (
	"ptop/internal/utils"

	"gorm.io/gorm"
)

type Country struct {
	ID   string `gorm:"primaryKey;size:21" json:"id"`
	Name string `gorm:"type:varchar(255);unique;not null" json:"name"`
}

func (c *Country) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID, err = utils.GenerateNanoID()
	}
	return
}
