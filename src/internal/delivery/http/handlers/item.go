package handlers

import (
	"net/http"
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
	group.POST("/create", h.CreateItem)
	group.DELETE("/delete", h.DeleteItem)
	group.PATCH("/update", h.UpdateItem)
	group.GET("/list", h.ListItems)
}

func (h *ItemHandler) CreateItem(c echo.Context) error {
	return echo.NewHTTPError(
		http.StatusNotImplemented, "Method CreateItem is not yet implemented.")
}

func (h *ItemHandler) UpdateItem(c echo.Context) error {
	return echo.NewHTTPError(
		http.StatusNotImplemented, "Method UpdateItem is not yet implemented.")
}

func (h *ItemHandler) DeleteItem(c echo.Context) error {
	return echo.NewHTTPError(
		http.StatusNotImplemented, "Method DeleteItem is not yet implemented.")
}

func (h *ItemHandler) ListItems(c echo.Context) error {
	return echo.NewHTTPError(
		http.StatusNotImplemented, "Method ListItems is not yet implemented.")
}
