package services

import (
	"context"

	"pocketeer/internal/platform/database"
	"pocketeer/internal/platform/database/models"
)

type ItemTypeService struct {
	db *database.DB
}

func NewItemTypeService(db *database.DB) *ItemTypeService {
	return &ItemTypeService{db}
}

type CreateItemTypeRequest struct {
	Name                string  `json:"name"`
	Description         *string `json:"description"`
	BaseMeasurementUnit string  `json:"base_measurement_unit"`
	// QUESTION(noatu): may I drop "Default" all over the codebase?
	DefaultDisplayMeasurementUnit string   `json:"default_display_measurement_unit"`
	DefaultQuantity               *float32 `json:"default_quantity"`
	ShortageTreshold              *float32 `json:"shortage_threshold"`
}

func (s *ItemTypeService) CreateItemType(ctx context.Context, auth0ID string, req CreateItemTypeRequest) (*models.ItemType, error) {
	ID, err := models.GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return nil, err
	}

	itemType := models.ItemType{
		UserID:                        ID,
		Name:                          req.Name,
		Description:                   req.Description,
		BaseMeasurementUnit:           req.BaseMeasurementUnit,
		DefaultDisplayMeasurementUnit: req.DefaultDisplayMeasurementUnit,
		DefaultQuantity:               req.DefaultQuantity,
		ShortageThreshold:             req.ShortageTreshold,
	}

	result := s.db.WithContext(ctx).Create(&itemType)
	if result.Error != nil {
		return nil, result.Error
	}

	return &itemType, nil
}
