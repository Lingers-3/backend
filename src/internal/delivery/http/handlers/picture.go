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
	group.GET("/:id", h.GetFile)
	group.GET("/:id/info", h.GetInfo)
	group.DELETE("/:id", h.Delete)
}

// Upload uploads a new picture
// @Summary      Upload picture
// @Description  Upload a new image file (max 10MB). Supported formats: JPEG, PNG, WebP, GIF
// @Tags         pictures
// @Accept       multipart/form-data
// @Produce      json
// @Param        image  formData  file  true  "Image file to upload"
// @Success      201    {object}  services.PictureInfo
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

// GetFile retrieves the picture file content
// @Summary      Get picture file
// @Description  Download the actual image file content
// @Tags         pictures
// @Produce      image/jpeg
// @Produce      image/png
// @Produce      image/webp
// @Produce      image/gif
// @Param        id   path  int  true  "Picture ID"
// @Success      200  {file}  binary  "Image file"
// @Failure      400  {object}  echo.HTTPError
// @Failure      401  {object}  echo.HTTPError
// @Failure      404  {object}  echo.HTTPError
// @Failure      500  {object}  echo.HTTPError
// @Router       /pictures/{id} [get]
// @Security     BearerAuth
func (h *PictureHandler) GetFile(c echo.Context) error {
	ID, err := GetIDParam(c)
	if err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	pictureFile, err := h.service.GetFile(c.Request().Context(), auth0ID, ID)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	c.Response().Header().Set("Content-Type", pictureFile.MimeType)
	c.Response().Header().Set("Content-Disposition", "inline; filename=\""+pictureFile.Filename+"\"")

	return c.Blob(http.StatusOK, pictureFile.MimeType, pictureFile.Content)
}

// GetInfo retrieves picture metadata
// @Summary      Get picture metadata
// @Description  Retrieve metadata information about a picture without downloading the file
// @Tags         pictures
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Picture ID"
// @Success      200  {object}  services.PictureInfo
// @Failure      400  {object}  echo.HTTPError
// @Failure      401  {object}  echo.HTTPError
// @Failure      404  {object}  echo.HTTPError
// @Failure      500  {object}  echo.HTTPError
// @Router       /pictures/{id}/info [get]
// @Security     BearerAuth
func (h *PictureHandler) GetInfo(c echo.Context) error {
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
