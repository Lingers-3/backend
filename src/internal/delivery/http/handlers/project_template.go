package handlers

import (
	"net/http"

	"pocketeer/internal/app/services"

	"github.com/labstack/echo/v4"
)

type TemplateHandler struct {
	service *services.TemplateService
}

func NewTemplateHandler(service *services.TemplateService) *TemplateHandler {
	return &TemplateHandler{service}
}

func (h *TemplateHandler) RegisterRoutes(router *echo.Group, middlewares ...echo.MiddlewareFunc) {
	group := router.Group("/templates", middlewares...)

	// CRUD Operations
	group.POST("", h.Create)
	group.PATCH("/:id", h.Update)
	group.GET("", h.GetAll)
	group.GET("/:id", h.Get)
	group.DELETE("/:id", h.Delete)

	// Additional Operations
	group.POST("/from-project/:projectId", h.CreateFromProject)

	// Resource Specification Management
	group.POST("/:id/resources", h.AddPlannedResource)
	group.DELETE("/:id/resources/:specId", h.RemovePlannedResource)
}

// Create
// @Summary      Create new project template
// @Description  Create a new template for future projects
// @Tags         templates
// @Accept       json
// @Produce      json
// @Param        template  body      services.TemplateCreateRequest  true  "Template Create Request"
// @Success      201       {object}  services.Template
// @Failure      400       {object}  echo.HTTPError
// @Failure      401       {object}  echo.HTTPError
// @Failure      500       {object}  echo.HTTPError
// @Router       /templates [post]
// @Security     BearerAuth
func (h *TemplateHandler) Create(c echo.Context) error {
	var payload services.TemplateCreateRequest
	if err := ParseAndValidatePayload(c, &payload); err != nil {
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

	return c.JSON(http.StatusCreated, result)
}

// CreateFromProject
// @Summary      Create template from an existing project
// @Description  Creates a new template by copying plan/actual metrics and resource specs from a project
// @Tags         templates
// @Accept       json
// @Produce      json
// @Param        projectId  path      int                         true  "Project ID"
// @Param        body       body      services.TemplateFromProjectRequest  true  "Template Name Request"
// @Success      201        {object}  services.Template
// @Router       /templates/from-project/{projectId} [post]
// @Security     BearerAuth
func (h *TemplateHandler) CreateFromProject(c echo.Context) error {
	projectID, err := GetIDParamWithName(c, "projectId")
	if err != nil {
		return err
	}

	var payload struct {
		Name string `json:"name" validate:"required"`
	}
	if err := ParseAndValidatePayload(c, &payload); err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.CreateFromProject(
		c.Request().Context(),
		auth0ID,
		projectID,
		payload.Name,
	)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusCreated, result)
}

// Update
// @Summary      Update template details
// @Tags         templates
// @Accept       json
// @Produce      json
// @Param        id    path      int                            true  "Template ID"s
// @Param        body  body      services.TemplateUpdateRequest  true  "Update Request"
// @Success      200   {object}  services.Template
// @Router       /templates/{id} [patch]
// @Security     BearerAuth
func (h *TemplateHandler) Update(c echo.Context) error {
	id, err := GetIDParam(c)
	if err != nil {
		return err
	}

	var payload services.TemplateUpdateRequest
	if err := ParseAndValidatePayload(c, &payload); err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.Update(c.Request().Context(), auth0ID, id, payload)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

// Get
// @Summary      Get template details
// @Tags         templates
// @Param        id   path      int  true  "Template ID"
// @Success      200  {object}  services.TemplateFull
// @Router       /templates/{id} [get]
// @Security     BearerAuth
func (h *TemplateHandler) Get(c echo.Context) error {
	id, err := GetIDParam(c)
	if err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.Get(c.Request().Context(), auth0ID, id)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

// GetAll
// @Summary      Search and filter templates
// @Tags         templates
// @Accept       json
// @Produce      json
// @Param        query         query     string  false  "Search query"
// @Param        sort_by       query     string  false  "Sort by column (name, usage_count, etc.)"
// @Param        sort_order    query     string  false  "Sort order (asc or desc)"
// @Param        page          query     int     false  "Page number (default 1)"
// @Param        page_size     query     int     false  "Items per page (default 20)"
// @Success      200           {array}   services.Template
// @Router       /templates [get]
// @Security     BearerAuth
func (h *TemplateHandler) GetAll(c echo.Context) error {
	req := services.TemplateSearchRequest{
		Page:     1,
		PageSize: 20,
	}

	if err := ParseAndValidatePayload(c, &req); err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.GetAll(c.Request().Context(), auth0ID, req)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

// Delete
// @Summary      Delete template
// @Tags         templates
// @Param        id   path      int  true  "Template ID"
// @Success      204  {object}  nil
// @Router       /templates/{id} [delete]
// @Security     BearerAuth
func (h *TemplateHandler) Delete(c echo.Context) error {
	id, err := GetIDParam(c)
	if err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	_, err = h.service.Delete(c.Request().Context(), auth0ID, id)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// AddPlannedResource
// @Summary      Add planned resource to template
// @Tags         templates
// @Param        id    path      int                                 true  "Template ID"
// @Param        body  body      services.AddPlannedResourceRequest  true  "Resource Spec Request"
// @Success      200   {object}  services.TemplateResourceSpecification
// @Router       /templates/{id}/resources [post]
// @Security     BearerAuth
func (h *TemplateHandler) AddPlannedResource(c echo.Context) error {
	id, err := GetIDParam(c)
	if err != nil {
		return err
	}

	var payload services.AddPlannedResourceRequest
	if err := ParseAndValidatePayload(c, &payload); err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.AddPlannedResource(c.Request().Context(), auth0ID, id, payload)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

// RemovePlannedResource
// @Summary      Remove planned resource from template
// @Tags         templates
// @Param        id      path      int  true  "Template ID"
// @Param        specId  path      int  true  "Specification ID"
// @Success      204     {object}  nil
// @Router       /templates/{id}/resources/{specId} [delete]
// @Security     BearerAuth
func (h *TemplateHandler) RemovePlannedResource(c echo.Context) error {
	templateID, err := GetIDParam(c)
	if err != nil {
		return err
	}

	specID, err := GetIDParamWithName(c, "specId")
	if err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	_, err = h.service.RemovePlannedResource(c.Request().Context(), auth0ID, templateID, specID)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.NoContent(http.StatusNoContent)
}
