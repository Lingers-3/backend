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

	ItemID uint `gorm:"not null"`
	Item   Item `gorm:"constraint:OnDelete:RESTRICT;"`
}

func (ResourceReservation) TableName() string {
	return "ResourceReservations"
}
