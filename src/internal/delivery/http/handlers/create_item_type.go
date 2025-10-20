package handlers

import (
	// "bytes"
	// "encoding/json"
	// "fmt"
	// "io"
	"log"
	"net/http"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func CreateItemTypeHandler(c echo.Context) error {
	sess, _ := session.Get("session", c)
	auth0ID, ok := sess.Values["auth0_id"].(string)
	if !ok || auth0ID == "" {
		// TODO(noatu): this log is utterly useless
		log.Printf("auth0ID not found in session")
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized: invalid session",
		})
	}

	// TODO(noatu): Pictures...
	var payload struct {
		Name                string   `json:"name"`
		BaseMeasurementUnit string   `json:"base_measurement_unit"`
		Description         *string  `json:"description"`
		Category            *string  `json:"category"`
		DefaultQuantity     *float32 `json:"default_quantity"`
		Width               *float32 `json:"width"`
		Height              *float32 `json:"height"`
		Depth               *float32 `json:"depth"`
	}
	if err := c.Bind(&payload); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
	}

	// TODO(noatu): payload validation

	// TODO(noatu): write to database

	return c.JSON(http.StatusOK, map[string]string{})

}
