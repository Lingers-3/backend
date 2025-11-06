package models

import (
	"time"

	"gorm.io/gorm"
)

type Item struct {
	gorm.Model

	Description            *string `gorm:"size:512"`
	Quantity               float32
	ExpirationDate         *time.Time
	DisplayMeasurementUnit string `gorm:"size:256"`
	PurchasePrice          *float32

	// https://gorm.io/docs/has_many.html
	ItemTypeID uint

	// https://gorm.io/docs/many_to_many.html
	// NOTE(pencelheimer): CASCADE for hard delete
	Tags []Tag `gorm:"many2many:item_tags;constraint:OnDelete:CASCADE;"`
}

func (Item) TableName() string {
	return "Items"
}
