package models

import "gorm.io/gorm"

type User struct {
	gorm.Model        // поля ID, CreatedAt, UpdatedAt, DeletedAt
	Name       string `gorm:"type:TEXT;not null"`
	Email      string `gorm:"type:TEXT;uniqueIndex;not null"`
}
