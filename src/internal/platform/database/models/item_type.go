package models

import (
	"gorm.io/gorm"
)

type ItemType struct {
	gorm.Model

	Name                          string  `gorm:"size:256"`
	Description                   *string `gorm:"size:512"`
	BaseMeasurementUnit           string  `gorm:"size:256"`
	DefaultDisplayMeasurementUnit string  `gorm:"size:256"`
	// HACK(noatu): using * for write as it is never null on read
	DefaultQuantity   *float32 `gorm:"default:1"`
	ShortageThreshold *float32

	// NOTE(noatu): GORM has two types of one-to-one relations:
	// 1. Belongs To: https://gorm.io/docs/belongs_to.html
	//    Store Picture, and PictureID will reference Picture.ID
	// 2. Has One: https://gorm.io/docs/has_one.html
	//    Store Picture, and Picture.ItemTypeID will reference ID
	// "Has One" would make Picture ItemType specific, so using "Belongs To"
	// HACK(noatu): using `*` as Picture is optional (not sure if it will work)
	PictureID *uint
	Picture   *Picture `gorm:"constraint:OnDelete:CASCADE;"`

	// https://gorm.io/docs/has_many.html
	UserID uint
	Items  []Item `gorm:"constraint:OnDelete:CASCADE;"`

	// https://gorm.io/docs/many_to_many.html
	Tags []Tag `gorm:"many2many:item_type_tags;"`
}

func (ItemType) TableName() string {
	return "ItemTypes"
}
