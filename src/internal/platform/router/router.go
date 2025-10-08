package router

import (
	"pocketeer/internal/delivery/http"
	"pocketeer/internal/platform/authenticator"

	"github.com/labstack/echo/v4"
)

func New(e *echo.Echo, auth *authenticator.Authenticator) *echo.Echo {
	api := e.Group("/api")

	api.GET("/login", http.LoginHandler(auth))
	api.GET("/callback", http.CallbackHandler(auth))
	api.GET("/profile", http.ProfileHandler)
	return e
}
