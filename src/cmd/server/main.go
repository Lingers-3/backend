package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"pocketeer/internal/config"
	"pocketeer/internal/platform/authenticator"
	"pocketeer/internal/platform/database"
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

	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize the database connection: %v", err)
	}

	// Close the database on program exit
	defer func() {
		// TODO(noatu): our code organization, errors and logging are 🍑
		log.Println("INFO: closing the database")
		if err := database.Close(db); err != nil {
			log.Printf("ERROR: %v", err)
			return
		}
		log.Println("INFO: database closed")
	}()

	auth, err := authenticator.New(cfg.Auth0Domain, cfg.Auth0ClientID, cfg.Auth0ClientSecret, cfg.Auth0CallbackURL)
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	e := echo.New()
	e.Use(session.Middleware(sessions.NewCookieStore([]byte(cfg.SessionSecret))))

	// TODO(noatu): this may as well be inlined
	// but it would be even better to move the echo (and auth) stuff in the router
	router.New(e, auth, db, cfg)

	socket := fmt.Sprintf("%s:%s", cfg.AppAddress, cfg.AppPort)
	log.Printf("Server listening on http://%s/", socket)
	e.Logger.Fatal(e.Start(socket))
}
