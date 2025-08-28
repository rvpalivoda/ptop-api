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

// OrderMessage представляет сообщение в чате ордера
// swagger:model
type OrderMessage struct {
    ID        string      `gorm:"primaryKey;size:21" json:"id"`
    ChatID    string      `gorm:"size:21;not null;index:idx_chat_created,priority:1" json:"chatID"`
    Chat      OrderChat   `gorm:"foreignKey:ChatID" json:"-"`
    ClientID  string      `gorm:"size:21;not null;index" json:"clientID"`
    Client    Client      `gorm:"foreignKey:ClientID" json:"-"`
    // SenderName — имя (username) отправителя сообщения
    SenderName string     `gorm:"-" json:"senderName"`
    Type      MessageType `gorm:"type:varchar(10);not null" json:"type"`
    Content   string      `gorm:"type:text;not null" json:"content"`
    FileURL   *string     `gorm:"type:text" json:"fileURL,omitempty"`
    FileSize  *int64      `json:"fileSize,omitempty"`
    FileType  *string     `gorm:"type:varchar(100)" json:"fileType,omitempty"`
	ReadAt    *time.Time  `gorm:"index" json:"readAt,omitempty"`
	CreatedAt time.Time   `gorm:"autoCreateTime;index:idx_chat_created,priority:2" json:"createdAt"`
	UpdatedAt time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
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
