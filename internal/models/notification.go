package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"ptop/internal/utils"
)

// Notification представляет уведомление
// swagger:model
type Notification struct {
	ID        string          `gorm:"primaryKey;size:21" json:"id"`
	ClientID  string          `gorm:"size:21;not null;index" json:"clientID"`
	Type      string          `gorm:"type:varchar(255);not null" json:"type"`
	Payload   json.RawMessage `gorm:"type:jsonb" json:"payload" swaggertype:"object"`
	SentAt    *time.Time      `gorm:"index" json:"sentAt"`
	ReadAt    *time.Time      `gorm:"index" json:"readAt"`
	CreatedAt time.Time       `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) (err error) {
	if n.ID == "" {
		n.ID, err = utils.GenerateNanoID()
	}
	return
}
