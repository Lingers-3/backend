package models

import (
	"time"

	"gorm.io/gorm"
)

type ProjectTemplate struct {
	gorm.Model

	UserID uint `gorm:"not null"`
	User   User `gorm:"constraint:OnDelete:CASCADE;"`

	Name        string `gorm:"uniqueIndex:idx_user_template_name;not null"`
	Description *string

	PlannedWorkTime *time.Duration
	PlannedIncome   *float32

	UsageCount int64

	RequiredResources []TemplateResourceSpecification `gorm:"foreignKey:TemplateID"`
}

func (ProjectTemplate) TableName() string {
	return "ProjectTemplates"
}
