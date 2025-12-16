package models

import "gorm.io/gorm"

type Picture struct {
	gorm.Model

	UserID           uint   `gorm:"not null;uniqueIndex:idx_user_hash"`         // ownership checks
	Hash             string `gorm:"size:64;uniqueIndex:idx_user_hash;not null"` // SHA256 hash
	OriginalFilename string `gorm:"size:256;not null"`
	MimeType         string `gorm:"size:100;not null"` // image/jpeg
	Size             int64  `gorm:"not null"`          // file size in bytes
}

func (Picture) TableName() string {
	return "Pictures"
}
