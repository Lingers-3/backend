package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"pocketeer/internal/config"
	"pocketeer/internal/platform/authenticator"
	"pocketeer/internal/platform/db"
	"pocketeer/internal/platform/router"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	db, err := db.Init(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize the database connection: %v", err)
	}

	// Close the database on program exit
	defer func() {
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("Failed to get a database object: %v", err)
			return
		}

		err = sqlDB.Close()
		if err != nil {
			log.Printf("Failed to close the database: %v", err)
			return
		}

		log.Println("Database closed successfully")
	}()

	auth, err := authenticator.New(cfg.Auth0Domain, cfg.Auth0ClientID, cfg.Auth0ClientSecret, cfg.Auth0CallbackURL)
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	e := echo.New()

	// TODO(pencelheimer): move it to the separate function?
	allowedOrigins := []string{"https://pocketeer.linerds.us", "http://localhost:5173", "pocketeer-dev.vercel.app"}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPut,
			http.MethodPatch,
			http.MethodPost,
			http.MethodDelete,
		},
		AllowCredentials: true,
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
		},
	}))

	e.Use(session.Middleware(sessions.NewCookieStore([]byte(cfg.SessionSecret))))

	router.New(e, auth, db, cfg)

	socket := fmt.Sprintf("%s:%s", cfg.AppAddress, cfg.AppPort)
	log.Printf("Server listening on http://%s/", socket)
	e.Logger.Fatal(e.Start(socket))
}
