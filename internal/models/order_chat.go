package models

import (
	"time"

	"gorm.io/gorm"
	"ptop/internal/utils"
)

type OrderChat struct {
	ID        string    `gorm:"primaryKey;size:21"`
	OrderID   string    `gorm:"size:21;not null;uniqueIndex"`
	Order     Order     `gorm:"foreignKey:OrderID" json:"-"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (c *OrderChat) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID, err = utils.GenerateNanoID()
	}
	return
}
