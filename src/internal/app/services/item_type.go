package services

import (
	"context"
	"errors"
	"log"
	"time"

	"pocketeer/internal/platform/database"
	"pocketeer/internal/platform/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	ItemIDs                []uint     `json:"item_ids"`
	TagIDs                 []uint     `json:"tag_ids"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	DeletedAt              *time.Time `json:"deleted_at,omitempty"`
}

func ItemTypeFromModel(m *models.ItemType) *ItemType {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	itemIDs := make([]uint, len(m.Items))
	for i, item := range m.Items {
		itemIDs[i] = item.ID
	}

	tagIDs := make([]uint, len(m.Tags))
	for i, tag := range m.Tags {
		tagIDs[i] = tag.ID
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
		ItemIDs:                itemIDs,
		TagIDs:                 tagIDs,
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
	PictureID              *uint    `json:"picture_id"`
	TagIDs                 []uint   `json:"tag_ids"`
}

func (s *ItemTypeService) Create(ctx context.Context, auth0ID string, req ItemTypeCreateRequest) (*ItemType, error) {
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return nil, err
	}

	itemType := &models.ItemType{
		UserID:                 userID,
		Name:                   req.Name,
		Description:            req.Description,
		BaseMeasurementUnit:    req.BaseMeasurementUnit,
		DisplayMeasurementUnit: req.DisplayMeasurementUnit,
		DefaultQuantity:        req.DefaultQuantity,
		ShortageThreshold:      req.ShortageTreshold,
		PictureID:              req.PictureID,
	}

	result := s.db.WithContext(ctx).Create(itemType)
	if result.Error != nil {
		log.Printf("ERROR: creating item type: %v", err)
		return nil, ErrDatabaseError
	}

	// Fetch only existing tags
	// NOTE(noatu): Not sure about handling errors, don't want to abort operation because of tags.
	// The tag id list will be returned, so frontend may check if it is correct.
	var tags []models.Tag
	if len(req.TagIDs) > 0 {
		err = s.db.WithContext(ctx).
			Where("id IN ? AND user_id = ?", req.TagIDs, userID).
			Find(&tags).Error
		if err != nil {
			log.Printf("ERROR: fetching tags: %v", err)
			// ignore
		}
	}
	// Associate tags
	if len(tags) > 0 {
		err = s.db.WithContext(ctx).Model(itemType).Association("Tags").Append(tags)
		if err != nil {
			log.Printf("ERROR: associating tags: %v", err)
			// ignore
		} else {
			itemType.Tags = tags // HACK(noatu): no need to refetch
		}
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
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			// Foreign key must be selected: https://gorm.io/gen/associations.html#Preload-with-select
			return db.Select("id", "item_type_id")
		}).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Select("id")
		}).
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
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			// Foreign key must be selected: https://gorm.io/gen/associations.html#Preload-with-select
			return db.Select("id", "item_type_id")
		}).
		Preload("Tags", func(db *gorm.DB) *gorm.DB {
			return db.Select("id")
		}).
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
	PictureID              *uint    `json:"picture_id"`
	TagIDs                 *[]uint  `json:"tag_ids"`
	Restore                *bool    `json:"restore"` // true = restore soft-deleted item
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
	if req.PictureID != nil {
		updates["picture_id"] = *req.PictureID
	}
	if req.Restore != nil && *req.Restore {
		updates["deleted_at"] = nil
	}

	if len(updates) > 0 {
		err = s.db.WithContext(ctx).
			Model(&itemType).
			Unscoped(). // include soft-deleted records, for restore scenarios
			Clauses(clause.Returning{}).
			Updates(updates).Error
		if err != nil {
			log.Printf("ERROR: updating item type: %v", err)
			return nil, ErrDatabaseError
		}
	}

	// NOTE(noatu): see the Create function
	if req.TagIDs != nil {
		var tags []models.Tag

		if len(*req.TagIDs) > 0 {
			err = s.db.WithContext(ctx).
				Where("id IN ? AND user_id = ?", *req.TagIDs, userID).
				Find(&tags).Error
			if err != nil {
				log.Printf("ERROR: fetching tags for update: %v", err)
				// ignore
			}
		}

		err = s.db.WithContext(ctx).Model(&itemType).Association("Tags").Replace(tags)
		if err != nil {
			log.Printf("ERROR: replacing associated tags: %v", err)
			// ignore
		} else {
			itemType.Tags = tags
		}
	}

	return ItemTypeFromModel(&itemType), nil
}

func (s *ItemTypeService) Delete(ctx context.Context, auth0ID string, itemTypeID uint, hard bool) (bool, error) {
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return false, err
	}

	var itemType models.ItemType
	err = s.db.WithContext(ctx).
		Unscoped().
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

	db := s.db.WithContext(ctx)
	if hard {
		db = db.Unscoped()
	}
	result := db.Delete(&itemType)

	if result.Error != nil {
		if hard && errors.Is(err, gorm.ErrForeignKeyViolated) {
			log.Printf("ERROR: item type hard delete violated constraint: %v", result.Error)
			return false, ErrForeignKeyViolated
		}

		log.Printf("ERROR: item type delete: %v", result.Error)
		return false, ErrDatabaseError
	}

	return hard, nil
}
