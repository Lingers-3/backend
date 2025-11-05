package models

import "gorm.io/gorm"

type Tag struct {
	gorm.Model

	// QUESTION(lili-ia): Do we need unique constraint user + tag's name?
	Name  string  `gorm:"size:256;not null;uniqueIndex:idx_user_name"`
	Color *string `gorm:"size:6"` // RGB

	UserID uint `gorm:"not null;uniqueIndex:idx_user_name"`

	Items     []Item     `gorm:"many2many:item_to_tags;"`
	ItemTypes []ItemType `gorm:"many2many:item_type_to_tags;"`
}

func (Tag) TableName() string {
	return "Tags"
}

type ItemToTag struct {
	ID     uint `gorm:"primaryKey"`
	ItemID uint `gorm:"index;not null"`
	TagID  uint `gorm:"index;not null"`
}

type ItemTypeToTag struct {
	ID         uint `gorm:"primaryKey"`
	ItemTypeID uint `gorm:"index;not null"`
	TagID      uint `gorm:"index;not null"`
}
