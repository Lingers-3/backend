package repositories

import (
	"context"

	"pocketeer/internal/platform/database/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	GetUserByAuth0ID(ctx context.Context, auth0ID string) (*models.User, error)
}

type gormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) GetUserByAuth0ID(ctx context.Context, auth0ID string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Select("id").First(&user, "auth0_id = ?", auth0ID)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}
