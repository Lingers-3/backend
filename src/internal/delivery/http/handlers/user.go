package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"pocketeer/internal/app/services"
	"pocketeer/internal/config"
	"pocketeer/internal/delivery/http/middleware"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service *services.UserService
	cfg     *config.Config
}

type User struct {
	Auth0ID       string `json:"sub"`
	Nickname      string `json:"nickname"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func NewUserHandler(service *services.UserService, cfg *config.Config) *UserHandler {
	return &UserHandler{service, cfg}
}

func (h *UserHandler) RegisterRoutes(router *echo.Group, middlewares ...echo.MiddlewareFunc) {
	group := router.Group("/users", middlewares...)
	group.GET("/me", h.Get)
	group.DELETE("/me", h.Delete)
}

// Get retrieves authenticated user information
// @Summary      Get user profile
// @Description  Retrieve the authenticated user's profile information from Auth0
// @Tags         user
// @Accept       json
// @Produce      json
// @Success      200  {object}  User
// @Failure      401  {object}  echo.HTTPError
// @Failure      403  {object}  echo.HTTPError  "Email not verified"
// @Failure      500  {object}  echo.HTTPError
// @Router       /user/me [get]
// @Security     BearerAuth
func (h *UserHandler) Get(c echo.Context) error {
	url := fmt.Sprintf("https://%s/userinfo", h.cfg.Auth0Domain)

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

// Delete deletes the authenticated user
// @Summary      Delete user account
// @Description  Permanently delete the authenticated user's account from both Auth0 and the database
// @Tags         user
// @Accept       json
// @Produce      json
// @Success      204  {object}  nil             "user deleted successfully"
// @Failure      401  {object}  echo.HTTPError
// @Failure      404  {object}  echo.HTTPError  "user not found"
// @Failure      500  {object}  echo.HTTPError
// @Router       /user/me [delete]
// @Security     BearerAuth
func (h *UserHandler) Delete(c echo.Context) error {
	auth0ID, ok := middleware.GetAuth0IDFromRequest(c)
	if !ok || auth0ID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	url := fmt.Sprintf("https://%s/api/v2/users/%s", h.cfg.Auth0Domain, auth0ID)

	token, err := getManagementToken(h.cfg.Auth0Domain, h.cfg.Auth0ClientID, h.cfg.Auth0ClientSecret)

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

	err = h.service.DeleteUserByAuth0ID(c.Request().Context(), auth0ID)

	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}
