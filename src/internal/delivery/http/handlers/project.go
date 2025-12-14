package handlers

import (
	"net/http"

	"pocketeer/internal/app/services"

	"github.com/labstack/echo/v4"
)

type ProjectHandler struct {
	service *services.ProjectService
}

func NewProjectHandler(service *services.ProjectService) *ProjectHandler {
	return &ProjectHandler{service}
}

func (h *ProjectHandler) RegisterRoutes(router *echo.Group, middlewares ...echo.MiddlewareFunc) {
	group := router.Group("/projects", middlewares...)

	group.POST("", h.Create)
	group.PATCH("/:id", h.Update)
	group.GET("", h.GetAll)
	group.GET("/:id", h.Get)
	group.DELETE("/:id", h.Delete)

	// Planning
	group.PATCH("/:id/plan", h.UpdatePlan)
	group.POST("/:id/plan/resources", h.AddPlannedResource)
	group.DELETE("/:id/plan/resources/:specId", h.RemovePlannedResource)

	// Transitions
	group.POST("/:id/start", h.Start)       // Planning -> Active
	group.POST("/:id/cancel", h.Cancel)     // Active -> Canceled
	group.POST("/:id/complete", h.Complete) // Active -> Completed

	// Tracking
	group.PATCH("/:id/actual", h.UpdateActualMetrics)
	group.POST("/:id/resources", h.AddActiveResource)
	group.PATCH("/:id/resources/:resId", h.UpdateResourceUsage)

	// Manual Reservation Management
	group.POST("/:id/resources/:specId/reservations", h.AddReservation)
	group.PATCH("/:id/resources/:specId/reservations/:resId", h.UpdateReservation)
	group.DELETE("/:id/resources/:specId/reservations/:resId", h.DeleteReservation)
}

// Create
// @Summary      Create new project plan
// @Description  Create a new project which is set to 'Planning' state
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        project  body      services.ProjectCreateRequest  true  "Project Create Request"
// @Success      201      {object}  services.Project
// @Failure      400      {object}  echo.HTTPError
// @Failure      401      {object}  echo.HTTPError
// @Failure      500      {object}  echo.HTTPError
// @Router       /projects [post]
// @Security     BearerAuth
func (h *ProjectHandler) Create(c echo.Context) error {
	var payload services.ProjectCreateRequest
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

// Update
// @Summary      Update project details
// @Description  Update project name or description
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        id    path      int                            true  "Project ID"
// @Param        body  body      services.ProjectUpdateRequest  true  "Update Request"
// @Success      200   {object}  services.Project
// @Router       /projects/{id} [patch]
// @Security     BearerAuth
func (h *ProjectHandler) Update(c echo.Context) error {
	id, err := GetIDParam(c)
	if err != nil {
		return err
	}

	var payload services.ProjectUpdateRequest
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

// GetAll
// @Summary      Search and filter projects
// @Description  Get a list of projects with optional filtering and sorting
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        query         query     string  false  "Search query"
// @Param        state         query     string  false  "Filter by state"
// @Success      200           {array}   services.Project
// @Router       /projects [get]
// @Security     BearerAuth
func (h *ProjectHandler) GetAll(c echo.Context) error {
	req := services.ProjectSearchRequest{
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

// Get
// @Summary      Get project details
// @Tags         projects
// @Param        id   path      int  true  "Project ID"
// @Success      200  {object}  services.ProjectFull
// @Router       /projects/{id} [get]
// @Security     BearerAuth
func (h *ProjectHandler) Get(c echo.Context) error {
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

// Delete
// @Summary      Delete project
// @Tags         projects
// @Param        id   path      int  true  "Project ID"
// @Success      204  {object}  nil
// @Router       /projects/{id} [delete]
// @Security     BearerAuth
func (h *ProjectHandler) Delete(c echo.Context) error {
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

// UpdatePlan
// @Summary      Update project plan
// @Tags         projects
// @Param        id    path      int                                true  "Project ID"
// @Param        body  body      services.ProjectPlanUpdateRequest  true  "Plan Update Request"
// @Success      200   {object}  services.Project
// @Router       /projects/{id}/plan [put]
// @Security     BearerAuth
func (h *ProjectHandler) UpdatePlan(c echo.Context) error {
	id, err := GetIDParam(c)
	if err != nil {
		return err
	}

	var payload services.ProjectPlanUpdateRequest
	if err := ParseAndValidatePayload(c, &payload); err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.UpdatePlan(c.Request().Context(), auth0ID, id, payload)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

// AddPlannedResource
// @Summary      Add planned resource
// @Tags         projects
// @Param        id    path      int                                 true  "Project ID"
// @Param        body  body      services.AddPlannedResourceRequest  true  "Resource Spec Request"
// @Success      200   {object}  services.ResourceSpecification
// @Router       /projects/{id}/plan/resources [post]
// @Security     BearerAuth
func (h *ProjectHandler) AddPlannedResource(c echo.Context) error {
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
// @Summary      Remove planned resource
// @Tags         projects
// @Param        id      path      int  true  "Project ID"
// @Param        specId  path      int  true  "Specification ID"
// @Success      204     {object}  nil
// @Router       /projects/{id}/plan/resources/{specId} [delete]
// @Security     BearerAuth
func (h *ProjectHandler) RemovePlannedResource(c echo.Context) error {
	projectID, err := GetIDParam(c)
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

	_, err = h.service.RemovePlannedResource(c.Request().Context(), auth0ID, projectID, specID)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// Start
// @Summary      Start project
// @Tags         projects
// @Param        id   path      int  true  "Project ID"
// @Success      200  {object}  services.ProjectFull
// @Router       /projects/{id}/start [post]
// @Security     BearerAuth
func (h *ProjectHandler) Start(c echo.Context) error {
	id, err := GetIDParam(c)
	if err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.Start(c.Request().Context(), auth0ID, id)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

// Cancel
// @Summary      Cancel project
// @Tags         projects
// @Param        id    path      int                            true  "Project ID"
// @Param        body  body      services.CancelProjectRequest  true  "Cancel Request"
// @Success      200   {object}  services.Project
// @Router       /projects/{id}/cancel [post]
// @Security     BearerAuth
func (h *ProjectHandler) Cancel(c echo.Context) error {
	id, err := GetIDParam(c)
	if err != nil {
		return err
	}

	var payload services.CancelProjectRequest
	if err := ParseAndValidatePayload(c, &payload); err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.Cancel(c.Request().Context(), auth0ID, id, payload)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

// Complete
// @Summary      Complete project
// @Tags         projects
// @Param        id    path      int                              true  "Project ID"
// @Param        body  body      services.CompleteProjectRequest  true  "Complete Request"
// @Success      200   {object}  services.Project
// @Router       /projects/{id}/complete [post]
// @Security     BearerAuth
func (h *ProjectHandler) Complete(c echo.Context) error {
	id, err := GetIDParam(c)
	if err != nil {
		return err
	}

	var payload services.CompleteProjectRequest
	if err := ParseAndValidatePayload(c, &payload); err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.Complete(c.Request().Context(), auth0ID, id, payload)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

// UpdateActualMetrics
// @Summary      Update actual metrics
// @Tags         projects
// @Param        id    path      int                                   true  "Project ID"
// @Param        body  body      services.ProjectActualMetricsRequest  true  "Metrics Update Request"
// @Success      200   {object}  services.Project
// @Router       /projects/{id}/actual [patch]
// @Security     BearerAuth
func (h *ProjectHandler) UpdateActualMetrics(c echo.Context) error {
	id, err := GetIDParam(c)
	if err != nil {
		return err
	}

	var payload services.ProjectActualMetricsRequest
	if err := ParseAndValidatePayload(c, &payload); err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.UpdateActualMetrics(c.Request().Context(), auth0ID, id, payload)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

// AddActiveResource
// @Summary      Add active resource
// @Tags         projects
// @Param        id    path      int                                true  "Project ID"
// @Param        body  body      services.AddActiveResourceRequest  true  "Active Resource Request"
// @Success      200   {object}  services.ProjectFull
// @Router       /projects/{id}/resources [post]
// @Security     BearerAuth
func (h *ProjectHandler) AddActiveResource(c echo.Context) error {
	id, err := GetIDParam(c)
	if err != nil {
		return err
	}

	var payload services.AddActiveResourceRequest
	if err := ParseAndValidatePayload(c, &payload); err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.AddActiveResource(c.Request().Context(), auth0ID, id, payload)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

// UpdateResourceUsage
// @Summary      Update resource usage
// @Tags         projects
// @Param        id     path      int                                  true  "Project ID"
// @Param        resId  path      int                                  true  "Reservation ID"
// @Param        body   body      services.UpdateResourceUsageRequest  true  "Usage Update Request"
// @Success      200    {object}  services.ResourceReservation
// @Router       /projects/{id}/resources/{resId} [patch]
// @Security     BearerAuth
func (h *ProjectHandler) UpdateResourceUsage(c echo.Context) error {
	projectID, err := GetIDParam(c)
	if err != nil {
		return err
	}

	resID, err := GetIDParamWithName(c, "resId")
	if err != nil {
		return err
	}

	var payload services.UpdateResourceUsageRequest
	if err := ParseAndValidatePayload(c, &payload); err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.UpdateResourceUsage(c.Request().Context(), auth0ID, projectID, resID, payload)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

// AddReservation
// @Summary      Add manual reservation
// @Tags         projects
// @Param        id      path      int                             true  "Project ID"
// @Param        specId  path      int                             true  "Resource Specification ID"
// @Param        body    body      services.AddReservationRequest  true  "Reservation Details"
// @Success      201     {object}  services.ResourceReservation
// @Router       /projects/{id}/resources/{specId}/reservations [post]
// @Security     BearerAuth
func (h *ProjectHandler) AddReservation(c echo.Context) error {
	projectID, err := GetIDParam(c)
	if err != nil {
		return err
	}

	specID, err := GetIDParamWithName(c, "specId")
	if err != nil {
		return err
	}

	var payload services.AddReservationRequest
	if err := ParseAndValidatePayload(c, &payload); err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.AddReservation(
		c.Request().Context(),
		auth0ID,
		projectID,
		specID,
		payload,
	)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusCreated, result)
}

// UpdateReservation
// @Summary      Update manual reservation
// @Tags         projects
// @Param        id      path      int                                true  "Project ID"
// @Param        specId  path      int                                true  "Resource Specification ID"
// @Param        resId   path      int                                true  "Reservation ID"
// @Param        body    body      services.UpdateReservationRequest  true  "Update Details"
// @Success      200     {object}  services.ResourceReservation
// @Router       /projects/{id}/resources/{specId}/reservations/{resId} [patch]
// @Security     BearerAuth
func (h *ProjectHandler) UpdateReservation(c echo.Context) error {
	projectID, err := GetIDParam(c)
	if err != nil {
		return err
	}

	specID, err := GetIDParamWithName(c, "specId")
	if err != nil {
		return err
	}

	resID, err := GetIDParamWithName(c, "resId")
	if err != nil {
		return err
	}

	var payload services.UpdateReservationRequest
	if err := ParseAndValidatePayload(c, &payload); err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	result, err := h.service.UpdateReservation(
		c.Request().Context(),
		auth0ID,
		projectID,
		specID,
		resID,
		payload,
	)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.JSON(http.StatusOK, result)
}

// DeleteReservation
// @Summary      Delete manual reservation
// @Tags         projects
// @Param        id      path      int  true  "Project ID"
// @Param        specId  path      int  true  "Resource Specification ID"
// @Param        resId   path      int  true  "Reservation ID"
// @Success      204     {object}  nil
// @Router       /projects/{id}/resources/{specId}/reservations/{resId} [delete]
// @Security     BearerAuth
func (h *ProjectHandler) DeleteReservation(c echo.Context) error {
	projectID, err := GetIDParam(c)
	if err != nil {
		return err
	}

	specID, err := GetIDParamWithName(c, "specId")
	if err != nil {
		return err
	}

	resID, err := GetIDParamWithName(c, "resId")
	if err != nil {
		return err
	}

	auth0ID, err := GetAuth0ID(c)
	if err != nil {
		return err
	}

	err = h.service.DeleteReservation(
		c.Request().Context(),
		auth0ID,
		projectID,
		specID,
		resID,
	)
	if err != nil {
		return ServiceErrToHttp(err)
	}

	return c.NoContent(http.StatusNoContent)
}
