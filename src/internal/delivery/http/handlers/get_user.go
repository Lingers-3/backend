package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"pocketeer/internal/config"
	"pocketeer/internal/delivery/http/middleware"

	"github.com/labstack/echo/v4"
)

func GetUserHandler(cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		url := fmt.Sprintf("https://%s/userinfo", cfg.Auth0Domain)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Printf("Failed to create Auth0 request: %v", err)
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		token, ok := middleware.GetAccessTokenFromRequest(c)

		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Auth0 request failed: %v", err)
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to read response body")
		}
		defer resp.Body.Close()

		return c.JSONBlob(http.StatusOK, bodyBytes)
	}
}
