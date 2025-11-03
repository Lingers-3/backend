package services

import (
	"context"
	"errors"
	"log"
	"time"

	"pocketeer/internal/platform/database"
	"pocketeer/internal/platform/database/models"

	"gorm.io/gorm"
)

type ItemTypeService struct {
	db *database.DB
}

func NewItemTypeService(db *database.DB) *ItemTypeService {
	return &ItemTypeService{db}
}

type ItemType struct { //  ∠( ᐛ 」∠)
	ID                     uint       `json:"id"`
	Name                   string     `json:"name"`
	Description            *string    `json:"description"`
	BaseMeasurementUnit    string     `json:"base_measurement_unit"`
	DisplayMeasurementUnit string     `json:"display_measurement_unit"`
	DefaultQuantity        *float32   `json:"default_quantity"`
	ShortageTreshold       *float32   `json:"shortage_threshold"`
	PictureID              *uint      `json:"picture_id,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	DeletedAt              *time.Time `json:"deleted_at,omitempty"`
}

func ItemTypeFromModel(m *models.ItemType) *ItemType {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return &ItemType{
		ID:                     m.ID,
		Name:                   m.Name,
		Description:            m.Description,
		BaseMeasurementUnit:    m.BaseMeasurementUnit,
		DisplayMeasurementUnit: m.DisplayMeasurementUnit,
		DefaultQuantity:        m.DefaultQuantity,
		ShortageTreshold:       m.ShortageThreshold,
		PictureID:              m.PictureID,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
		DeletedAt:              deletedAt,
	}
}

type ItemTypeCreateRequest struct {
	Name                   string   `json:"name"`
	Description            *string  `json:"description"`
	BaseMeasurementUnit    string   `json:"base_measurement_unit"`
	DisplayMeasurementUnit string   `json:"display_measurement_unit"`
	DefaultQuantity        *float32 `json:"default_quantity"`
	ShortageTreshold       *float32 `json:"shortage_threshold"`
}

func ItemTypeModelFromCreateRequest(r ItemTypeCreateRequest, userID uint) *models.ItemType {
	return &models.ItemType{
		UserID:                 userID,
		Name:                   r.Name,
		Description:            r.Description,
		BaseMeasurementUnit:    r.BaseMeasurementUnit,
		DisplayMeasurementUnit: r.DisplayMeasurementUnit,
		DefaultQuantity:        r.DefaultQuantity,
		ShortageThreshold:      r.ShortageTreshold,
	}
}

func (s *ItemTypeService) Create(ctx context.Context, auth0ID string, req ItemTypeCreateRequest) (*ItemType, error) {
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return nil, err
	}

	itemType := ItemTypeModelFromCreateRequest(req, userID)

	result := s.db.WithContext(ctx).Create(itemType)
	if result.Error != nil {
		log.Printf("ERROR: creating item type: %v", err)
		return nil, ErrDatabaseError
	}

	return ItemTypeFromModel(itemType), nil
}

func (s *ItemTypeService) Get(ctx context.Context, auth0ID string, itemTypeID uint) (*ItemType, error) {
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return nil, err
	}

	var itemType models.ItemType
	err = s.db.WithContext(ctx).
		Unscoped(). // include soft-deleted records
		Where("id = ? AND user_id = ?", itemTypeID, userID).
		First(&itemType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrItemTypeNotFound
		}
		log.Printf("ERROR: fetching item type: %v", err)
		return nil, ErrDatabaseError
	}

	return ItemTypeFromModel(&itemType), nil
}

func (s *ItemTypeService) GetAll(ctx context.Context, auth0ID string) ([]*ItemType, error) {
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return nil, err
	}

	var itemTypes []models.ItemType
	err = s.db.WithContext(ctx).
		Unscoped(). // include soft-deleted records
		Where("user_id = ?", userID).Find(&itemTypes).Error
	if err != nil {
		log.Printf("ERROR: fetching item types: %v", err)
		return nil, ErrDatabaseError
	}

	responses := make([]*ItemType, len(itemTypes))
	for i := range itemTypes {
		responses[i] = ItemTypeFromModel(&itemTypes[i])
	}

	return responses, nil
}

type ItemTypeUpdateRequest struct {
	Name                   *string  `json:"name"`
	Description            *string  `json:"description"`
	BaseMeasurementUnit    *string  `json:"base_measurement_unit"`
	DisplayMeasurementUnit *string  `json:"display_measurement_unit"`
	DefaultQuantity        *float32 `json:"default_quantity"`
	ShortageThreshold      *float32 `json:"shortage_threshold"`
}

func (s *ItemTypeService) Update(ctx context.Context, auth0ID string, itemTypeID uint, req ItemTypeUpdateRequest) (*ItemType, error) {
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return nil, err
	}

	// Verify existence and ownership
	var itemType models.ItemType
	err = s.db.WithContext(ctx).
		Unscoped(). // include soft-deleted records, for restore scenarios
		Where("id = ? AND user_id = ?", itemTypeID, userID).
		First(&itemType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrItemTypeNotFound
		}
		log.Printf("ERROR: fetching item type: %v", err)
		return nil, ErrDatabaseError
	}

	// Update via map https://gorm.io/docs/update.html#Updates-multiple-columns
	updates := make(map[string]any)
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.BaseMeasurementUnit != nil {
		updates["base_measurement_unit"] = *req.BaseMeasurementUnit
	}
	if req.DisplayMeasurementUnit != nil {
		updates["display_measurement_unit"] = *req.DisplayMeasurementUnit
	}
	if req.DefaultQuantity != nil {
		updates["default_quantity"] = *req.DefaultQuantity
	}
	if req.ShortageThreshold != nil {
		updates["shortage_threshold"] = *req.ShortageThreshold
	}

	err = s.db.WithContext(ctx).Model(&itemType).Updates(updates).Error
	if err != nil {
		log.Printf("ERROR: updating item type: %v", err)
		return nil, ErrDatabaseError
	}

	return ItemTypeFromModel(&itemType), nil
}

func (s *ItemTypeService) Delete(ctx context.Context, auth0ID string, itemTypeID uint) (hard bool, err error) {
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return false, err
	}

	// Verify existence and ownership
	var itemType models.ItemType
	err = s.db.WithContext(ctx).
		Unscoped(). // include soft-deleted records
		Select("id").
		Where("id = ? AND user_id = ?", itemTypeID, userID).
		First(&itemType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrItemTypeNotFound
		}
		log.Printf("ERROR: searching for item type: %v", err)
		return false, ErrDatabaseError
	}

	// HACK(noatu): try hard delete first and rely on RESTRICT constraint
	err = s.db.WithContext(ctx).Unscoped().Delete(&itemType).Error
	if err != nil {
		// Soft Delete (more readable when nested imo)
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			err = s.db.WithContext(ctx).Delete(&itemType).Error
			if err != nil {
				log.Printf("ERROR: item type soft delete: %v", err)
				return false, ErrDatabaseError
			}
			return false, nil
		}

		log.Printf("ERROR: item type hard delete: %v", err)
		return false, ErrDatabaseError
	}

	return true, nil
}
