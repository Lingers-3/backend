package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"pocketeer/internal/config"
	"pocketeer/internal/delivery/http/middleware"
	"pocketeer/internal/platform/db/models"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func DeleteUserHandler(cfg *config.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		auth0ID, ok := middleware.GetAuth0IDFromRequest(c)
		if !ok || auth0ID == "" {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		url := fmt.Sprintf("https://%s/api/v2/users/%s", cfg.Auth0Domain, auth0ID)

		token, err := getManagementToken(cfg.Auth0Domain, cfg.Auth0ClientID, cfg.Auth0ClientSecret)

		if err != nil {
			log.Printf("Failed to get Auth0 management token: %v", err)
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			log.Printf("Failed to create Auth0 request: %v", err)
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Auth0 request failed: %v", err)
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			bodyBytes, _ := io.ReadAll(resp.Body)
			log.Printf("Auth0 responded with status %d: %s", resp.StatusCode, string(bodyBytes))
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		result := db.Unscoped().Where("auth0_id = ?", auth0ID).Delete(&models.User{})
		if result.Error != nil {
			log.Printf("DB error: %v", result.Error)
			return echo.NewHTTPError(http.StatusUnauthorized)
		}
		if result.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "user deleted successfully",
		})
	}
}
