package models

import (
	"gorm.io/gorm"
)

type ResourceReservation struct {
	gorm.Model

	ReservedQuantity float32
	UsedQuantity     float32

	ProjectID uint `gorm:"not null"`
	Project   Project

	// https://gorm.io/docs/has_many.html
	ResourceSpecificationID uint `gorm:"not null"`

	// https://gorm.io/docs/belongs_to.html
	ResourceSpecification ResourceSpecification `gorm:"constraint:OnDelete:RESTRICT;"`

	ItemID uint `gorm:"not null"`
	Item   Item `gorm:"constraint:OnDelete:RESTRICT;"`
}

func (ResourceReservation) TableName() string {
	return "ResourceReservations"
}
