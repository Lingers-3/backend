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
	db          *database.DB
	userService *UserService
}

func NewItemTypeService(db *database.DB, userService *UserService) *ItemTypeService {
	return &ItemTypeService{db, userService}
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

type ItemTypeFull struct {
	ID                     uint     `json:"id"`
	Name                   string   `json:"name"`
	Description            *string  `json:"description"`
	BaseMeasurementUnit    string   `json:"base_measurement_unit"`
	DisplayMeasurementUnit string   `json:"display_measurement_unit"`
	DefaultQuantity        *float32 `json:"default_quantity"`
	ShortageThreshold      *float32 `json:"shortage_threshold"`
	PictureID              *uint    `json:"picture_id,omitempty"`

	Items []ItemFull `json:"items"`
	Tags  []Tag      `json:"tags"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ItemTypeCreateRequest struct {
	Name                   string   `json:"name" validate:"required,max=256"`
	Description            *string  `json:"description" validate:"omitempty,max=512"`
	BaseMeasurementUnit    string   `json:"base_measurement_unit" validate:"required,max=256"`
	DisplayMeasurementUnit string   `json:"display_measurement_unit" validate:"required,max=256"`
	DefaultQuantity        *float32 `json:"default_quantity" validate:"omitempty,gte=0,lte=1000000"`
	ShortageTreshold       *float32 `json:"shortage_threshold" validate:"omitempty,gte=0,lte=1000000"`
	PictureID              *uint    `json:"picture_id" validate:"omitempty,gt=0"`
	TagIDs                 []uint   `json:"tag_ids" validate:"dive,gt=0"`
}

func (s *ItemTypeService) Create(ctx context.Context, auth0ID string, req ItemTypeCreateRequest) (*ItemType, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
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
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
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
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
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
	Name                   *string  `json:"name" validate:"omitempty,max=256"`
	Description            *string  `json:"description" validate:"omitempty,max=512"`
	BaseMeasurementUnit    *string  `json:"base_measurement_unit" validate:"omitempty,max=256"`
	DisplayMeasurementUnit *string  `json:"display_measurement_unit" validate:"omitempty,max=256"`
	DefaultQuantity        *float32 `json:"default_quantity" validate:"omitempty,gte=0,lte=1000000"`
	ShortageThreshold      *float32 `json:"shortage_threshold" validate:"omitempty,gte=0,lte=1000000"`
	PictureID              *uint    `json:"picture_id" validate:"omitempty,gt=0"`
	TagIDs                 *[]uint  `json:"tag_ids" validate:"omitempty,dive,gt=0"`
	Restore                *bool    `json:"restore"` // true = restore soft-deleted item
}

func (s *ItemTypeService) Update(ctx context.Context, auth0ID string, itemTypeID uint, req ItemTypeUpdateRequest) (*ItemType, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
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
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
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

func (s *ItemTypeService) GetAllFull(ctx context.Context, auth0ID string) ([]*ItemTypeFull, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	var itemTypes []models.ItemType

	err = s.db.WithContext(ctx).
		Unscoped().
		Preload("Tags").
		Preload("Items").
		Preload("Items.Tags").
		Where("user_id = ?", userID).
		Find(&itemTypes).Error

	if err != nil {
		log.Printf("ERROR: fetching full ItemType hierarchy: %v", err)
		return nil, ErrDatabaseError
	}

	responses := make([]*ItemTypeFull, len(itemTypes))
	for i := range itemTypes {
		responses[i] = ItemTypeFullFromModel(&itemTypes[i])
	}

	return responses, nil
}
