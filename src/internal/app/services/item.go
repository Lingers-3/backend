package services

import (
	"context"
	"time"

	"pocketeer/internal/app/services"
	"pocketeer/internal/platform/database"
	"pocketeer/internal/platform/database/models"
)

type ItemService struct {
	db *database.DB
}

func NewItemService(db *database.DB) *ItemService {
	return &ItemService{db}
}

type CreateItemRequest struct {
	ItemTypeID uint

	Description            *string
	Quantity               *float32
	ExpirationDate         *time.Time
	DisplayMeasurementUnit *string
	PurchasePrice          *float32
}
func (s *ItemService) CreateItem(ctx context.Context, req CreateItemRequest) (*models.Item, error) {
	return nil, services.NewErrNotImplemented("CreateItem")
}

type UpdateItemRequest struct {
	Description            *string
	Quantity               *float32
	ExpirationDate         *time.Time
	DisplayMeasurementUnit *string
	PurchasePrice          *float32
}

func (s *ItemService) UpdateItem(ctx context.Context, req UpdateItemRequest) (*models.Item, error) {
	return nil, services.NewErrNotImplemented("UpdateItem")
}

type DeleteItemRequest struct {
	ID uint
}

func (s *ItemService) DeleteItem(ctx context.Context, req DeleteItemRequest) (error) {
	return nil, services.NewErrNotImplemented("DeleteItem")
}

func (s *ItemService) GetItems(ctx context.Context) ([]*models.Item, error) {
	return nil, services.NewErrNotImplemented("GetItems")
}
