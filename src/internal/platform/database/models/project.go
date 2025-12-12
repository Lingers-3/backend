package models

import (
	"time"

	"gorm.io/gorm"
)

type ProjectState string

const (
	ProjectStatePlanning  ProjectState = "Planning"
	ProjectStateActive    ProjectState = "Active"
	ProjectStateCompleted ProjectState = "Completed"
	ProjectStateCanceled  ProjectState = "Canceled"
)

func (s ProjectState) IsValid() bool {
	switch s {
	case ProjectStatePlanning, ProjectStateActive, ProjectStateCompleted, ProjectStateCanceled:
		return true
	}
	return false
}

type Project struct {
	gorm.Model

	Name        string  `gorm:"size:256;not null"`
	Description *string `gorm:"size:512"`

	State ProjectState `gorm:"size:50;not null;check:state IN ('Planning', 'Active', 'Completed', 'Canceled')"`

	PlannedDeadline *time.Time
	PlannedIncome   *float32
	PlannedWorkTime *time.Duration

	ActualDeadline *time.Time
	ActualIncome   *float32
	ActualWorkTime *time.Duration

	StartedAt  *time.Time
	FinishedAt *time.Time

	UserID uint `gorm:"not null"`
	User   User `gorm:"constraint:OnDelete:CASCADE;"`

	ResourceSpecifications []ResourceSpecification `gorm:"constraint:OnDelete:CASCADE;"`
	ResourceReservations   []ResourceReservation   `gorm:"constraint:OnDelete:CASCADE;"`
}

func (Project) TableName() string {
	return "Projects"
}
