package models

import "gorm.io/gorm"

type TemplateResourceSpecification struct {
	gorm.Model

	TemplateID uint            `gorm:"not null;uniqueIndex:idx_template_item_type"`
	Template   ProjectTemplate `gorm:"constraint:OnDelete:CASCADE;"`

	ItemTypeID      uint         `gorm:"not null;uniqueIndex:idx_template_item_type"`
	ItemType        ItemType     `gorm:"constraint:OnDelete:RESTRICT;"`
	ResourceType    ResourceType `gorm:"size:50;not null;check:resource_type IN ('Consumable', 'Instrument')"`
	PlannedQuantity float32      `gorm:"not null"`
}

func (TemplateResourceSpecification) TableName() string {
	return "TemplateResourceSpecifications"
}
