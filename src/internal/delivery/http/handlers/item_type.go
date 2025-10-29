package handlers

import (
	"errors"
	"log"
	"net/http"

	"pocketeer/internal/app/services"
	"pocketeer/internal/delivery/http/middleware"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// type DeleteItemTypeRequest struct {
// 	ID uint `json:"id"`
// }

// type DeleteItemTypeResponse struct {
// 	Hard bool `json:"hard"` // false = soft delete
// 	// Error *uint `json:"error"`
// }

type ItemTypeHandler struct {
	service *services.ItemTypeService
}

func NewItemTypeHandler(service *services.ItemTypeService) *ItemTypeHandler {
	return &ItemTypeHandler{service}
}

// TODO(noatu): Pictures...
func (h *ItemTypeHandler) CreateItemType(c echo.Context) error {
	var payload services.CreateItemTypeRequest
	if err := c.Bind(&payload); err != nil {
		log.Printf("ERROR: parsing request body: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	// QUESTION(noatu): should authorization be checked before http request?
	auth0ID, err := middleware.GetAuth0ID(c)
	if err != nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	_, err = h.service.CreateItemType(c.Request().Context(), auth0ID, payload)
	if err != nil {
		log.Printf("ERROR: creating item type: %v", err)
		// HACK(noatu): TLDR: this is wrong, but no idea what is right
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		// WARN(noatu): sending full err will probably expose too much
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusOK)
}
