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

	group.GET("/full", h.GetAllFull)
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

// Get retrieves a specific item by ID
// @Summary      Get item by ID
// @Description  Retrieve detailed information about a specific item
// @Tags         items
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Item ID"
// @Success      200  {object}  services.Item
// @Failure      400  {object}  echo.HTTPError
// @Failure      401  {object}  echo.HTTPError
// @Failure      404  {object}  echo.HTTPError
// @Failure      500  {object}  echo.HTTPError
// @Router       /items/{id} [get]
// @Security     BearerAuth
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

// GetAll retrieves all items for the authenticated user
// @Summary      List all items
// @Description  Get a list of all items belonging to the authenticated user
// @Tags         items
// @Accept       json
// @Produce      json
// @Success      200  {array}   services.Item
// @Failure      401  {object}  echo.HTTPError
// @Failure      500  {object}  echo.HTTPError
// @Router       /items [get]
// @Security     BearerAuth
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

// Update modifies an existing item
// @Summary      Update item
// @Description  Update properties of an existing item
// @Tags         items
// @Accept       json
// @Produce      json
// @Param        id    path      int                         true  "Item ID"
// @Param        item  body      services.ItemUpdateRequest  true  "Item Update Request"
// @Success      200   {object}  services.Item
// @Failure      400   {object}  echo.HTTPError
// @Failure      401   {object}  echo.HTTPError
// @Failure      404   {object}  echo.HTTPError
// @Failure      500   {object}  echo.HTTPError
// @Router       /items/{id} [patch]
// @Security     BearerAuth
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

// Delete removes an item
// @Summary      Delete item
// @Description  Soft or hard delete an item. Use ?force=true for hard delete
// @Tags         items
// @Accept       json
// @Produce      json
// @Param        id     path      int     true   "Item ID"
// @Param        force  query     boolean false  "Force hard delete"
// @Success      200    {object}  ItemDeleteResponse
// @Failure      400    {object}  echo.HTTPError
// @Failure      401    {object}  echo.HTTPError
// @Failure      404    {object}  echo.HTTPError
// @Failure      500    {object}  echo.HTTPError
// @Router       /items/{id} [delete]
// @Security     BearerAuth
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

// GetAllFull retrieves items with full details
// @Summary      List items with full details
// @Description  Get a list of items with nested tags, optionally filtered by item type
// @Tags         items
// @Accept       json
// @Produce      json
// @Param        item_type_id  query     int  false  "Filter by Item Type ID"
// @Success      200           {array}   services.ItemFull
// @Failure      400           {object}  echo.HTTPError
// @Failure      401           {object}  echo.HTTPError
// @Failure      500           {object}  echo.HTTPError
// @Router       /items/full [get]
// @Security     BearerAuth
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
