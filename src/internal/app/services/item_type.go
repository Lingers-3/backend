package services

import (
	"context"

	"pocketeer/internal/platform/database"
	"pocketeer/internal/platform/database/models"
)

type CreateItemTypeRequest struct {
	Name                string   `json:"name"`
	BaseMeasurementUnit string   `json:"base_measurement_unit"`
	Description         *string  `json:"description"`
	Category            *string  `json:"category"`
	DefaultQuantity     *float32 `json:"default_quantity"`
	Width               *float32 `json:"width"`
	Height              *float32 `json:"height"`
	Depth               *float32 `json:"depth"`
}

type ItemTypeService struct {
	db *database.DB
}

func NewItemTypeService(db *database.DB) *ItemTypeService {
	return &ItemTypeService{db}
}

func (s *ItemTypeService) CreateItemType(ctx context.Context, auth0ID string, req CreateItemTypeRequest) (*models.ItemType, error) {
	ID, err := models.GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return nil, err
	}

	itemType := models.ItemType{
		UserID:              ID,
		Name:                req.Name,
		Description:         req.Description,
		Category:            req.Category,
		BaseMeasurementUnit: req.BaseMeasurementUnit,
		Width:               req.Width,
		Height:              req.Height,
		Depth:               req.Depth,
		DefaultQuantity:     req.DefaultQuantity,
	}

	result := s.db.WithContext(ctx).Create(&itemType)
	if result.Error != nil {
		return nil, result.Error
	}

	return &itemType, nil
}
