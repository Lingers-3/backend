package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"pocketeer/internal/config"
	"pocketeer/internal/delivery/http/middleware"

	"github.com/labstack/echo/v4"
)

type UpdatePasswordRequest struct {
	// TODO(noatu): why camel case?
	NewPassword string `json:"newPassword"`
}

// UpdatePasswordHandler changes user password
// @Summary      Update user password
// @Description  Change the authenticated user's password via Auth0
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        password  body      UpdatePasswordRequest  true  "New Password"
// @Success      200       {object}  map[string]string      "message: password updated successfully"
// @Failure      400       {object}  echo.HTTPError
// @Failure      401       {object}  echo.HTTPError
// @Failure      500       {object}  echo.HTTPError
// @Router       /user/password [patch]
// @Security     BearerAuth
func UpdatePasswordHandler(cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var pwd UpdatePasswordRequest
		if err := c.Bind(&pwd); err != nil {
			log.Printf("Failed to parse request body: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
		}

		auth0ID, ok := middleware.GetAuth0IDFromRequest(c)
		if !ok || auth0ID == "" {
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		url := fmt.Sprintf("https://%s/api/v2/users/%s", cfg.Auth0Domain, auth0ID)
		payload := map[string]string{
			"password":   pwd.NewPassword,
			"connection": "Username-Password-Authentication",
		}
		body, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to encode request payload: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(body))

		token, err := getManagementToken(cfg.Auth0Domain, cfg.Auth0ClientID, cfg.Auth0ClientSecret)

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Auth0 request failed: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			log.Printf("Auth0 responded with status %d: %s", resp.StatusCode, string(bodyBytes))
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "password updated successfully",
		})
	}
}

func getManagementToken(domain, clientID, clientSecret string) (string, error) {
	url := fmt.Sprintf("https://%s/oauth/token", domain)

	payload := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"audience":      fmt.Sprintf("https://%s/api/v2/", domain),
		"grant_type":    "client_credentials",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	token, ok := result["access_token"]
	if !ok {
		return "", fmt.Errorf("access_token not found in response")
	}

	return token.(string), nil
}
