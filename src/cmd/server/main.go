package main

import (
	"encoding/gob"
	"log"
	"pocketeer/internal/config"
	"pocketeer/internal/platform/authenticator"
	"pocketeer/internal/platform/router"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func init() {
	gob.Register(map[string]interface{}{})
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, relying on environment variables")
	}

	cfg := config.Load()

	auth, err := authenticator.New(cfg.Auth0ClientID, cfg.Auth0ClientSecret, cfg.Auth0CallbackURL)
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	e := echo.New()
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("our-secret-key"))))

	router.New(e, auth)

	log.Printf("Server listening on http://localhost:%s/", cfg.AppPort)
	e.Logger.Fatal(e.Start(":" + cfg.AppPort))
}
