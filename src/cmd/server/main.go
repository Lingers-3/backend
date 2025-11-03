package main

import (
	"encoding/gob"
	"fmt"
	"log"

	"pocketeer/internal/app/services"
	"pocketeer/internal/config"
	"pocketeer/internal/delivery/http/handlers"
	"pocketeer/internal/delivery/http/middleware"
	"pocketeer/internal/platform/authenticator"
	"pocketeer/internal/platform/database"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func init() {
	gob.Register(map[string]any{})
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, relying on environment variables")
	}

	cfg := config.Load()

	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize the database connection: %v", err)
	}

	// Close the database on program exit
	defer func() {
		log.Println("INFO: closing the database")
		if err := database.Close(db); err != nil {
			log.Printf("ERROR: %v", err)
			return
		}
		log.Println("INFO: database closed")
	}()

	authenticator, err := authenticator.New(cfg.Auth0Domain, cfg.Auth0ClientID, cfg.Auth0ClientSecret, cfg.Auth0CallbackURL)
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}
	authMiddleware := middleware.AuthMiddleware(authenticator)

	itemTypeHandler := handlers.NewItemTypeHandler(services.NewItemTypeService(db))
	itemHandler := handlers.NewItemHandler(services.NewItemService(db))

	e := echo.New()
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(cfg.SessionSecret))))

	api := e.Group("/api")

	auth := api.Group("/auth")
	auth.GET("/login", handlers.LoginHandler(authenticator))
	auth.GET("/callback", handlers.CallbackHandler(authenticator, db))
	auth.GET("/logout", handlers.LogoutHandler(cfg))
	auth.POST("/change-password", handlers.UpdatePasswordHandler(cfg))

	users := api.Group("/users")
	users.GET("/me", handlers.ProfileHandler, authMiddleware)

	itemTypeHandler.RegisterRoutes(api, authMiddleware)
	itemHandler.RegisterRoutes(api, authMiddleware)

	// TODO(pencelheimer): Echo already prints this info
	// NOTE(noatu): Then delete it?
	socket := fmt.Sprintf("%s:%s", cfg.AppAddress, cfg.AppPort)
	log.Printf("Server listening on http://%s/", socket)
	e.Logger.Fatal(e.Start(socket))
}
