package services

import (
	"context"
	"time"

	"pocketeer/internal/platform/database"
	"pocketeer/internal/platform/database/models"
)

type ProjectService struct {
	db          *database.DB
	userService *UserService
	itemService *ItemService
}

func NewProjectService(db *database.DB, userService *UserService, itemService *ItemService) *ProjectService {
	return &ProjectService{db, userService, itemService}
}

type Project struct {
	ID          uint                `json:"id"`
	Name        string              `json:"name"`
	Description *string             `json:"description"`
	State       models.ProjectState `json:"state"`

	PlannedDeadline *time.Time     `json:"planned_deadline"`
	PlannedIncome   *float32       `json:"planned_income"`
	PlannedWorkTime *time.Duration `json:"planned_work_time"`

	ActualDeadline *time.Time     `json:"actual_deadline"`
	ActualIncome   *float32       `json:"actual_income"`
	ActualWorkTime *time.Duration `json:"actual_work_time"`

	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ResourceSpecificationDTO struct {
	ID              uint                `json:"id"`
	ItemTypeID      uint                `json:"item_type_id"`
	ItemTypeName    string              `json:"item_type_name"`
	ResourceType    models.ResourceType `json:"resource_type"`
	PlannedQuantity float32             `json:"planned_quantity"`
}

type ResourceReservationDTO struct {
	ID               uint    `json:"id"`
	ItemID           uint    `json:"item_id"`
	ItemDescription  *string `json:"item_description"` // NOTE(pencelheimer): For convenience
	ReservedQuantity float32 `json:"reserved_quantity"`
	UsedQuantity     float32 `json:"used_quantity"`
}

type ProjectFull struct {
	Project
	Specifications []ResourceSpecificationDTO `json:"specifications"`
	Reservations   []ResourceReservationDTO   `json:"reservations"`
}

type ProjectCreateRequest struct {
	Name string `json:"name" validate:"required,min=1,max=256"`
}

func (s *ProjectService) Create(ctx context.Context, auth0ID string, req ProjectCreateRequest) (*Project, error) {
	return nil, nil
}

// when State == Planning
type ProjectPlanUpdateRequest struct {
	Description     *string    `json:"description" validate:"omitempty,max=512"`
	PlannedDeadline *time.Time `json:"planned_deadline" validate:"omitempty,gt=now"`
	PlannedIncome   *float32   `json:"planned_income" validate:"omitempty,gte=0"`
	PlannedWorkTime *int64     `json:"planned_work_time" validate:"omitempty,gte=0"`
}

func (s *ProjectService) UpdatePlan(ctx context.Context, auth0ID string, projectID uint, req ProjectPlanUpdateRequest) (*Project, error) {
	return nil, nil
}

// when State == Planning
type AddPlannedResourceRequest struct {
	ItemTypeID      uint    `json:"item_type_id" validate:"required,gt=0"`
	ResourceType    string  `json:"resource_type" validate:"required,oneof=Consumable Instrument"`
	PlannedQuantity float32 `json:"planned_quantity" validate:"required,gt=0"`
}

func (s *ProjectService) AddPlannedResource(ctx context.Context, auth0ID string, projectID uint, req AddPlannedResourceRequest) (*ResourceSpecificationDTO, error) {
	return nil, nil
}

// when State == Planning
func (s *ProjectService) RemovePlannedResource(ctx context.Context, auth0ID string, projectID uint, specID uint) (bool, error) {
	return false, nil
}

// Transition: Planning -> Active. Triggers auto-reservation.
func (s *ProjectService) Start(ctx context.Context, auth0ID string, projectID uint) (*ProjectFull, error) {
	return nil, nil
}

// when State == Active
type ProjectActualMetricsRequest struct {
	ActualDeadline *time.Time `json:"actual_deadline" validate:"omitempty"`
	ActualIncome   *float32   `json:"actual_income" validate:"omitempty,gte=0"`
	AddedWorkTime  *int64     `json:"added_work_time" validate:"omitempty,gte=0"`
}

func (s *ProjectService) UpdateActualMetrics(ctx context.Context, auth0ID string, projectID uint, req ProjectActualMetricsRequest) (*Project, error) {
	return nil, nil
}

// when State == Active. Creates Spec + Reservation
type AddActiveResourceRequest struct {
	ItemTypeID       uint    `json:"item_type_id" validate:"required,gt=0"`
	ResourceType     string  `json:"resource_type" validate:"required,oneof=Consumable Instrument"`
	ReservedQuantity float32 `json:"reserved_quantity" validate:"required,gt=0"`
	UsedQuantity     float32 `json:"used_quantity" validate:"gte=0"`
}

func (s *ProjectService) AddActiveResource(ctx context.Context, auth0ID string, projectID uint, req AddActiveResourceRequest) (*ProjectFull, error) {
	return nil, nil
}

// when State == Active
type UpdateResourceUsageRequest struct {
	UsedQuantity float32 `json:"used_quantity" validate:"required,gte=0"`
}

func (s *ProjectService) UpdateResourceUsage(ctx context.Context, auth0ID string, projectID uint, reservationID uint, req UpdateResourceUsageRequest) (*ResourceReservationDTO, error) {
	return nil, nil
}

// Transition: Active -> Canceled. Releases reservations.
type CancelProjectRequest struct {
	ReturnItemsToInventory bool `json:"return_items_to_inventory"`
}

func (s *ProjectService) Cancel(ctx context.Context, auth0ID string, projectID uint, req CancelProjectRequest) (*Project, error) {
	return nil, nil
}

// Transition: Active -> Completed. Finalizes usage.
type CompleteProjectRequest struct {
	ActualRevenue     *float32 `json:"actual_revenue" validate:"omitempty,gte=0"`
	FinalizeInventory bool     `json:"finalize_inventory"` // NOTE(pencelheimer): If true, remaining reserved but unused items are returned
}

func (s *ProjectService) Complete(ctx context.Context, auth0ID string, projectID uint, req CompleteProjectRequest) (*Project, error) {
	return nil, nil
}

func (s *ProjectService) Get(ctx context.Context, auth0ID string, projectID uint) (*ProjectFull, error) {
	return nil, nil
}

type ProjectSearchRequest struct {
    Query       string `json:"query" query:"query" validate:"omitempty,min=3"`
    State       string `json:"state" query:"state" validate:"omitempty,oneof=Planning Active Completed Canceled"`
    HasDeadline *bool  `json:"has_deadline" query:"has_deadline"`
    SortBy      string `json:"sort_by" query:"sort_by" validate:"omitempty,oneof=name deadline created_at updated_at"`
    SortOrder   string `json:"sort_order" query:"sort_order" validate:"omitempty,oneof=asc desc"`
    Page        int    `json:"page" query:"page" validate:"gte=1"`
    PageSize    int    `json:"page_size" query:"page_size" validate:"gte=1,lte=100"`
}

func (s *ProjectService) GetAll(ctx context.Context, auth0ID string, req ProjectSearchRequest) ([]*Project, error) {
	return nil, nil
}

// NOTE(pencelheimer): Usually only allowed for Planning state or soft-delete for others
func (s *ProjectService) Delete(ctx context.Context, auth0ID string, projectID uint) (bool, error) {
	return false, nil
}

func ProjectFromModel(m *models.Project) *Project {
	// TODO(pencelheimer): Mapping logic
	return &Project{}
}

func ProjectFullFromModel(m *models.Project) *ProjectFull {
	// TODO(pencelheimer): Mapping logic
	return &ProjectFull{}
}
