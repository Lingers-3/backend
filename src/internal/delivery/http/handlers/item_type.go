package handlers

import (
	"net/http"

	"pocketeer/internal/app/services"

	"github.com/labstack/echo/v4"
)

type ItemTypeHandler struct {
	service *services.ItemTypeService
}

func NewItemTypeHandler(service *services.ItemTypeService) *ItemTypeHandler {
	return &ItemTypeHandler{service}
}

func (h *ItemTypeHandler) RegisterRoutes(router *echo.Group, middlewares ...echo.MiddlewareFunc) {
	group := router.Group("item-types", middlewares...)
	group.POST("", h.Create)
	group.GET("", h.GetAll)
	group.GET("/:id", h.Get)
	group.PATCH("/:id", h.Update)
	group.DELETE("/:id", h.Delete)
}

func (h *ItemTypeHandler) Create(c echo.Context) error {
	var payload services.ItemTypeCreateRequest
	err := ParsePayload(c, &payload)
	if err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.Create(c.Request().Context(), auth0ID, payload)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *ItemTypeHandler) Get(c echo.Context) error {
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

func (h *ItemTypeHandler) GetAll(c echo.Context) error {
	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.GetAll(c.Request().Context(), auth0ID)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

func (h *ItemTypeHandler) Update(c echo.Context) error {
	ID, err := GetIDParam(c)
	if err != nil {
		return err
	}

	var payload services.ItemTypeUpdateRequest
	err = ParsePayload(c, &payload)
	if err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.Update(c.Request().Context(), auth0ID, ID, payload)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

type ItemTypeDeleteResponse struct {
	Hard bool `json:"hard"` // false = soft delete
}

func (h *ItemTypeHandler) Delete(c echo.Context) error {
	ID, err := GetIDParam(c)
	if err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	hard, err := h.service.Delete(c.Request().Context(), auth0ID, ID)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, ItemTypeDeleteResponse{hard})
}
