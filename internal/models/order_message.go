package models

import (
	"time"

	"gorm.io/gorm"
	"ptop/internal/utils"
)

type MessageType string

const (
	MessageTypeText   MessageType = "TEXT"
	MessageTypeSystem MessageType = "SYSTEM"
	MessageTypeFile   MessageType = "FILE"
)

type OrderMessage struct {
	ID        string      `gorm:"primaryKey;size:21"`
	ChatID    string      `gorm:"size:21;not null;index"`
	Chat      OrderChat   `gorm:"foreignKey:ChatID" json:"-"`
	ClientID  string      `gorm:"size:21;not null;index"`
	Client    Client      `gorm:"foreignKey:ClientID" json:"-"`
	Type      MessageType `gorm:"type:varchar(10);not null"`
	Content   string      `gorm:"type:text;not null"`
	CreatedAt time.Time   `gorm:"autoCreateTime;index"`
	UpdatedAt time.Time   `gorm:"autoUpdateTime"`
}

func (m *OrderMessage) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID, err = utils.GenerateNanoID()
	}
	return
}
