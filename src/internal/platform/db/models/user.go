package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Auth0ID                       string `gorm:"unique;size:256"`
	Email                         string `gorm:"unique;size:256"`
	NotifyOnExpiration            bool   `gorm:"default:true"`
	NotifyOnDeadline              bool   `gorm:"default:true"`
	NotifyOnThreshold             bool   `gorm:"default:true"`
	NotifyOnInsufficientResources bool   `gorm:"default:true"`

	// https://gorm.io/docs/has_many.html
	ItemTypes []ItemType `gorm:"constraint:OnDelete:CASCADE;"`
}

// TableName method will override the default table name used by GORM
// GORM can determine the table name automagically, but in custom SQL queries this method could be usefull
func (User) TableName() string {
	return "Users" // NOTE(noatu): consisent case with "ItemType"
}
