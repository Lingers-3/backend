package services

import (
	"context"

	"pocketeer/internal/platform/database/models"
	"pocketeer/internal/platform/database/repositories"
)

type CreateItemTypeRequest struct {
	Auth0ID             string
	Name                string
	BaseMeasurementUnit string
	Description         *string
	Category            *string
	DefaultQuantity     *float32
	Width               *float32
	Height              *float32
	Depth               *float32
}

type ItemTypeService struct {
	userRepo     repositories.UserRepository
	itemTypeRepo repositories.ItemTypeRepository
}

func NewItemTypeService(userRepo repositories.UserRepository, itemTypeRepo repositories.ItemTypeRepository) *ItemTypeService {
	return &ItemTypeService{userRepo, itemTypeRepo}
}

func (s *ItemTypeService) CreateItemType(ctx context.Context, req CreateItemTypeRequest) (*models.ItemType, error) {
	user, err := s.userRepo.GetUserByAuth0ID(ctx, req.Auth0ID)
	if err != nil {
		return nil, err
	}

	itemType := &models.ItemType{
		UserID:              user.ID,
		Name:                req.Name,
		Description:         req.Description,
		Category:            req.Category,
		BaseMeasurementUnit: req.BaseMeasurementUnit,
		Width:               req.Width,
		Height:              req.Height,
		Depth:               req.Depth,
		DefaultQuantity:     req.DefaultQuantity,
	}

	if err := s.itemTypeRepo.Create(ctx, itemType); err != nil {
		return nil, err
	}

	return itemType, nil
}
