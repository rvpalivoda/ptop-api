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
	if db.Dialector.Name() == "sqlite" {
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
	if db.Dialector.Name() == "sqlite" {
		return "text"
	}
	return "enum('low','medium','high')"
}

type PaymentMethod struct {
	ID                    string   `gorm:"primaryKey;size:21"`
	Name                  string   `gorm:"type:varchar(255);unique;not null"`
	MethodGroup           string   `gorm:"type:varchar(100)"`
	Provider              string   `gorm:"type:varchar(100)"`
	TypicalFiatCCY        string   `gorm:"type:varchar(10)"`
	Regions               []string `gorm:"type:json;serializer:json"`
	IsRealtime            bool
	IsReversible          bool
	SettlementMinutes     uint
	ChargebackWindowHours uint
	FeeSide               FeeSideType      `gorm:"type:enum('sender','receiver','shared')"`
	KycLevelHint          KycLevelHintType `gorm:"type:enum('low','medium','high')"`
}

func (p *PaymentMethod) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID, err = utils.GenerateNanoID()
	}
	return
}
