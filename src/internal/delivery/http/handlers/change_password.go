package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"pocketeer/internal/config"

	"github.com/labstack/echo/v4"
)

type UpdatePasswordRequest struct {
	NewPassword string `json:"new_password"`
}

// type UpdatePasswordResponse struct {
// 	Error *string `json:"error"`
// 	Message *string `json:"message"`
// }

// QUESTION(noatu): What's the point of returning meaningless messages?
// Is http.Status* not enough?
func UpdatePasswordHandler(cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var pwd UpdatePasswordRequest
		if err := c.Bind(&pwd); err != nil {
			log.Printf("Failed to parse request body: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
		}

		// QUESTION(noatu): should authorization be checked before http request?
		auth0ID, err := GetAuth0ID(c)
		if err != nil {
			// QUESTION(noatu): status code says it all?
			// return c.NoContent(http.StatusUnauthorized)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
		}

		url := fmt.Sprintf("https://%s/api/v2/users/%s", cfg.Auth0Domain, auth0ID)
		payload := map[string]string{
			"password":   pwd.NewPassword,
			"connection": "Username-Password-Authentication",
		}
		body, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to encode request payload: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "internal error",
			})
		}

		req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(body))

		token, err := getManagementToken(cfg.Auth0Domain, cfg.Auth0ClientID, cfg.Auth0ClientSecret)

		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Auth0 request failed: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "internal error",
			})
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			log.Printf("Auth0 responded with status %d: %s", resp.StatusCode, string(bodyBytes))
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "internal error",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "password updated successfully",
		})
	}
}

// QUESTION(noatu): This is OAuth, not public API per se,
// why not inline it? TokenType field seems unneeded (but what do I know?).
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
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

	var tokenRes TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return "", err
	}

	return tokenRes.AccessToken, nil
}
