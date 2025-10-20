package router

import (
	"pocketeer/internal/config"
	"pocketeer/internal/delivery/http/handlers"
	"pocketeer/internal/delivery/http/middleware"
	"pocketeer/internal/platform/authenticator"
	"pocketeer/internal/platform/database"

	"github.com/labstack/echo/v4"
)

func New(e *echo.Echo, authenticator *authenticator.Authenticator, db *database.DB, cfg *config.Config) *echo.Echo {
	api := e.Group("/api")

	auth := api.Group("/auth")
	auth.GET("/login", handlers.LoginHandler(authenticator))
	auth.GET("/callback", handlers.CallbackHandler(authenticator, db))
	auth.GET("/logout", handlers.LogoutHandler(cfg))
	auth.POST("/change-password", handlers.UpdatePasswordHandler(cfg))

	users := api.Group("/users")
	users.GET("/me", handlers.ProfileHandler, middleware.AuthMiddleware(authenticator))

	item_types := api.Group("/item-types")
	item_types.POST("/create", handlers.CreateItemTypeHandler, middleware.AuthMiddleware(authenticator))

	return e
}
