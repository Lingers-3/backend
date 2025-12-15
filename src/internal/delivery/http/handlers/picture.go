package handlers

import (
	"net/http"

	"pocketeer/internal/app/services"

	"github.com/labstack/echo/v4"
)

type PictureHandler struct {
	service *services.PictureService
}

func NewPictureHandler(service *services.PictureService) *PictureHandler {
	return &PictureHandler{service}
}

func (h *PictureHandler) RegisterRoutes(router *echo.Group, middlewares ...echo.MiddlewareFunc) {
	group := router.Group("/pictures", middlewares...)
	group.POST("", h.Upload)
	group.GET("/:id", h.Get)
	group.DELETE("/:id", h.Delete)
}

// Represents the static file serving endpoint.
// @Summary      Get a static picture
// @Description  Serves an image file from the static directory
// @Tags         pictures
// @Produce      image/png, image/jpeg, image/webp, image/gif
// @Param        hash path string true "Image Hash (Filename)"
// @Success      200        {file}    file
// @Failure      404        {string}  string  "Not Found"
// @Router       /pictures/static/{hash} [get]
func ServeStaticPictures() {
	// NOTE(noatu): This function exists solely for Swagger documentation.
	// The actual serving is handled by e.Static("/pictures", ...) in main.go
}

// Upload uploads a new picture
// @Summary      Upload picture
// @Description  Upload a new image file (max 10MB). Supported formats: JPEG, PNG, WebP, GIF
// @Tags         pictures
// @Accept       multipart/form-data
// @Produce      json
// @Param        image  formData  file  true  "Image file to upload"
// @Success      201    {object}  services.Picture
// @Failure      400    {object}  echo.HTTPError  "Invalid file or image file required"
// @Failure      401    {object}  echo.HTTPError
// @Failure      413    {object}  echo.HTTPError  "Image too large (max 10MB)"
// @Failure      415    {object}  echo.HTTPError  "Invalid image format"
// @Failure      500    {object}  echo.HTTPError
// @Router       /pictures [post]
// @Security     BearerAuth
func (h *PictureHandler) Upload(c echo.Context) error {
	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	fileHeader, err := c.FormFile("image")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "image file required")
	}

	result, err := h.service.Upload(c.Request().Context(), auth0ID, fileHeader)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusCreated, result)
}

// Get retrieves picture metadata
// @Summary      Get picture metadata
// @Description  Retrieve metadata information about a picture without downloading the file
// @Tags         pictures
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Picture ID"
// @Success      200  {object}  services.Picture
// @Failure      400  {object}  echo.HTTPError
// @Failure      401  {object}  echo.HTTPError
// @Failure      404  {object}  echo.HTTPError
// @Failure      500  {object}  echo.HTTPError
// @Router       /pictures/{id} [get]
// @Security     BearerAuth
func (h *PictureHandler) Get(c echo.Context) error {
	ID, err := GetIDParam(c)
	if err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.Get(c.Request().Context(), auth0ID, ID)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

// Delete removes a picture
// @Summary      Delete picture
// @Description  Permanently delete a picture and its file if not referenced by other users
// @Tags         pictures
// @Accept       json
// @Produce      json
// @Param        id   path  int  true  "Picture ID"
// @Success      200  "Picture deleted successfully"
// @Failure      400  {object}  echo.HTTPError
// @Failure      401  {object}  echo.HTTPError
// @Failure      404  {object}  echo.HTTPError
// @Failure      409  {object}  echo.HTTPError  "Picture is still in use"
// @Failure      500  {object}  echo.HTTPError
// @Router       /pictures/{id} [delete]
// @Security     BearerAuth
func (h *PictureHandler) Delete(c echo.Context) error {
	ID, err := GetIDParam(c)
	if err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	err = h.service.Delete(c.Request().Context(), auth0ID, ID)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.NoContent(http.StatusOK)
}
