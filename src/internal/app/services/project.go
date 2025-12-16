package services

import (
	"context"
	"errors"
	"time"

	"pocketeer/internal/platform/database"
	"pocketeer/internal/platform/database/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

type ResourceSpecification struct {
	ID              uint                `json:"id"`
	ItemTypeID      uint                `json:"item_type_id"`
	ItemTypeName    string              `json:"item_type_name"`
	ResourceType    models.ResourceType `json:"resource_type"`
	PlannedQuantity float32             `json:"planned_quantity"`
}

type ResourceSpecificationFull struct {
	ResourceSpecification
	Reservations []ResourceReservation `json:"reservations"`
}

type ResourceReservation struct {
	ID                      uint    `json:"id"`
	ResourceSpecificationID uint    `json:"resource_specification_id"`
	ItemID                  uint    `json:"item_id"`
	ItemDescription         *string `json:"item_description"` // NOTE(pencelheimer): For convenience
	ReservedQuantity        float32 `json:"reserved_quantity"`
	UsedQuantity            float32 `json:"used_quantity"`
}

type ProjectFull struct {
	Project
	Specifications []ResourceSpecificationFull `json:"specifications"`
}

type ProjectCreateRequest struct {
	Name            string     `json:"name" validate:"required,min=1,max=256"`
	Description     *string    `json:"description" validate:"omitempty,max=512"`
	PlannedDeadline *time.Time `json:"planned_deadline" validate:"omitempty,gt=now"`
	PlannedIncome   *float32   `json:"planned_income" validate:"omitempty,gte=0"`
	PlannedWorkTime *int64     `json:"planned_work_time" validate:"omitempty,gte=0"`
}

func (s *ProjectService) Create(ctx context.Context, auth0ID string, req ProjectCreateRequest) (*Project, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	var plannedWorkTime *time.Duration
	if req.PlannedWorkTime != nil {
		val := time.Duration(*req.PlannedWorkTime)
		plannedWorkTime = &val
	}

	project := models.Project{
		Name:   req.Name,
		State:  models.ProjectStatePlanning,
		UserID: userID,

		Description:     req.Description,
		PlannedDeadline: req.PlannedDeadline,
		PlannedIncome:   req.PlannedIncome,
		PlannedWorkTime: plannedWorkTime,
	}

	result := s.db.WithContext(ctx).Create(&project)
	if result.Error != nil {
		return nil, ErrDatabaseError
	}

	return ProjectFromModel(&project), nil
}

type ProjectCreateFromTemplateRequest struct {
	TemplateID      uint       `json:"template_id" validate:"required,gt=0"`
	Name            string     `json:"name" validate:"required,min=1,max=256"`
	Description     *string    `json:"description" validate:"omitempty,max=512"`
	PlannedDeadline *time.Time `json:"planned_deadline" validate:"omitempty,gt=now"`
}

func (s *ProjectService) CreateFromTemplate(
	ctx context.Context,
	auth0ID string,
	req ProjectCreateFromTemplateRequest,
) (*ProjectFull, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	var template models.ProjectTemplate
	err = s.db.WithContext(ctx).
		Preload("ResourceSpecifications").
		Where("id = ?", req.TemplateID).
		First(&template).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, ErrDatabaseError
	}

	project := models.Project{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		State:       models.ProjectStatePlanning,

		PlannedIncome:   template.PlannedIncome,
		PlannedWorkTime: template.PlannedWorkTime,
		PlannedDeadline: req.PlannedDeadline,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&project).Error; err != nil {
			return err
		}

		var resourceSpecs []models.ResourceSpecification
		for _, trs := range template.ResourceSpecifications {
			spec := models.ResourceSpecification{
				ProjectID:       project.ID,
				ItemTypeID:      trs.ItemTypeID,
				ResourceType:    trs.ResourceType,
				PlannedQuantity: trs.PlannedQuantity,
			}
			resourceSpecs = append(resourceSpecs, spec)
		}

		if len(resourceSpecs) > 0 {
			if err := tx.Create(&resourceSpecs).Error; err != nil {
				return err
			}
		}

		if err := tx.Model(&models.ProjectTemplate{}).
			Where("id = ?", template.ID).
			UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	fullProject, err := s.Get(ctx, auth0ID, project.ID)
	if err != nil {
		return nil, err
	}

	return fullProject, nil
}

type ProjectUpdateRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=256"`
	Description *string `json:"description" validate:"omitempty,max=512"`
}

func (s *ProjectService) Update(ctx context.Context, auth0ID string, projectID uint, req ProjectUpdateRequest) (*Project, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	var project models.Project
	err = s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", projectID, userID).
		First(&project).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, ErrDatabaseError
	}

	hasUpdates := false

	if req.Name != nil && *req.Name != project.Name {
		var count int64
		err := s.db.WithContext(ctx).Model(&models.Project{}).
			Where("user_id = ? AND name = ? AND id != ?", userID, *req.Name, projectID).
			Count(&count).Error
		if err != nil {
			return nil, ErrDatabaseError
		}

		if count > 0 {
			return nil, ErrProjectAlreadyExists
		}

		project.Name = *req.Name
		hasUpdates = true
	}

	if req.Description != nil {
		project.Description = req.Description
		hasUpdates = true
	}

	if hasUpdates {
		if err := s.db.WithContext(ctx).Save(&project).Error; err != nil {
			return nil, ErrDatabaseError
		}
	}

	return ProjectFromModel(&project), nil
}

// when State == Planning
type ProjectPlanUpdateRequest struct {
	PlannedDeadline *time.Time `json:"planned_deadline" validate:"omitempty,gt=now"`
	PlannedIncome   *float32   `json:"planned_income" validate:"omitempty,gte=0"`
	PlannedWorkTime *int64     `json:"planned_work_time" validate:"omitempty,gte=0"`
}

func (s *ProjectService) UpdatePlan(
	ctx context.Context,
	auth0ID string,
	projectID uint,
	req ProjectPlanUpdateRequest,
) (*Project, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	var project models.Project
	err = s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", projectID, userID).
		First(&project).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, ErrDatabaseError
	}

	if project.State != models.ProjectStatePlanning {
		return nil, ErrProjectNotPlanning
	}

	hasUpdates := false

	if req.PlannedDeadline != nil {
		project.PlannedDeadline = req.PlannedDeadline
		hasUpdates = true
	}
	if req.PlannedIncome != nil {
		project.PlannedIncome = req.PlannedIncome
		hasUpdates = true
	}
	if req.PlannedWorkTime != nil {
		dur := time.Duration(*req.PlannedWorkTime)
		project.PlannedWorkTime = &dur
		hasUpdates = true
	}

	if hasUpdates {
		if err := s.db.WithContext(ctx).Save(&project).Error; err != nil {
			return nil, ErrDatabaseError
		}
	}

	return ProjectFromModel(&project), nil
}

// when State == Planning
type AddPlannedResourceRequest struct {
	ItemTypeID      uint    `json:"item_type_id" validate:"required,gt=0"`
	ResourceType    string  `json:"resource_type" validate:"required,oneof=Consumable Instrument"`
	PlannedQuantity float32 `json:"planned_quantity" validate:"required,gt=0"`
}

func (s *ProjectService) AddPlannedResource(
	ctx context.Context,
	auth0ID string,
	projectID uint,
	req AddPlannedResourceRequest,
) (*ResourceSpecification, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	var project models.Project
	err = s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", projectID, userID).
		First(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, ErrDatabaseError
	}

	if project.State != models.ProjectStatePlanning {
		return nil, ErrProjectNotPlanning
	}

	var itemType models.ItemType
	err = s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", req.ItemTypeID, userID).
		First(&itemType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrItemTypeNotFound
		}
		return nil, ErrDatabaseError
	}

	spec := models.ResourceSpecification{
		ProjectID:       project.ID,
		ItemTypeID:      req.ItemTypeID,
		ResourceType:    models.ResourceType(req.ResourceType),
		PlannedQuantity: req.PlannedQuantity,
	}
	if err := s.db.WithContext(ctx).Create(&spec).Error; err != nil {
		return nil, ErrDatabaseError
	}

	spec.ItemType = itemType

	dto := ResourceSpecificationFromModel(spec)
	return &dto, nil
}

// when State == Planning
func (s *ProjectService) RemovePlannedResource(ctx context.Context, auth0ID string, projectID uint, specID uint) (bool, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return false, err
	}

	var project models.Project
	err = s.db.WithContext(ctx).
		Select("id", "state").
		Where("id = ? AND user_id = ?", projectID, userID).
		First(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrProjectNotFound
		}
		return false, ErrDatabaseError
	}

	if project.State != models.ProjectStatePlanning {
		return false, ErrProjectNotPlanning
	}

	result := s.db.WithContext(ctx).
		Unscoped().
		Where("id = ? AND project_id = ?", specID, project.ID).
		Delete(&models.ResourceSpecification{})
	if result.Error != nil {
		return false, ErrDatabaseError
	}

	if result.RowsAffected == 0 {
		return false, ErrResourceSpecificationNotFound
	}

	return true, nil
}

// Transition: Planning -> Active. Triggers auto-reservation.
func (s *ProjectService) Start(ctx context.Context, auth0ID string, projectID uint) (*ProjectFull, error) {
	tx_func := func(tx *gorm.DB) error {
		userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
		if err != nil {
			return err
		}

		var project models.Project
		if err := tx.
			Preload("ResourceSpecifications").
			Clauses(clause.Locking{Strength: "UPDATE"}). // NOTE(pencelheimer): lock to prevent race conditions
			Where("id = ? AND user_id = ?", projectID, userID).
			First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProjectNotFound
			}
			return ErrDatabaseError
		}

		if project.State != models.ProjectStatePlanning {
			return ErrProjectAlreadyActive
		}

		for _, spec := range project.ResourceSpecifications {
			remainingToReserve := spec.PlannedQuantity

			var items []models.Item
			if err := tx.
				Where("item_type_id = ?", spec.ItemTypeID).
				Order("expiration_date ASC NULLS LAST, id ASC").
				Find(&items).Error; err != nil {
				return ErrDatabaseError
			}

			for _, item := range items {
				if remainingToReserve <= 0 {
					break
				}

				var reservedSum float32
				if err := tx.Model(&models.ResourceReservation{}).
					Where("item_id = ?", item.ID).
					Select("COALESCE(SUM(reserved_quantity), 0)").
					Scan(&reservedSum).Error; err != nil {
					return ErrDatabaseError
				}

				available := item.Quantity - reservedSum

				if available <= 0 {
					continue
				}

				var amountToTake float32
				if available < remainingToReserve {
					amountToTake = available
				} else {
					amountToTake = remainingToReserve
				}

				reservation := models.ResourceReservation{
					ProjectID:               project.ID,
					ItemID:                  item.ID,
					ResourceSpecificationID: spec.ID,
					ReservedQuantity:        amountToTake,
					UsedQuantity:            0,
				}

				if err := tx.Create(&reservation).Error; err != nil {
					return ErrDatabaseError
				}

				remainingToReserve -= amountToTake
			}

			// NOTE(pencelheimer): remainingToReserve > 0 after the loop end means resource shortage, we should create an alert about it.
		}

		now := time.Now()
		project.State = models.ProjectStateActive
		project.StartedAt = &now

		if err := tx.Save(&project).Error; err != nil {
			return ErrDatabaseError
		}

		return nil
	}

	if err := s.db.WithContext(ctx).Transaction(tx_func); err != nil {
		return nil, err
	}

	return s.Get(ctx, auth0ID, projectID)
}

// when State == Active
type ProjectActualMetricsRequest struct {
	ActualDeadline *time.Time `json:"actual_deadline" validate:"omitempty"`
	ActualIncome   *float32   `json:"actual_income" validate:"omitempty,gte=0"`
	ActualWorkTime *int64     `json:"actual_work_time" validate:"omitempty,gte=0"`
}

func (s *ProjectService) UpdateActualMetrics(
	ctx context.Context,
	auth0ID string,
	projectID uint,
	req ProjectActualMetricsRequest,
) (*Project, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	var project models.Project
	err = s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", projectID, userID).
		First(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, ErrDatabaseError
	}

	if project.State != models.ProjectStateActive {
		return nil, ErrProjectNotActive
	}

	hasUpdates := false

	if req.ActualDeadline != nil {
		project.ActualDeadline = req.ActualDeadline
		hasUpdates = true
	}

	if req.ActualIncome != nil {
		project.ActualIncome = req.ActualIncome
		hasUpdates = true
	}

	if req.ActualWorkTime != nil {
		newTotal := time.Duration(*req.ActualWorkTime)
		project.ActualWorkTime = &newTotal
		hasUpdates = true
	}

	if hasUpdates {
		if err := s.db.WithContext(ctx).Save(&project).Error; err != nil {
			return nil, ErrDatabaseError
		}
	}

	return ProjectFromModel(&project), nil
}

// when State == Active
type ActiveResourceItemRequest struct {
	ItemID   uint    `json:"item_id" validate:"required"`
	Reserved float32 `json:"reserved" validate:"required,gt=0"`
	Used     float32 `json:"used" validate:"gte=0"`
}

type AddActiveResourceRequest struct {
	ItemTypeID   uint                        `json:"item_type_id" validate:"required,gt=0"`
	ResourceType string                      `json:"resource_type" validate:"required,oneof=Consumable Instrument"`
	Resources    []ActiveResourceItemRequest `json:"resources" validate:"required,min=1,dive"`
}

func (s *ProjectService) AddActiveResource(
	ctx context.Context,
	auth0ID string,
	projectID uint,
	req AddActiveResourceRequest,
) (*ProjectFull, error) {
	for _, res := range req.Resources {
		if res.Used > res.Reserved {
			return nil, ErrInvalidQuantity
		}
	}

	tx_func := func(tx *gorm.DB) error {
		userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
		if err != nil {
			return err
		}

		var project models.Project

		// NOTE(pencelheimer): lock to prevent race conditions
		err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", projectID, userID).
			First(&project).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProjectNotFound
			}
			return ErrDatabaseError
		}

		if project.State != models.ProjectStateActive {
			return ErrProjectNotActive
		}

		var itemType models.ItemType
		if err := tx.Where("id = ? AND user_id = ?", req.ItemTypeID, userID).First(&itemType).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrItemTypeNotFound
			}
			return ErrDatabaseError
		}

		spec := models.ResourceSpecification{
			ProjectID:       project.ID,
			ItemTypeID:      req.ItemTypeID,
			ResourceType:    models.ResourceType(req.ResourceType),
			PlannedQuantity: 0, // NOTE(pencelheimer): unplanned resource
		}

		if err := tx.Create(&spec).Error; err != nil {
			return ErrDatabaseError
		}

		for _, resReq := range req.Resources {
			var item models.Item
			if err := tx.Where("id = ? AND item_type_id = ?", resReq.ItemID, req.ItemTypeID).First(&item).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return ErrItemMismatch
				}
				return ErrDatabaseError
			}

			var reservedSum float32
			if err := tx.Model(&models.ResourceReservation{}).
				Where("item_id = ?", item.ID).
				Select("COALESCE(SUM(reserved_quantity), 0)").
				Scan(&reservedSum).Error; err != nil {
				return ErrDatabaseError
			}

			available := item.Quantity - reservedSum

			if available < resReq.Reserved {
				return ErrInsufficientResources
			}

			reservation := models.ResourceReservation{
				ProjectID:               project.ID,
				ItemID:                  item.ID,
				ResourceSpecificationID: spec.ID,
				ReservedQuantity:        resReq.Reserved,
				UsedQuantity:            resReq.Used,
			}

			if err := tx.Create(&reservation).Error; err != nil {
				return ErrDatabaseError
			}
		}

		return nil
	}

	err := s.db.WithContext(ctx).Transaction(tx_func)
	if err != nil {
		return nil, err
	}

	return s.Get(ctx, auth0ID, projectID)
}

// when State == Active
type UpdateResourceUsageRequest struct {
	UsedQuantity float32 `json:"used_quantity" validate:"required,gte=0"`
}

func (s *ProjectService) UpdateResourceUsage(
	ctx context.Context,
	auth0ID string,
	projectID uint,
	reservationID uint,
	req UpdateResourceUsageRequest,
) (*ResourceReservation, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	var reservation models.ResourceReservation
	err = s.db.WithContext(ctx).
		Preload("Project").
		Preload("Item").
		Where("id = ?", reservationID).
		First(&reservation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrResourceReservationNotFound
		}
		return nil, ErrDatabaseError
	}

	if reservation.Project.UserID != userID {
		return nil, ErrProjectNotFound
	}

	if reservation.ProjectID != projectID {
		return nil, ErrProjectNotFound
	}

	if reservation.Project.State != models.ProjectStateActive {
		return nil, ErrProjectNotActive
	}

	reservation.UsedQuantity = req.UsedQuantity

	if err := s.db.WithContext(ctx).Save(&reservation).Error; err != nil {
		return nil, ErrDatabaseError
	}

	dto := ResourceReservationDTOFromModel(reservation)
	return &dto, nil
}

// Transition: Active -> Canceled. Releases reservations.
type CancelProjectRequest struct {
	ReturnItemsToInventory bool `json:"return_items_to_inventory"`
}

func (s *ProjectService) Cancel(ctx context.Context, auth0ID string, projectID uint, req CancelProjectRequest) (*Project, error) {
	var project models.Project

	tx_func := func(tx *gorm.DB) error {
		userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
		if err != nil {
			return err
		}

		if err := tx.
			Preload("ResourceReservations").
			Preload("ResourceReservations.Item").
			Preload("ResourceReservations.ResourceSpecification").
			Clauses(clause.Locking{Strength: "UPDATE"}). // NOTE(pencelheimer): lock to prevent race conditions
			Where("id = ? AND user_id = ?", projectID, userID).
			First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProjectNotFound
			}
			return ErrDatabaseError
		}

		if project.State != models.ProjectStateActive {
			return ErrProjectNotActive
		}

		for _, reservation := range project.ResourceReservations {
			isConsumable := reservation.ResourceSpecification.ResourceType == models.ResourceTypeConsumable

			if !req.ReturnItemsToInventory && isConsumable && reservation.UsedQuantity > 0 {
				item := reservation.Item
				item.Quantity -= reservation.UsedQuantity

				if err := tx.Save(&item).Error; err != nil {
					return ErrDatabaseError
				}
			}
		}

		now := time.Now()
		project.State = models.ProjectStateCanceled
		project.FinishedAt = &now

		if err := tx.Save(&project).Error; err != nil {
			return ErrDatabaseError
		}

		return nil
	}
	err := s.db.WithContext(ctx).Transaction(tx_func)

	if err != nil {
		return nil, err
	}

	return ProjectFromModel(&project), nil
}

// Transition: Active -> Completed. Finalizes usage.
type CompleteProjectRequest struct {
	ActualRevenue     *float32 `json:"actual_revenue" validate:"omitempty,gte=0"`
	FinalizeInventory bool     `json:"finalize_inventory"` // NOTE(pencelheimer): If true, remaining reserved but unused items are returned
}

func (s *ProjectService) Complete(ctx context.Context, auth0ID string, projectID uint, req CompleteProjectRequest) (*Project, error) {
	var project models.Project
	tx_func := func(tx *gorm.DB) error {
		userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
		if err != nil {
			return err
		}

		if err := tx.
			Preload("ResourceReservations").
			Preload("ResourceReservations.Item").
			Preload("ResourceReservations.ResourceSpecification").
			Clauses(clause.Locking{Strength: "UPDATE"}). // NOTE(pencelheimer): lock to prevent race conditions
			Where("id = ? AND user_id = ?", projectID, userID).
			First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProjectNotFound
			}
			return ErrDatabaseError
		}

		if project.State != models.ProjectStateActive {
			return ErrProjectNotActive
		}

		for _, reservation := range project.ResourceReservations {
			isConsumable := reservation.ResourceSpecification.ResourceType == models.ResourceTypeConsumable

			if req.FinalizeInventory && isConsumable && reservation.UsedQuantity > 0 {
				item := reservation.Item
				item.Quantity -= reservation.UsedQuantity

				if err := tx.Save(&item).Error; err != nil {
					return ErrDatabaseError
				}
			}
		}

		now := time.Now()
		project.State = models.ProjectStateCompleted
		project.FinishedAt = &now

		if req.ActualRevenue != nil {
			project.ActualIncome = req.ActualRevenue
		}

		if err := tx.Save(&project).Error; err != nil {
			return ErrDatabaseError
		}

		return nil
	}
	err := s.db.WithContext(ctx).Transaction(tx_func)

	if err != nil {
		return nil, err
	}

	return ProjectFromModel(&project), nil
}

func (s *ProjectService) Get(ctx context.Context, auth0ID string, projectID uint) (*ProjectFull, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	var project models.Project
	err = s.db.WithContext(ctx).
		// NOTE(noatu): Eager load associations for the full project view
		Preload("ResourceSpecifications.ItemType").
		Preload("ResourceSpecifications.ResourceReservations.Item").
		Where("id = ? AND user_id = ?", projectID, userID).
		First(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, ErrDatabaseError
	}

	return ProjectFullFromModel(&project), nil
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
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	query := s.db.WithContext(ctx).Model(&models.Project{}).Where("user_id = ?", userID)

	if req.Query != "" {
		// NOTE(noatu): using case insensitive matching
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+req.Query+"%", "%"+req.Query+"%")
	}
	if req.State != "" {
		query = query.Where("state = ?", req.State)
	}
	if req.HasDeadline != nil {
		if *req.HasDeadline {
			query = query.Where("planned_deadline IS NOT NULL")
		} else {
			query = query.Where("planned_deadline IS NULL")
		}
	}
	if req.SortBy != "" {
		order := "asc"
		if req.SortOrder != "" {
			order = req.SortOrder
		}
		query = query.Order(req.SortBy + " " + order)
	}

	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	var projects []models.Project
	if err := query.Find(&projects).Error; err != nil {
		return nil, ErrDatabaseError
	}

	result := make([]*Project, len(projects))
	for i, p := range projects {
		result[i] = ProjectFromModel(&p)
	}

	return result, nil
}

// NOTE(pencelheimer): Usually only allowed for Planning state or soft-delete for others
func (s *ProjectService) Delete(ctx context.Context, auth0ID string, projectID uint) (bool, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return false, err
	}

	var project models.Project
	err = s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", projectID, userID).
		First(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrProjectNotFound
		}
		return false, ErrDatabaseError
	}

	if project.State == models.ProjectStatePlanning {
		// Hard delete for projects in the "Planning" state
		if err := s.db.WithContext(ctx).Unscoped().Delete(&project).Error; err != nil {
			return false, ErrDatabaseError
		}
	} else {
		// Soft delete for all other states
		if err := s.db.WithContext(ctx).Delete(&project).Error; err != nil {
			return false, ErrDatabaseError
		}
	}

	return true, nil
}

type AddReservationRequest struct {
	ItemID   uint    `json:"item_id" validate:"required"`
	Reserved float32 `json:"reserved" validate:"required,gt=0"`
	Used     float32 `json:"used" validate:"gte=0"`
}

func (s *ProjectService) AddReservation(
	ctx context.Context,
	auth0ID string,
	projectID uint,
	specID uint,
	req AddReservationRequest,
) (*ResourceReservation, error) {
	if req.Used > req.Reserved {
		return nil, ErrInvalidQuantity
	}

	var newReservation models.ResourceReservation

	tx_func := func(tx *gorm.DB) error {
		userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
		if err != nil {
			return err
		}

		var spec models.ResourceSpecification
		if err := tx.Joins("Project").
			Where(`"ResourceSpecifications".id = ? AND "ResourceSpecifications".project_id = ?`, specID, projectID).
			Where(`"Project".user_id = ?`, userID).
			First(&spec).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrResourceSpecificationNotFound
			}
			return ErrDatabaseError
		}

		if spec.Project.State != models.ProjectStateActive {
			return ErrProjectNotActive
		}

		var item models.Item
		if err := tx.Where("id = ? AND item_type_id = ?", req.ItemID, spec.ItemTypeID).
			First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrItemMismatch
			}
			return ErrDatabaseError
		}

		var reservedSum float32
		if err := tx.Model(&models.ResourceReservation{}).
			Where("item_id = ?", item.ID).
			Select("COALESCE(SUM(reserved_quantity), 0)").
			Scan(&reservedSum).Error; err != nil {
			return ErrDatabaseError
		}

		available := item.Quantity - reservedSum
		if available < req.Reserved {
			return ErrInsufficientResources
		}

		newReservation = models.ResourceReservation{
			ProjectID:               projectID,
			ResourceSpecificationID: specID,
			ItemID:                  item.ID,
			ReservedQuantity:        req.Reserved,
			UsedQuantity:            req.Used,
		}

		if err := tx.Create(&newReservation).Error; err != nil {
			return ErrDatabaseError
		}

		newReservation.Item = item
		newReservation.Project = spec.Project

		return nil
	}

	err := s.db.WithContext(ctx).Transaction(tx_func)
	if err != nil {
		return nil, err
	}

	dto := ResourceReservationDTOFromModel(newReservation)
	return &dto, nil
}

type UpdateReservationRequest struct {
	Reserved *float32 `json:"reserved" validate:"omitempty,gt=0"`
	Used     *float32 `json:"used" validate:"omitempty,gte=0"`
}

func (s *ProjectService) UpdateReservation(
	ctx context.Context,
	auth0ID string,
	projectID uint,
	specID uint,
	reservationID uint,
	req UpdateReservationRequest,
) (*ResourceReservation, error) {
	var reservation models.ResourceReservation

	tx_func := func(tx *gorm.DB) error {
		userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
		if err != nil {
			return err
		}

		if err := tx.Preload("Project").Preload("Item").
			Where("id = ? AND project_id = ? AND resource_specification_id = ?", reservationID, projectID, specID).
			First(&reservation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrResourceSpecificationNotFound
			}
			return ErrDatabaseError
		}

		if reservation.Project.UserID != userID {
			return ErrProjectNotFound
		}

		if reservation.Project.State != models.ProjectStateActive {
			return ErrProjectNotActive
		}

		newReserved := reservation.ReservedQuantity
		if req.Reserved != nil {
			newReserved = *req.Reserved
		}

		newUsed := reservation.UsedQuantity
		if req.Used != nil {
			newUsed = *req.Used
		}

		if newUsed > newReserved {
			return ErrInvalidQuantity
		}

		if req.Reserved != nil && *req.Reserved > reservation.ReservedQuantity {
			delta := *req.Reserved - reservation.ReservedQuantity

			var totalReservedForItem float32
			if err := tx.Model(&models.ResourceReservation{}).
				Where("item_id = ?", reservation.ItemID).
				Select("COALESCE(SUM(reserved_quantity), 0)").
				Scan(&totalReservedForItem).Error; err != nil {
				return ErrDatabaseError
			}

			currentlyAvailable := reservation.Item.Quantity - totalReservedForItem
			if currentlyAvailable < delta {
				return ErrInsufficientResources
			}
		}

		if req.Reserved != nil {
			reservation.ReservedQuantity = *req.Reserved
		}

		if req.Used != nil {
			reservation.UsedQuantity = *req.Used
		}

		if err := tx.Save(&reservation).Error; err != nil {
			return ErrDatabaseError
		}

		return nil
	}

	err := s.db.WithContext(ctx).Transaction(tx_func)
	if err != nil {
		return nil, err
	}

	dto := ResourceReservationDTOFromModel(reservation)
	return &dto, nil
}

func (s *ProjectService) DeleteReservation(
	ctx context.Context,
	auth0ID string,
	projectID uint,
	specID uint,
	reservationID uint,
) error {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return err
	}

	tx_func := func(tx *gorm.DB) error {
		var reservation models.ResourceReservation

		if err := tx.Preload("Project").
			Where("id = ? AND project_id = ? AND resource_specification_id = ?", reservationID, projectID, specID).
			First(&reservation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrResourceSpecificationNotFound
			}
			return ErrDatabaseError
		}

		if reservation.Project.UserID != userID {
			return ErrProjectNotFound
		}

		if reservation.Project.State != models.ProjectStateActive {
			return ErrProjectNotActive
		}

		if err := tx.Delete(&reservation).Error; err != nil {
			return ErrDatabaseError
		}

		return nil
	}

	err = s.db.WithContext(ctx).Transaction(tx_func)

	return err
}
