package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Auth0ID                       string `gorm:"unique"`
	Email                         string `gorm:"unique"`
	NotifyOnExpiration            bool   `gorm:"default:true"`
	NotifyOnDeadline              bool   `gorm:"default:true"`
	NotifyOnThreshold             bool   `gorm:"default:true"`
	NotifyOnInsufficientResources bool   `gorm:"default:true"`
}

// TableName method will override the default table name used by GORM
// GORM can determine the table name automagically, but in custom SQL queries this method could be usefull
func (User) TableName() string {
	return "users"
}
