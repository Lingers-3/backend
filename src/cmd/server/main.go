// @title           Pocketeer API
// @version         1.0
// @description     Inventory management system API

// @host      localhost:3000
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

	_ "pocketeer/docs"
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
	"github.com/swaggo/echo-swagger"
)

func init() {
	gob.Register(map[string]any{})
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i any) error {
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
	itemTypeService := services.NewItemTypeService(db, userService)
	itemService := services.NewItemService(db, userService)
	tagService := services.NewTagService(db, userService)
	pictureService := services.NewPictureService(db, userService)
	projectService := services.NewProjectService(db, userService, itemService)
	templateService := services.NewTemplateService(db, userService, projectService)

	userHandler := handlers.NewUserHandler(userService, cfg)
	itemTypeHandler := handlers.NewItemTypeHandler(itemTypeService)
	itemHandler := handlers.NewItemHandler(itemService)
	tagHandler := handlers.NewTagHandler(tagService)
	pictureHandler := handlers.NewPictureHandler(pictureService)
	projectHandler := handlers.NewProjectHandler(projectService)
	templateHandler := handlers.NewTemplateHandler(templateService)

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
	projectHandler.RegisterRoutes(api, authMiddleware)
	templateHandler.RegisterRoutes(api, authMiddleware)

	// TODO(noatu): move to config
	e.Static("/pictures", "/var/pocketeer/img")
	e.Static("/api/pictures/static", "/var/pocketeer/img")
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", cfg.AppAddress, cfg.AppPort)))
}
