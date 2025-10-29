package router

import (
	"pocketeer/internal/config"
	"pocketeer/internal/delivery/http/handlers"
	"pocketeer/internal/delivery/http/middleware"
	"pocketeer/internal/platform/authenticator"
	"pocketeer/internal/platform/db"

	"github.com/labstack/echo/v4"
)

func New(e *echo.Echo, authenticator *authenticator.Authenticator, db *db.DB, cfg *config.Config) *echo.Echo {
	api := e.Group("/api")

	auth := api.Group("/auth")
	auth.GET("/login", handlers.LoginHandler(authenticator))
	auth.GET("/callback", handlers.CallbackHandler(authenticator, db))
	auth.GET("/logout", handlers.LogoutHandler(cfg))
	auth.POST("/change-password", handlers.UpdatePasswordHandler(cfg))
	auth.GET("/post-login", handlers.PostLoginHandler(db), middleware.AuthMiddleware(authenticator))

	users := api.Group("/users")
	users.GET("/me", handlers.GetUserHandler(cfg), middleware.AuthMiddleware(authenticator))
	users.DELETE("/me", handlers.DeleteUserHandler(cfg, db), middleware.AuthMiddleware(authenticator))

	return e
}
