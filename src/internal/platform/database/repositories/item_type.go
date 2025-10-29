package repositories

import (
	"context"

	"pocketeer/internal/platform/database/models"

	"gorm.io/gorm"
)

// TODO(noatu): CRUD is an anti-pattern. Need to nuke repositories all together.
type ItemTypeRepository interface {
	Create(ctx context.Context, itemType *models.ItemType) error
	Read(ctx context.Context, id uint) (*models.ItemType, error)
	Update(ctx context.Context, itemType *models.ItemType) error
	HardDelete(ctx context.Context, id uint) error
	SoftDelete(ctx context.Context, id uint) error
}

type gormItemTypeRepository struct {
	db *gorm.DB
}

func NewGormItemTypeRepository(db *gorm.DB) ItemTypeRepository {
	return &gormItemTypeRepository{db: db}
}

func (r *gormItemTypeRepository) Create(ctx context.Context, itemType *models.ItemType) error {
	return r.db.WithContext(ctx).Create(itemType).Error
}

func (r *gormItemTypeRepository) Read(ctx context.Context, id uint) (*models.ItemType, error) {
	var itemType models.ItemType
	result := r.db.WithContext(ctx).First(&itemType, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &itemType, nil
}

func (r *gormItemTypeRepository) Update(ctx context.Context, itemType *models.ItemType) error {
	return r.db.WithContext(ctx).Save(itemType).Error
}

func (r *gormItemTypeRepository) HardDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&models.ItemType{}, id).Error
}

func (r *gormItemTypeRepository) SoftDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.ItemType{}, id).Error
}
