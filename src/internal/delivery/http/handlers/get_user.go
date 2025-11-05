package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"pocketeer/internal/config"
	"pocketeer/internal/delivery/http/middleware"

	"github.com/labstack/echo/v4"
)

type User struct {
	Auth0ID       string `json:"sub"`
	Nickname      string `json:"nickname"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

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
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		var user User
		if err := json.Unmarshal(bodyBytes, &user); err != nil {
			log.Printf("Failed to unmarshal Auth0 response: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		if !user.EmailVerified {
			return echo.NewHTTPError(http.StatusForbidden, "email not verified")
		}

		return c.JSON(http.StatusOK, user)
	}
}
