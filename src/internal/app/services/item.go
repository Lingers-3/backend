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

type ItemService struct {
	db *database.DB
}

func NewItemService(db *database.DB) *ItemService {
	return &ItemService{db}
}

type Item struct {
	ID                     uint       `json:"id"`
	Description            *string    `json:"description"`
	Quantity               float32    `json:"quantity"`
	ExpirationDate         *time.Time `json:"expiration_date"`
	DisplayMeasurementUnit string     `json:"display_measurement_unit"`
	PurchasePrice          *float32   `json:"purchase_price"`
	ItemTypeID             uint       `json:"item_type_id"`
	TagIDs                 []uint     `json:"tag_ids"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	DeletedAt              *time.Time `json:"deleted_at,omitempty"`
}

func ItemFromModel(m *models.Item) *Item {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	tagIDs := make([]uint, len(m.Tags))
	for i, tag := range m.Tags {
		tagIDs[i] = tag.ID
	}

	return &Item{
		ID:                     m.ID,
		Description:            m.Description,
		Quantity:               m.Quantity,
		ExpirationDate:         m.ExpirationDate,
		DisplayMeasurementUnit: m.DisplayMeasurementUnit,
		PurchasePrice:          m.PurchasePrice,
		ItemTypeID:             m.ItemTypeID,
		TagIDs:                 tagIDs,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
		DeletedAt:              deletedAt,
	}
}

type ItemCreateRequest struct {
	ItemTypeID             uint       `json:"item_type_id"`
	Description            *string    `json:"description"`
	Quantity               *float32   `json:"quantity"`
	ExpirationDate         *time.Time `json:"expiration_date"`
	DisplayMeasurementUnit *string    `json:"display_measurement_unit"`
	PurchasePrice          *float32   `json:"purchase_price"`
	TagIDs                 []uint     `json:"tag_ids"`
}

func (s *ItemService) Create(ctx context.Context, auth0ID string, req ItemCreateRequest) (*Item, error) {
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return nil, err
	}

	defaults, err := models.GetItemTypeDefaultFields(ctx, s.db, req.ItemTypeID, userID)
	if err != nil {
		log.Printf("ERROR: retrieving item type default fields: %v", err)
		return nil, ErrItemTypeNotFound
	}

	item := models.Item{
		ItemTypeID:             req.ItemTypeID,
		Description:            req.Description,
		Quantity:               defaults.DefaultQuantity,
		ExpirationDate:         req.ExpirationDate,
		DisplayMeasurementUnit: defaults.DefaultDisplayMeasurementUnit,
		PurchasePrice:          req.PurchasePrice,
	}

	if req.Quantity != nil {
		item.Quantity = *req.Quantity
	}

	if req.DisplayMeasurementUnit != nil {
		item.DisplayMeasurementUnit = *req.DisplayMeasurementUnit
	}

	result := s.db.WithContext(ctx).Create(&item)
	if result.Error != nil {
		log.Printf("ERROR: creating item: %v", err)
		return nil, ErrDatabaseError
	}

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

	if len(tags) > 0 {
		err = s.db.WithContext(ctx).Model(&item).Association("Tags").Append(tags)
		if err != nil {
			log.Printf("ERROR: associating tags: %v", err)
			// ignore
		} else {
			item.Tags = tags
		}
	}

	return ItemFromModel(&item), nil
}

func (s *ItemService) Get(ctx context.Context, auth0ID string, itemID uint) (*Item, error) {
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return nil, err
	}

	var item models.Item
	err = s.db.WithContext(ctx).
		Unscoped().
		Preload("Tags", func(db *gorm.DB) *gorm.DB { return db.Select("id") }).
		Joins("JOIN ItemTypes ON ItemTypes.id = Items.item_type_id").
		Where("ItemTypes.user_id = ? AND Items.id = ?", userID, itemID).
		First(&item).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrItemNotFound
		}
		log.Printf("ERROR: fetching item: %v", err)
		return nil, ErrDatabaseError
	}

	return ItemFromModel(&item), nil
}

func (s *ItemService) GetAll(ctx context.Context, auth0ID string) ([]*Item, error) {
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return nil, err
	}

	var items []models.Item
	err = s.db.WithContext(ctx).
		Unscoped().
		Preload("Tags", func(db *gorm.DB) *gorm.DB { return db.Select("id") }).
		Joins("JOIN ItemTypes ON ItemTypes.id = Items.item_type_id").
		Where("ItemTypes.user_id = ?", userID).
		Find(&items).Error

	if err != nil {
		log.Printf("ERROR: fetching items: %v", err)
		return nil, ErrDatabaseError
	}

	responses := make([]*Item, len(items))
	for i := range items {
		responses[i] = ItemFromModel(&items[i])
	}

	return responses, nil
}

type ItemUpdateRequest struct {
	Description            *string    `json:"description"`
	Quantity               *float32   `json:"quantity"`
	ExpirationDate         *time.Time `json:"expiration_date"`
	DisplayMeasurementUnit *string    `json:"display_measurement_unit"`
	PurchasePrice          *float32   `json:"purchase_price"`
	TagIDs                 *[]uint    `json:"tag_ids"`
}

func (s *ItemService) Update(ctx context.Context, auth0ID string, ID uint, req ItemUpdateRequest) (*Item, error) {
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return nil, err
	}

	var item models.Item
	result := s.db.WithContext(ctx).
		Preload("Tags").
		Joins("JOIN ItemTypes ON ItemTypes.id = Items.item_type_id").
		Where("ItemTypes.user_id = ? AND Items.id = ?", userID, ID).
		First(&item)

	if result.Error != nil {
		log.Printf("ERROR: retrieving item: %v", result.Error)
		return nil, ErrItemNotFound
	}

	updates := make(map[string]any)

	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.Quantity != nil {
		updates["quantity"] = *req.Quantity
	}
	if req.ExpirationDate != nil {
		updates["expiration_date"] = req.ExpirationDate
	}
	if req.DisplayMeasurementUnit != nil {
		updates["display_measurement_unit"] = *req.DisplayMeasurementUnit
	}
	if req.PurchasePrice != nil {
		updates["purchase_price"] = req.PurchasePrice
	}

	if len(updates) > 0 {
		result = s.db.WithContext(ctx).Model(&item).Updates(updates)
		if result.Error != nil {
			log.Printf("ERROR: updating item: %v", result.Error) // Використовуйте result.Error
			return nil, ErrDatabaseError
		}
	}

	if req.TagIDs != nil {
		var tags []models.Tag

		if len(*req.TagIDs) > 0 {
			err = s.db.WithContext(ctx).
				Where("id IN ? AND user_id = ?", *req.TagIDs, userID).
				Find(&tags).Error

			if err != nil {
				log.Printf("ERROR: fetching tags for update: %v", err)
				return nil, ErrDatabaseError
			}
		}

		err = s.db.WithContext(ctx).Model(&item).Association("Tags").Replace(tags)
		if err != nil {
			log.Printf("ERROR: replacing associated tags: %v", err)
			return nil, ErrDatabaseError
		}

		item.Tags = tags
	}

	return ItemFromModel(&item), nil
}

func (s *ItemService) Delete(ctx context.Context, auth0ID string, itemID uint) (hard bool, err error) {
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return false, err
	}

	var item models.Item
	err = s.db.WithContext(ctx).
		Unscoped().
		Select("Items.id").
		Joins("JOIN ItemTypes ON ItemTypes.id = Items.item_type_id").
		Where("ItemTypes.user_id = ? AND Items.id = ?", userID, itemID).
		First(&item).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrItemNotFound
		}
		log.Printf("ERROR: searching for item: %v", err)
		return false, ErrDatabaseError
	}

	err = s.db.WithContext(ctx).Unscoped().Delete(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrForeignKeyViolated) {
			err = s.db.WithContext(ctx).Delete(&item).Error
			if err != nil {
				log.Printf("ERROR: item soft delete: %v", err)
				return false, ErrDatabaseError
			}
			return false, nil
		}

		log.Printf("ERROR: item hard delete: %v", err)
		return false, ErrDatabaseError
	}

	return true, nil
}
