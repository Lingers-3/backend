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
	var user models.User

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		log.Printf("ERROR: Failed to begin transaction: %v", tx.Error)
		return 0, ErrDatabaseError
	}

	err := tx.Unscoped().Where("auth0_id = ?", auth0ID).First(&user).Error
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ErrUserNotFound
		}
		log.Printf("ERROR: Failed to find user with auth0ID %s: %v", auth0ID, err)
		return 0, ErrDatabaseError
	}

	var itemTypeIDs []uint
	err = tx.Model(&models.ItemType{}).
		Unscoped().
		Where("user_id = ?", user.ID).
		Pluck("id", &itemTypeIDs).Error
	if err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to pluck ItemType IDs for user %d: %v", user.ID, err)
		return 0, ErrDatabaseError
	}

	if len(itemTypeIDs) > 0 {
		deleteItemsResult := tx.Unscoped().
			Where("item_type_id IN (?)", itemTypeIDs).
			Delete(&models.Item{})

		if deleteItemsResult.Error != nil {
			tx.Rollback()
			log.Printf("ERROR: Failed to delete related Items for user %d: %v", user.ID, deleteItemsResult.Error)
			return 0, ErrDatabaseError
		}
	}

	deleteUserResult := tx.Unscoped().Delete(&user)

	if deleteUserResult.Error != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to delete user with ID %d: %v", user.ID, deleteUserResult.Error)
		return 0, ErrDatabaseError
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("ERROR: Failed to commit transaction: %v", err)
		return 0, ErrDatabaseError
	}

	return deleteUserResult.RowsAffected, nil
}
