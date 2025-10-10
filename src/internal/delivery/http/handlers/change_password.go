package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"pocketeer/internal/config"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type UpdatePwdRequest struct {
	NewPassword string `json:"newPassword"`
}

func UpdatePasswordHandler(cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var pwd UpdatePwdRequest
		if err := c.Bind(&pwd); err != nil {
			log.Printf("Failed to parse request body: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
		}

		sess, _ := session.Get("session", c)
		sub, ok := sess.Values["sub"].(string)
		if !ok || sub == "" {
			log.Printf("Sub not found in session")
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "unauthorized: invalid session",
			})
		}

		url := fmt.Sprintf("https://%s/api/v2/users/%s", cfg.Auth0Domain, sub)
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
