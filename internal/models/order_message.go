package models

import (
	"errors"
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
	ChatID    string      `gorm:"size:21;not null;index:idx_chat_created,priority:1"`
	Chat      OrderChat   `gorm:"foreignKey:ChatID" json:"-"`
	ClientID  string      `gorm:"size:21;not null;index"`
	Client    Client      `gorm:"foreignKey:ClientID" json:"-"`
	Type      MessageType `gorm:"type:varchar(10);not null"`
	Content   string      `gorm:"type:text;not null"`
	FileURL   *string     `gorm:"type:text"`
	FileSize  *int64
	FileType  *string    `gorm:"type:varchar(100)"`
	ReadAt    *time.Time `gorm:"index"`
	CreatedAt time.Time  `gorm:"autoCreateTime;index:idx_chat_created,priority:2"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}

func (m *OrderMessage) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID, err = utils.GenerateNanoID()
		if err != nil {
			return
		}
	}
	if m.Type == MessageTypeFile && (m.FileURL == nil || *m.FileURL == "") {
		return errors.New("file url required")
	}
	return
}
