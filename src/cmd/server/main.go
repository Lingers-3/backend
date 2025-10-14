package main

import (
	"encoding/gob"
	"log"
	"pocketeer/internal/config"
	"pocketeer/internal/platform/authenticator"
	"pocketeer/internal/platform/db"
	"pocketeer/internal/platform/router"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

// HACK(pencelheimer): the fuck is this?
func init() {
	gob.Register(map[string]any{})
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, relying on environment variables")
	}

	cfg := config.Load()

	// TODO(pencelheimer): make db.Init return an error, or make authenticator.New panic
	db := db.Init(cfg)

	auth, err := authenticator.New(cfg.Auth0Domain, cfg.Auth0ClientID, cfg.Auth0ClientSecret, cfg.Auth0CallbackURL)
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	e := echo.New()
	e.Use(session.Middleware(sessions.NewCookieStore([]byte("our-secret-key")))) // HACK(pencelheimer): add a secret definition via config?

	router.New(e, auth, db, cfg)

	// HACK(pencelheimer): add an interface/address definition via config?
	log.Printf("Server listening on http://localhost:%s/", cfg.AppPort)
	e.Logger.Fatal(e.Start(":" + cfg.AppPort))
}
