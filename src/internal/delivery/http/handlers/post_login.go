package handlers

import (
	"errors"
	"log"
	"net/http"

	"pocketeer/internal/platform/database/models"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func PostLoginHandler(db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var body struct {
			Sub           string `json:"sub"`
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
		}

		if err := c.Bind(&body); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid JSON body")
		}

		if body.Sub == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "missing sub")
		}

		if !body.EmailVerified {
			return echo.NewHTTPError(http.StatusForbidden, "email not verified")
		}

		var user models.User
		result := db.First(&user, "auth0_id = ?", body.Sub)

		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				user = models.User{
					Auth0ID: body.Sub,
					Email:   body.Email,
				}
				if err := db.Create(&user).Error; err != nil {
					log.Printf("failed to create user: %v", err)
					return echo.NewHTTPError(http.StatusInternalServerError)
				}
				log.Printf("New user created: %s (%s)", body.Sub, body.Email)

			} else {
				log.Printf("DB error: %v", result.Error)
				return echo.NewHTTPError(http.StatusInternalServerError)
			}
		}

		return c.NoContent(http.StatusOK)
	}
}
