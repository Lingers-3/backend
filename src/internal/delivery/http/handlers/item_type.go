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
	group := router.Group("/item-types", middlewares...)
	group.POST("", h.Create)
	group.GET("", h.GetAll)
	group.GET("/:id", h.Get)
	group.PATCH("/:id", h.Update)
	group.DELETE("/:id", h.Delete)

	group.GET("/full", h.GetAllFull)
}

// Create creates a new item type
// @Summary      Create a new item type
// @Description  Create a new item type with specified properties
// @Tags         item-types
// @Accept       json
// @Produce      json
// @Param        item_type  body      services.ItemTypeCreateRequest  true  "Item Type Creation Request"
// @Success      200        {object}  services.ItemType
// @Failure      400        {object}  echo.HTTPError
// @Failure      401        {object}  echo.HTTPError
// @Failure      500        {object}  echo.HTTPError
// @Router       /item-types [post]
// @Security     BearerAuth
func (h *ItemTypeHandler) Create(c echo.Context) error {
	var payload services.ItemTypeCreateRequest
	err := ParseAndValidatePayload(c, &payload)
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

// Get retrieves a specific item type by ID
// @Summary      Get item type by ID
// @Description  Retrieve detailed information about a specific item type
// @Tags         item-types
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Item Type ID"
// @Success      200  {object}  services.ItemType
// @Failure      400  {object}  echo.HTTPError
// @Failure      401  {object}  echo.HTTPError
// @Failure      404  {object}  echo.HTTPError
// @Failure      500  {object}  echo.HTTPError
// @Router       /item-types/{id} [get]
// @Security     BearerAuth
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

// GetAll retrieves all item types for the authenticated user
// @Summary      List all item types
// @Description  Get a list of all item types belonging to the authenticated user
// @Tags         item-types
// @Accept       json
// @Produce      json
// @Success      200  {array}   services.ItemType
// @Failure      401  {object}  echo.HTTPError
// @Failure      500  {object}  echo.HTTPError
// @Router       /item-types [get]
// @Security     BearerAuth
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

// GetAll retrieves all item types for the authenticated user
// @Summary      List all item types
// @Description  Get a list of all item types belonging to the authenticated user
// @Tags         item-types
// @Accept       json
// @Produce      json
// @Success      200  {array}   services.ItemType
// @Failure      401  {object}  echo.HTTPError
// @Failure      500  {object}  echo.HTTPError
// @Router       /item-types [get]
// @Security     BearerAuth
func (h *ItemTypeHandler) Update(c echo.Context) error {
	ID, err := GetIDParam(c)
	if err != nil {
		return err
	}

	var payload services.ItemTypeUpdateRequest
	err = ParseAndValidatePayload(c, &payload)
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

// Delete removes an item type
// @Summary      Delete item type
// @Description  Soft or hard delete an item type. Use ?force=true for hard delete
// @Tags         item-types
// @Accept       json
// @Produce      json
// @Param        id     path      int     true   "Item Type ID"
// @Param        force  query     boolean false  "Force hard delete"
// @Success      200    {object}  ItemTypeDeleteResponse
// @Failure      400    {object}  echo.HTTPError
// @Failure      401    {object}  echo.HTTPError
// @Failure      404    {object}  echo.HTTPError
// @Failure      500    {object}  echo.HTTPError
// @Router       /item-types/{id} [delete]
// @Security     BearerAuth
func (h *ItemTypeHandler) Delete(c echo.Context) error {
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

	return c.JSON(http.StatusOK, ItemTypeDeleteResponse{hard})
}

// GetAllFull retrieves all item types with full details including items and tags
// @Summary      List all item types with full details
// @Description  Get a complete list of all item types with nested items and tags
// @Tags         item-types
// @Accept       json
// @Produce      json
// @Success      200  {array}   services.ItemTypeFull
// @Failure      401  {object}  echo.HTTPError
// @Failure      500  {object}  echo.HTTPError
// @Router       /item-types/full [get]
// @Security     BearerAuth
func (h *ItemTypeHandler) GetAllFull(c echo.Context) error {
	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.GetAllFull(c.Request().Context(), auth0ID)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}
