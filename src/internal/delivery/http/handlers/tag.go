package handlers

import (
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
	group := router.Group("/tags", middlewares...)
	group.POST("", h.Create)
	group.GET("", h.GetAll)
	group.GET("/:id", h.Get)
	group.PATCH("/:id", h.Update)
	group.DELETE("/:id", h.Delete)
}

// Create creates a new tag
// @Summary      Create a new tag
// @Description  Create a new tag and optionally associate it with an item or item type
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        tag  body      services.CreateTagRequest  true  "Tag Creation Request"
// @Success      201  {object}  services.TagFull
// @Failure      400  {object}  echo.HTTPError
// @Failure      401  {object}  echo.HTTPError
// @Failure      409  {object}  echo.HTTPError  "Tag already exists"
// @Failure      500  {object}  echo.HTTPError
// @Router       /tags [post]
// @Security     BearerAuth
func (h *TagHandler) Create(c echo.Context) error {
	auth0ID, ok := middleware.GetAuth0IDFromRequest(c)
	if !ok || auth0ID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	var payload services.CreateTagRequest
	err := ParseAndValidatePayload(c, &payload)
	if err != nil {
		return err
	}

	tag, err := h.service.Create(c.Request().Context(), auth0ID, payload)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusCreated, tag)
}

// GetAll retrieves all tags for the authenticated user
// @Summary      List all tags
// @Description  Get a list of all tags belonging to the authenticated user
// @Tags         tags
// @Accept       json
// @Produce      json
// @Success      200  {array}   services.TagFull
// @Failure      401  {object}  echo.HTTPError
// @Failure      500  {object}  echo.HTTPError
// @Router       /tags [get]
// @Security     BearerAuth
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

// Get retrieves a specific tag by ID
// @Summary      Get tag by ID
// @Description  Retrieve detailed information about a specific tag including associated items and item types
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Tag ID"
// @Success      200  {object}  services.TagFull
// @Failure      400  {object}  echo.HTTPError
// @Failure      401  {object}  echo.HTTPError
// @Failure      404  {object}  echo.HTTPError
// @Failure      500  {object}  echo.HTTPError
// @Router       /tags/{id} [get]
// @Security     BearerAuth
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

// Update modifies an existing tag
// @Summary      Update tag
// @Description  Update properties of an existing tag
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        id   path      int                         true  "Tag ID"
// @Param        tag  body      services.UpdateTagRequest   true  "Tag Update Request"
// @Success      200  {object}  services.TagFull
// @Failure      400  {object}  echo.HTTPError
// @Failure      401  {object}  echo.HTTPError
// @Failure      404  {object}  echo.HTTPError
// @Failure      409  {object}  echo.HTTPError  "Tag already exists"
// @Failure      500  {object}  echo.HTTPError
// @Router       /tags/{id} [patch]
// @Security     BearerAuth
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
	err = ParseAndValidatePayload(c, &payload)
	if err != nil {
		return err
	}

	tag, err := h.service.Update(c.Request().Context(), auth0ID, tagID, payload)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, tag)
}

// Delete removes a tag
// @Summary      Delete tag
// @Description  Permanently delete a tag (hard delete only)
// @Tags         tags
// @Accept       json
// @Produce      json
// @Param        id   path  int  true  "Tag ID"
// @Success      204  "No Content"
// @Failure      400  {object}  echo.HTTPError
// @Failure      401  {object}  echo.HTTPError
// @Failure      404  {object}  echo.HTTPError
// @Failure      500  {object}  echo.HTTPError
// @Router       /tags/{id} [delete]
// @Security     BearerAuth
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
