package handlers

import (
	"log"
	"net/http"
	"pocketeer/internal/app/services"
	"pocketeer/internal/delivery/http/middleware"

	"github.com/labstack/echo/v4"
)

type TagHandler struct {
	service *services.TagService
}

func NewTagHandler(service *services.TagService) *TagHandler {
	return &TagHandler{service}
}

func (h *TagHandler) RegisterRoutes(router *echo.Group, middlewares ...echo.MiddlewareFunc) {
	group := router.Group("tags", middlewares...)
	group.POST("", h.Create)
	group.GET("", h.GetAll)
	group.GET("/:id", h.Get)
	group.PATCH("/:id", h.Update)
	group.DELETE("/:id", h.Delete)
}

func (h *TagHandler) Create(c echo.Context) error {
	auth0ID, ok := middleware.GetAuth0IDFromRequest(c)
	if !ok || auth0ID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	var payload services.CreateTagRequest
	if err := c.Bind(&payload); err != nil {
		log.Printf("Failed parsing payload: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	tag, err := h.service.Create(c.Request().Context(), auth0ID, payload)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusCreated, tag)
}

func (h *TagHandler) GetAll(c echo.Context) error {
	auth0ID, ok := middleware.GetAuth0IDFromRequest(c)
	if !ok || auth0ID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	tags, err := h.service.GetAll(c.Request().Context(), auth0ID)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, tags)
}

func (h *TagHandler) Get(c echo.Context) error {
	auth0ID, ok := middleware.GetAuth0IDFromRequest(c)
	if !ok || auth0ID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	tagID, err := GetIDParam(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tag ID")
	}

	tag, err := h.service.Get(c.Request().Context(), auth0ID, tagID)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, tag)
}

func (h *TagHandler) Update(c echo.Context) error {
	auth0ID, ok := middleware.GetAuth0IDFromRequest(c)
	if !ok || auth0ID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	tagID, err := GetIDParam(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tag ID")
	}

	var payload services.UpdateTagRequest
	if err := c.Bind(&payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid payload")
	}

	tag, err := h.service.Update(c.Request().Context(), auth0ID, tagID, payload)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, tag)
}

func (h *TagHandler) Delete(c echo.Context) error {
	auth0ID, ok := middleware.GetAuth0IDFromRequest(c)
	if !ok || auth0ID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	tagID, err := GetIDParam(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid tag ID")
	}

	err = h.service.Delete(c.Request().Context(), auth0ID, tagID)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.NoContent(http.StatusNoContent)
}
