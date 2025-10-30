package models

import (
	"gorm.io/gorm"
)

type Tag struct {
	gorm.Model

	Name  string  `gorm:"size:256"`
	Color *string `gorm:"size:6"` // RGB

	// https://gorm.io/docs/has_many.html
	UserID uint
}

func (Tag) TableName() string {
	return "Tags"
}
