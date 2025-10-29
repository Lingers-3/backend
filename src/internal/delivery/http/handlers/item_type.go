package handlers

import (
	"log"
	"net/http"

	"pocketeer/internal/app/services"
	"pocketeer/internal/delivery/http/middleware"

	"github.com/labstack/echo/v4"
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

type CreateItemTypeResponse struct {
	// ID uint `json:"id"` // created item type id
	// Error *uint `json:"error"`
}

type DeleteItemTypeRequest struct {
	ID uint `json:"id"`
}

type DeleteItemTypeResponse struct {
	Hard bool `json:"hard"` // false = soft delete
	// Error *uint `json:"error"`
}

type ItemTypeHandler struct {
	service *services.ItemTypeService
}

func NewItemTypeHandler(service *services.ItemTypeService) *ItemTypeHandler {
	return &ItemTypeHandler{service}
}

// TODO(noatu): Pictures...
func (h *ItemTypeHandler) CreateItemType(c echo.Context) error {
	var payload CreateItemTypeRequest
	if err := c.Bind(&payload); err != nil {
		log.Printf("ERROR: parsing request body: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	// QUESTION(noatu): should authorization be checked before http request?
	auth0ID, err := middleware.GetAuth0ID(c)
	if err != nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	serviceReq := services.CreateItemTypeRequest{
		Auth0ID:             auth0ID,
		Name:                payload.Name,
		BaseMeasurementUnit: payload.BaseMeasurementUnit,
		Description:         payload.Description,
		Category:            payload.Category,
		DefaultQuantity:     payload.DefaultQuantity,
		Width:               payload.Width,
		Height:              payload.Height,
		Depth:               payload.Depth,
	}

	_, err = h.service.CreateItemType(c.Request().Context(), serviceReq)
	if err != nil {
		log.Printf("ERROR: creating item type: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, CreateItemTypeResponse{})
}
