package models

import (
	"time"

	"gorm.io/gorm"
	"ptop/internal/utils"
)

type Token struct {
	ID        string    `gorm:"primaryKey;size:21"`
	ClientID  string    `gorm:"index;not null"`
	Token     string    `gorm:"type:varchar(255);not null;unique"`
	Type      string    `gorm:"type:varchar(10);not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (t *Token) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == "" {
		t.ID, err = utils.GenerateNanoID()
	}
	return
}
