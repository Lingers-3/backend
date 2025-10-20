package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Picture struct {
	gorm.Model

	BlobID   uuid.UUID
	Filetype string `gorm:"size:8"`
}

func (Picture) TableName() string {
	return "Pictures"
}
