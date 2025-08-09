package models

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"ptop/internal/utils"
)

type FeeSideType string

const (
	FeeSideSender   FeeSideType = "sender"
	FeeSideReceiver FeeSideType = "receiver"
	FeeSideShared   FeeSideType = "shared"
)

func (FeeSideType) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite", "postgres":
		return "text"
	}
	return "enum('sender','receiver','shared')"
}

type KycLevelHintType string

const (
	KycLevelHintLow    KycLevelHintType = "low"
	KycLevelHintMedium KycLevelHintType = "medium"
	KycLevelHintHigh   KycLevelHintType = "high"
)

func (KycLevelHintType) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "sqlite", "postgres":
		return "text"
	}
	return "enum('low','medium','high')"
}

type PaymentMethod struct {
	ID                    string    `gorm:"primaryKey;size:21" json:"id"`
	Name                  string    `gorm:"type:varchar(255);unique;not null" json:"name"`
	MethodGroup           string    `gorm:"type:varchar(100)" json:"methodGroup"`
	Provider              string    `gorm:"type:varchar(100)" json:"provider"`
	TypicalFiatCCY        string    `gorm:"type:varchar(32)" json:"typicalFiatCCY"`
	Regions               []string  `gorm:"type:json;serializer:json" json:"regions"`
	Countries             []Country `gorm:"many2many:payment_method_countries" json:"countries"`
	IsRealtime            bool      `json:"isRealtime"`
	IsReversible          bool      `json:"isReversible"`
	SettlementMinutes     uint      ``
	ChargebackWindowHours uint
	FeeSide               FeeSideType
	KycLevelHint          KycLevelHintType
}

func (p *PaymentMethod) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID, err = utils.GenerateNanoID()
	}
	return
}
