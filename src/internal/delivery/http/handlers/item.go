package handlers

import (
	"net/http"
	"strconv"

	"pocketeer/internal/app/services"

	"github.com/labstack/echo/v4"
)

type ItemHandler struct {
	service *services.ItemService
}

func NewItemHandler(service *services.ItemService) *ItemHandler {
	return &ItemHandler{service}
}

func (h *ItemHandler) RegisterRoutes(router *echo.Group, middlewares ...echo.MiddlewareFunc) {
	group := router.Group("/items", middlewares...)
	group.POST("", h.Create)
	group.GET("", h.GetAll)
	group.GET("/:id", h.Get)
	group.PATCH("/:id", h.Update)
	group.DELETE("/:id", h.Delete)
}

func (h *ItemHandler) Create(c echo.Context) error {
	var payload services.ItemCreateRequest
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

func (h *ItemHandler) Get(c echo.Context) error {
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

func (h *ItemHandler) GetAll(c echo.Context) error {
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

func (h *ItemHandler) Update(c echo.Context) error {
	ID, err := GetIDParam(c)
	if err != nil {
		return err
	}

	var payload services.ItemUpdateRequest
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

type ItemDeleteResponse struct {
	Hard bool `json:"hard"` // false = soft delete
}

func (h *ItemHandler) Delete(c echo.Context) error {
	ID, err := GetIDParam(c)
	if err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	forceParam := c.QueryParam("force")
	isHardDelete := (forceParam == "true")

	hard, err := h.service.Delete(c.Request().Context(), auth0ID, ID, isHardDelete)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, ItemDeleteResponse{hard})
}

func (h *ItemHandler) GetAllFull(c echo.Context) error {
	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	var itemTypeID *uint
	if param := c.QueryParam("item_type_id"); param != "" {
		id, err := strconv.ParseUint(param, 10, 32)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid item_type_id format")
		}
		uID := uint(id)
		itemTypeID = &uID
	}

	result, err := h.service.GetAllFull(c.Request().Context(), auth0ID, itemTypeID)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}
