package models

import "gorm.io/gorm"

type Tag struct {
	gorm.Model

	Name  string  `gorm:"size:256;not null;uniqueIndex:idx_user_name"`
	Color *string `gorm:"size:6"` // RGB

	// https://gorm.io/docs/has_many.html
	UserID uint `gorm:"not null;uniqueIndex:idx_user_name"`
}

func (Tag) TableName() string {
	return "Tags"
}
