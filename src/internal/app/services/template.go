package services

import (
	"context"
	"errors"
	"time"

	"pocketeer/internal/platform/database"
	"pocketeer/internal/platform/database/models"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Template struct {
	ID              uint           `json:"id"`
	Name            string         `json:"name"`
	Description     *string        `json:"description"`
	PlannedWorkTime *time.Duration `json:"planned_work_time"`
	PlannedIncome   *float32       `json:"planned_income"`
	UsageCount      int64          `json:"usage_count"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type TemplateResourceSpecification struct {
	ID              uint    `json:"id"`
	TemplateID      uint    `json:"template_id"`
	ItemTypeID      uint    `json:"item_type_id"`
	ItemTypeName    string  `json:"item_type_name"`
	PlannedQuantity float32 `json:"planned_quantity"`
}

type TemplateFull struct {
	Template
	Specifications []TemplateResourceSpecification `json:"specifications"`
}

type TemplateCreateRequest struct {
	Name            string   `json:"name" validate:"required,min=3,max=256"`
	Description     *string  `json:"description" validate:"omitempty,max=512"`
	PlannedWorkTime *int64   `json:"planned_work_time" validate:"omitempty,gte=0"`
	PlannedIncome   *float32 `json:"planned_income" validate:"omitempty,gte=0"`
}

type TemplateUpdateRequest struct {
	Name            string   `json:"name" validate:"omitempty,min=3,max=256"`
	Description     *string  `json:"description" validate:"omitempty,max=512"`
	PlannedWorkTime *int64   `json:"planned_work_time" validate:"omitempty,gte=0"`
	PlannedIncome   *float32 `json:"planned_income" validate:"omitempty,gte=0"`
}

type TemplateSearchRequest struct {
	Query     string `json:"query" query:"query" validate:"omitempty,min=3"`
	SortBy    string `json:"sort_by" query:"sort_by" validate:"omitempty,oneof=name planned_work_time planned_income usage_count created_at"`
	SortOrder string `json:"sort_order" query:"sort_order" validate:"omitempty,oneof=asc desc"`
	Page      int    `json:"page" query:"page" validate:"gte=1"`
	PageSize  int    `json:"page_size" query:"page_size" validate:"gte=1,lte=100"`
}

type TemplateService struct {
	db             *database.DB
	userService    *UserService
	projectService *ProjectService
}

func NewTemplateService(db *database.DB, userService *UserService, projectService *ProjectService) *TemplateService {
	return &TemplateService{db, userService, projectService}
}

func TemplateResourceSpecificationFromModel(m models.TemplateResourceSpecification) TemplateResourceSpecification {
	return TemplateResourceSpecification{
		ID:              m.ID,
		TemplateID:      m.TemplateID,
		ItemTypeID:      m.ItemTypeID,
		ItemTypeName:    m.ItemType.Name,
		PlannedQuantity: m.PlannedQuantity,
	}
}

func TemplateFromModel(m *models.ProjectTemplate) *Template {
	var workTime *time.Duration
	if m.PlannedWorkTime.Microseconds() != 0 {
		val := time.Duration(m.PlannedWorkTime.Microseconds()) * time.Microsecond
		workTime = &val
	}

	var income *float32
	if m.PlannedIncome != nil {
		income = m.PlannedIncome
	}

	return &Template{
		ID:              m.ID,
		Name:            m.Name,
		Description:     m.Description,
		PlannedWorkTime: workTime,
		PlannedIncome:   income,
		UsageCount:      m.UsageCount,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func TemplateFullFromModel(m *models.ProjectTemplate) *TemplateFull {
	full := TemplateFull{
		Template:       *TemplateFromModel(m),
		Specifications: make([]TemplateResourceSpecification, len(m.RequiredResources)),
	}
	for i, spec := range m.RequiredResources {
		full.Specifications[i] = TemplateResourceSpecificationFromModel(spec)
	}
	return &full
}

func (s *TemplateService) Create(ctx context.Context, auth0ID string, req TemplateCreateRequest) (*Template, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	var existingTemplate models.ProjectTemplate
	err = s.db.WithContext(ctx).
		Where("user_id = ? AND name = ?", userID, req.Name).
		First(&existingTemplate).Error

	if err == nil {
		return nil, ErrTemplateNameAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDatabaseError
	}

	var plannedWorkTime *time.Duration
	if req.PlannedWorkTime != nil {
		val := time.Duration(*req.PlannedWorkTime)
		plannedWorkTime = &val
	}

	template := models.ProjectTemplate{
		Name:            req.Name,
		UserID:          userID,
		Description:     req.Description,
		PlannedIncome:   req.PlannedIncome,
		PlannedWorkTime: plannedWorkTime,
	}

	if err := s.db.WithContext(ctx).Create(&template).Error; err != nil {
		return nil, ErrDatabaseError
	}

	return TemplateFromModel(&template), nil
}

func (s *TemplateService) CreateFromProject(ctx context.Context, auth0ID string, projectID uint, templateName string) (*Template, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	var existingTemplate models.ProjectTemplate
	err = s.db.WithContext(ctx).
		Where("user_id = ? AND name = ?", userID, templateName).
		First(&existingTemplate).Error

	if err == nil {
		return nil, ErrTemplateNameAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrDatabaseError
	}

	projectFull, err := s.projectService.Get(ctx, auth0ID, projectID)
	if err != nil {
		return nil, err
	}

	project := projectFull.Project

	var templateWorkTime *time.Duration
	var templateIncome *float32
	resourceMap := make(map[uint]float32)

	if project.State == models.ProjectStatePlanning {
		templateWorkTime = project.PlannedWorkTime
		templateIncome = project.PlannedIncome

		for _, spec := range projectFull.Specifications {
			resourceMap[spec.ItemTypeID] = spec.PlannedQuantity
		}
	} else {
		if project.ActualIncome != nil {
			templateIncome = project.ActualIncome
		} else {
			templateIncome = project.PlannedIncome
		}

		if project.ActualWorkTime != nil {
			templateWorkTime = project.ActualWorkTime
		} else {
			templateWorkTime = project.PlannedWorkTime
		}

		for _, spec := range projectFull.Specifications {
			var usedQuantitySum float32
			for _, res := range spec.Reservations {
				usedQuantitySum += res.UsedQuantity
			}
			if usedQuantitySum > 0 {
				resourceMap[spec.ItemTypeID] = usedQuantitySum
			}
		}
	}

	var newTemplate models.ProjectTemplate

	tx_func := func(tx *gorm.DB) error {
		var workTimeDuration *time.Duration
		if templateWorkTime != nil {
			workTimeDuration = templateWorkTime
		}

		newTemplate = models.ProjectTemplate{
			UserID:          userID,
			Name:            templateName,
			Description:     project.Description,
			PlannedWorkTime: workTimeDuration,
			PlannedIncome:   templateIncome,
			UsageCount:      0,
		}

		if err := tx.Create(&newTemplate).Error; err != nil {
			return ErrDatabaseError
		}

		for itemTypeID, quantity := range resourceMap {
			spec := models.TemplateResourceSpecification{
				TemplateID:      newTemplate.ID,
				ItemTypeID:      itemTypeID,
				PlannedQuantity: quantity,
			}
			if err := tx.Create(&spec).Error; err != nil {
				return ErrDatabaseError
			}
		}

		return nil
	}

	if err := s.db.WithContext(ctx).Transaction(tx_func); err != nil {
		return nil, err
	}

	return TemplateFromModel(&newTemplate), nil
}

func (s *TemplateService) Update(ctx context.Context, auth0ID string, templateID uint, req TemplateUpdateRequest) (*Template, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	var template models.ProjectTemplate
	err = s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", templateID, userID).
		First(&template).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, ErrDatabaseError
	}

	hasUpdates := false

	if req.Name != "" && req.Name != template.Name {
		var existing models.ProjectTemplate
		err = s.db.WithContext(ctx).
			Where("user_id = ? AND name = ? AND id != ?", userID, req.Name, templateID).
			First(&existing).Error

		if err == nil {
			return nil, ErrTemplateNameAlreadyExists
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDatabaseError
		}
		template.Name = req.Name
		hasUpdates = true
	}

	if req.Description != nil {
		template.Description = req.Description
		hasUpdates = true
	}
	if req.PlannedIncome != nil {
		template.PlannedIncome = req.PlannedIncome
		hasUpdates = true
	}
	if req.PlannedWorkTime != nil {
		val := time.Duration(*req.PlannedWorkTime)
		template.PlannedWorkTime = &val
		hasUpdates = true
	}

	if hasUpdates {
		if err := s.db.WithContext(ctx).Save(&template).Error; err != nil {
			return nil, ErrDatabaseError
		}
	}

	return TemplateFromModel(&template), nil
}

func (s *TemplateService) Get(ctx context.Context, auth0ID string, templateID uint) (*TemplateFull, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	var template models.ProjectTemplate
	err = s.db.WithContext(ctx).
		Preload("ResourceSpecifications.ItemType").
		Where("id = ? AND user_id = ?", templateID, userID).
		First(&template).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, ErrDatabaseError
	}

	return TemplateFullFromModel(&template), nil
}

func (s *TemplateService) GetAll(ctx context.Context, auth0ID string, req TemplateSearchRequest) ([]*Template, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	query := s.db.WithContext(ctx).Model(&models.ProjectTemplate{}).Where("user_id = ?", userID)

	if req.Query != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+req.Query+"%", "%"+req.Query+"%")
	}

	orderBy := "created_at DESC"
	if req.SortBy != "" {
		validSortColumns := map[string]string{
			"name":              "name",
			"planned_work_time": "planned_work_time",
			"planned_income":    "planned_income",
			"usage_count":       "usage_count",
			"created_at":        "created_at",
		}
		column, ok := validSortColumns[req.SortBy]
		if !ok {
			return nil, ErrTemplateServiceInvalidSort
		}

		order := "ASC"
		if req.SortOrder == "desc" {
			order = "DESC"
		}
		orderBy = column + " " + order
	}
	query = query.Order(orderBy)

	offset := (req.Page - 1) * req.PageSize
	query = query.Offset(offset).Limit(req.PageSize)

	var templates []models.ProjectTemplate
	if err := query.Find(&templates).Error; err != nil {
		return nil, ErrDatabaseError
	}

	result := make([]*Template, len(templates))
	for i, t := range templates {
		result[i] = TemplateFromModel(&t)
	}

	return result, nil
}

func (s *TemplateService) Delete(ctx context.Context, auth0ID string, templateID uint) (bool, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return false, err
	}

	result := s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", templateID, userID).
		Delete(&models.ProjectTemplate{})

	if result.Error != nil {
		return false, ErrDatabaseError
	}

	if result.RowsAffected == 0 {
		return false, ErrTemplateNotFound
	}

	return true, nil
}

func (s *TemplateService) AddPlannedResource(
	ctx context.Context,
	auth0ID string,
	templateID uint,
	req AddPlannedResourceRequest,
) (*TemplateResourceSpecification, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, err
	}

	var template models.ProjectTemplate
	err = s.db.WithContext(ctx).
		Select("id").
		Where("id = ? AND user_id = ?", templateID, userID).
		First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, ErrDatabaseError
	}

	var itemType models.ItemType
	err = s.db.WithContext(ctx).
		Select("id", "name").
		Where("id = ? AND user_id = ?", req.ItemTypeID, userID).
		First(&itemType).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrItemTypeNotFound
		}
		return nil, ErrDatabaseError
	}

	templ := models.TemplateResourceSpecification{
		TemplateID:      templateID,
		ItemTypeID:      req.ItemTypeID,
		ResourceType:    models.ResourceType(req.ResourceType),
		PlannedQuantity: req.PlannedQuantity,
	}
	if err := s.db.WithContext(ctx).Create(&templ).Error; err != nil {
		if isUniqueConstraintError(err) {
			return nil, ErrTemplateResourceAlreadyExists
		}
		return nil, ErrDatabaseError
	}

	templ.ItemType = itemType

	dto := TemplateResourceSpecificationFromModel(templ)
	return &dto, nil
}

func (s *TemplateService) RemovePlannedResource(
	ctx context.Context,
	auth0ID string,
	templateID uint,
	specID uint,
) (bool, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return false, err
	}

	var template models.ProjectTemplate
	err = s.db.WithContext(ctx).
		Select("id").
		Where("id = ? AND user_id = ?", templateID, userID).
		First(&template).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrTemplateNotFound
		}
		return false, ErrDatabaseError
	}

	result := s.db.WithContext(ctx).
		Unscoped().
		Where("id = ? AND template_id = ?", specID, templateID).
		Delete(&models.TemplateResourceSpecification{})
	if result.Error != nil {
		return false, ErrDatabaseError
	}

	if result.RowsAffected == 0 {
		return false, ErrResourceSpecificationNotFound
	}

	return true, nil
}

func isUniqueConstraintError(err error) bool {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
