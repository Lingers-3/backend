package models

import (
	"gorm.io/gorm"
)

type ResourceType string

const (
	ResourceTypeConsumable ResourceType = "Consumable"
	ResourceTypeInstrument ResourceType = "Instrument"
)

type ResourceSpecification struct {
	gorm.Model

	ResourceType ResourceType `gorm:"size:50;not null;check:resource_type IN ('Consumable', 'Instrument')"`

	PlannedQuantity float32

	ProjectID uint    `gorm:"not null"`
	Project   Project `gorm:"constraint:OnDelete:CASCADE;"`

	ItemTypeID uint     `gorm:"not null"`
	ItemType   ItemType `gorm:"constraint:OnDelete:RESTRICT;"`
}

func (ResourceSpecification) TableName() string {
	return "ResourceSpecifications"
}
