package services

import (
	"context"
	"errors"
	"log"
	"pocketeer/internal/platform/database"

	"pocketeer/internal/platform/database/models"

	"gorm.io/gorm"
)

type UserService struct {
	db *database.DB
}

func NewUserService(db *database.DB) *UserService {
	return &UserService{db}
}

func (s *UserService) GetUserIDByAuth0ID(ctx context.Context, auth0ID string) (uint, error) {
	var model struct{ ID uint }
	err := s.db.WithContext(ctx).Model(&models.User{}).First(&model, "auth0_id = ?", auth0ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrUnauthenticated
		}
		log.Printf("ERROR: searching for user: %v", err)
		return 0, ErrDatabaseError
	}

	return model.ID, nil
}

func (s *UserService) EnsureActiveUser(ctx context.Context, auth0ID string, email string) (uint, error) {
	var user models.User

	result := s.db.WithContext(ctx).Unscoped().
		Where("auth0_id = ? OR email = ?", auth0ID, email).
		First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			newUser := models.User{
				Auth0ID: auth0ID,
				Email:   email,
			}

			if err := s.db.WithContext(ctx).Create(&newUser).Error; err != nil {
				log.Printf("ERROR: failed to create new user %s: %v", auth0ID, err)
				return 0, ErrDatabaseError
			}
			log.Printf("New user created: %s (%s)", auth0ID, email)

			return newUser.ID, nil

		} else {
			log.Printf("ERROR: DB error during user search for %s: %v", auth0ID, result.Error)
			return 0, ErrDatabaseError
		}
	}

	if user.DeletedAt.Valid {
		if err := s.db.WithContext(ctx).Model(&user).Update("deleted_at", nil).Error; err != nil {
			log.Printf("ERROR: failed to restore soft-deleted user %s: %v", auth0ID, err)
			return 0, ErrDatabaseError
		}
		log.Printf("Restored soft-deleted user: %s", auth0ID)
	}

	return user.ID, nil
}

func (s *UserService) DeleteUserByAuth0ID(ctx context.Context, auth0ID string) (int64, error) {
	result := s.db.WithContext(ctx).Unscoped().Where("auth0_id = ?", auth0ID).Delete(&models.User{})

	if result.Error != nil {
		log.Printf("ERROR: Failed to delete user with auth0ID %s: %v", auth0ID, result.Error)
		return 0, ErrDatabaseError
	}

	if result.RowsAffected == 0 {
		return 0, ErrUserNotFound
	}

	return result.RowsAffected, nil
}
