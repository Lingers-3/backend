package models

import (
	"context"

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

// NOTE(noatu): this one will be very repetetive
func GetUserIDByAuth0ID(ctx context.Context, db *gorm.DB, auth0ID string) (uint, error) {
	var model struct{ ID uint }
	result := db.WithContext(ctx).Model(&User{}).First(&model, "auth0_id = ?", auth0ID)
	if result.Error != nil {
		return 0, result.Error
	}
	return model.ID, nil
}
