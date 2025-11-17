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
