package handlers

import (
	"log"
	"net/http"
	"pocketeer/internal/delivery/http/middleware"
	"pocketeer/internal/platform/database/models"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
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
	ID uint `json:"id"` // created item type id
	// Error *uint `json:"error"`
}

// TODO(noatu): Pictures...
// FIXME(noatu): passing db to http handler is EVIL
func CreateItemTypeHandler(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		// TODO(noatu): payload data validation?

		var user models.User
		result := db.Select("id").First(&user, "auth0_id = ?", auth0ID)
		if result.Error != nil {
			log.Printf("ERROR: searching user in database: %v", result.Error)
			return c.NoContent(http.StatusInternalServerError)
		}

		// TODO(noatu): write to database
		item_type := models.ItemType{
			UserID: user.ID,

			Name:                payload.Name,
			Description:         payload.Description,
			Category:            payload.Category,
			BaseMeasurementUnit: payload.BaseMeasurementUnit,
			DefaultQuantity:     payload.DefaultQuantity,
			Width:               payload.Width,
			Height:              payload.Height,
			Depth:               payload.Depth,

			// PictureID:           payload.PictureID,
			// Picture:             payload.Picture,
		}

		result = db.Create(&item_type)
		if result.Error != nil {
			log.Printf("ERROR: creating item type in database: %v", result.Error)
			return c.NoContent(http.StatusInternalServerError)
		}

		return c.JSON(http.StatusOK, CreateItemTypeResponse{ID: item_type.ID})
	}
}
