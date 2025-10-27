package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"pocketeer/internal/config"
	"pocketeer/internal/platform/db/models"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func DeleteUserHandler(cfg *config.Config, db *gorm.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			log.Printf("session get error: %v", err)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		auth0ID, ok := sess.Values["auth0_id"].(string)
		if !ok || auth0ID == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		url := fmt.Sprintf("https://%s/api/v2/users/%s", cfg.Auth0Domain, auth0ID)

		token, err := getManagementToken(cfg.Auth0Domain, cfg.Auth0ClientID, cfg.Auth0ClientSecret)

		if err != nil {
			log.Printf("Failed to get Auth0 management token: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "internal error",
			})
		}

		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			log.Printf("Failed to create Auth0 request: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "internal error",
			})
		}

		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Auth0 request failed: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "internal error",
			})
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNoContent {
			bodyBytes, _ := io.ReadAll(resp.Body)
			log.Printf("Auth0 responded with status %d: %s", resp.StatusCode, string(bodyBytes))
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "internal error",
			})
		}

		result := db.Unscoped().Where("auth0_id = ?", auth0ID).Delete(&models.User{})
		if result.Error != nil {
			log.Printf("DB error: %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "internal error",
			})
		}
		if result.RowsAffected == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "user not found",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "user deleted successfully",
		})
	}
}
