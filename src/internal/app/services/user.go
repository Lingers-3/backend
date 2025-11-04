package services

import (
	"context"
	"errors"
	"log"

	"pocketeer/internal/platform/database/models"

	"gorm.io/gorm"
)

// QUESTION(noatu): service where T-T?

func GetUserIDByAuth0ID(ctx context.Context, db *gorm.DB, auth0ID string) (uint, error) {
	var model struct{ ID uint }
	err := db.WithContext(ctx).Model(&models.User{}).First(&model, "auth0_id = ?", auth0ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrUnauthenticated
		}
		log.Printf("ERROR: searching for user: %v", err)
		return 0, ErrDatabaseError
	}

	return model.ID, nil
}
