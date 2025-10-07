package router

import (
	"pocketeer/internal/delivery/http"
	"pocketeer/internal/platform/authenticator"

	"github.com/labstack/echo/v4"
)

func New(e *echo.Echo, auth *authenticator.Authenticator) *echo.Echo {

	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "ok")
	})
	e.GET("/login", http.LoginHandler(auth))
	e.GET("/callback", http.CallbackHandler(auth))
	e.GET("/profile", http.ProfileHandler)
	return e
}
