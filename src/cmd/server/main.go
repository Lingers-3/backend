// @title           Pocketeer API
// @version         1.0
// @description     Inventory management system API
// @termsOfService  https://pocketeer.linerds.us/terms

// @contact.name   API Support
// @contact.url    https://pocketeer.linerds.us/support
// @contact.email  support@pocketeer.linerds.us

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      pocketeer-api.linerds.us
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your bearer token in the format: Bearer {token}

// @externalDocs.description  OpenAPI Specification
// @externalDocs.url          https://swagger.io/resources/open-api/
package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"

	"pocketeer/internal/app/services"
	"pocketeer/internal/config"
	"pocketeer/internal/delivery/http/handlers"
	internalMiddleware "pocketeer/internal/delivery/http/middleware"
	"pocketeer/internal/platform/authenticator"
	"pocketeer/internal/platform/database"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func init() {
	gob.Register(map[string]any{})
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
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

	authenticator, err := authenticator.New(
		cfg.Auth0Domain,
		cfg.Auth0ClientID,
		cfg.Auth0ClientSecret,
		cfg.Auth0Audience,
		cfg.Auth0CallbackURL)
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}
	userService := services.NewUserService(db)
	itemTypeHandler := handlers.NewItemTypeHandler(services.NewItemTypeService(db, userService))
	itemHandler := handlers.NewItemHandler(services.NewItemService(db, userService))
	tagHandler := handlers.NewTagHandler(services.NewTagService(db, userService))
	pictureHandler := handlers.NewPictureHandler(services.NewPictureService(db, userService))
	userHandler := handlers.NewUserHandler(userService, cfg)

	e := echo.New()

	e.Debug = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Validator = &CustomValidator{validator: validator.New()}

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		c.Logger().Error(err)
		e.DefaultHTTPErrorHandler(err, c)
	}

	// TODO(pencelheimer): move it to the separate function?
	allowedOrigins := []string{"https://pocketeer.linerds.us", "http://localhost:5173", "https://pocketeer-dev.vercel.app"}
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

	api := e.Group("/api")

	authMiddleware := internalMiddleware.AuthMiddleware(authenticator)
	auth := api.Group("/auth")
	auth.GET("/login", handlers.LoginHandler(authenticator))
	auth.GET("/callback", handlers.CallbackHandler(authenticator))
	auth.GET("/logout", handlers.LogoutHandler(cfg))
	auth.POST("/change-password", handlers.UpdatePasswordHandler(cfg))
	auth.POST("/post-login", handlers.PostLoginHandler(userService), authMiddleware)

	itemTypeHandler.RegisterRoutes(api, authMiddleware)
	itemHandler.RegisterRoutes(api, authMiddleware)
	tagHandler.RegisterRoutes(api, authMiddleware)
	pictureHandler.RegisterRoutes(api, authMiddleware)
	userHandler.RegisterRoutes(api, authMiddleware)

	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", cfg.AppAddress, cfg.AppPort)))
}
