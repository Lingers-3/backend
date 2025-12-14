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

func (s *UserService) DeleteUserByAuth0ID(ctx context.Context, auth0ID string) error {
	var user models.User

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		log.Printf("ERROR: Failed to begin transaction: %v", tx.Error)
		return ErrDatabaseError
	}

	err := tx.Unscoped().Where("auth0_id = ?", auth0ID).First(&user).Error
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		log.Printf("ERROR: Failed to find user with auth0ID %s: %v", auth0ID, err)
		return ErrDatabaseError
	}

	// ResourceReservations
	if err := tx.Exec(`
		DELETE FROM "ResourceReservations"
		WHERE project_id IN (SELECT id FROM "Projects" WHERE user_id = ?)`, user.ID).Error; err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to delete resource reservations for user %d: %v", user.ID, err)
		return ErrDatabaseError
	}

	// ResourceSpecifications
	if err := tx.Exec(`
		DELETE FROM "ResourceSpecifications"
		WHERE project_id IN (SELECT id FROM "Projects" WHERE user_id = ?)`, user.ID).Error; err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to delete resource specifications for user %d: %v", user.ID, err)
		return ErrDatabaseError
	}

	// Projects
	if err := tx.Unscoped().Where("user_id = ?", user.ID).Delete(&models.Project{}).Error; err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to delete projects for user %d: %v", user.ID, err)
		return ErrDatabaseError
	}

	// Items
	if err := tx.Exec(`
		DELETE FROM "Items"
		WHERE item_type_id IN (SELECT id FROM "ItemTypes" WHERE user_id = ?)`, user.ID).Error; err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to delete items for user %d: %v", user.ID, err)
		return ErrDatabaseError
	}

	// ItemTypes
	if err := tx.Unscoped().Where("user_id = ?", user.ID).Delete(&models.ItemType{}).Error; err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to delete item types for user %d: %v", user.ID, err)
		return ErrDatabaseError
	}

	// Tags
	if err := tx.Unscoped().Where("user_id = ?", user.ID).Delete(&models.Tag{}).Error; err != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to delete tags for user %d: %v", user.ID, err)
		return ErrDatabaseError
	}

	deleteUserResult := tx.Unscoped().Delete(&user)
	if deleteUserResult.Error != nil {
		tx.Rollback()
		log.Printf("ERROR: Failed to delete user with ID %d: %v", user.ID, deleteUserResult.Error)
		return ErrDatabaseError
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("ERROR: Failed to commit transaction: %v", err)
		return ErrDatabaseError
	}

	return nil
}
