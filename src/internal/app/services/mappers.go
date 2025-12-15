package services

import (
	"pocketeer/internal/platform/database/models"
	"time"
)

func TagFromModel(m models.Tag) Tag {
	return Tag{
		ID:        m.ID,
		Name:      m.Name,
		Color:     m.Color,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func TagFullFromModel(m *models.Tag, itemIDs, itemTypeIDs []uint) *TagFull {
	if m == nil {
		return nil
	}

	return &TagFull{
		ID:          m.ID,
		UserID:      m.UserID,
		Color:       m.Color,
		Name:        m.Name,
		ItemIDs:     itemIDs,
		ItemTypeIDs: itemTypeIDs,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func ItemTypeFromModel(m *models.ItemType) *ItemType {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	itemIDs := make([]uint, len(m.Items))
	for i, item := range m.Items {
		itemIDs[i] = item.ID
	}

	tagIDs := make([]uint, len(m.Tags))
	for i, tag := range m.Tags {
		tagIDs[i] = tag.ID
	}

	var pictureHash *string
	if m.Picture != nil {
		pictureHash = &m.Picture.Hash
	}

	return &ItemType{
		ID:                     m.ID,
		Name:                   m.Name,
		Description:            m.Description,
		BaseMeasurementUnit:    m.BaseMeasurementUnit,
		DisplayMeasurementUnit: m.DisplayMeasurementUnit,
		DefaultQuantity:        m.DefaultQuantity,
		ShortageTreshold:       m.ShortageThreshold,
		PictureID:              m.PictureID,
		PictureHash:            pictureHash,
		ItemIDs:                itemIDs,
		TagIDs:                 tagIDs,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
		DeletedAt:              deletedAt,
	}
}

func ItemTypeFullFromModel(m *models.ItemType) *ItemTypeFull {
	items := make([]ItemFull, len(m.Items))
	for i, item := range m.Items {
		items[i] = ItemFullFromModel(item)
	}

	tags := make([]Tag, len(m.Tags))
	for i, tag := range m.Tags {
		tags[i] = TagFromModel(tag)
	}

	var pictureHash *string
	if m.Picture != nil {
		pictureHash = &m.Picture.Hash
	}

	return &ItemTypeFull{
		ID:                     m.ID,
		Name:                   m.Name,
		Description:            m.Description,
		BaseMeasurementUnit:    m.BaseMeasurementUnit,
		DisplayMeasurementUnit: m.DisplayMeasurementUnit,
		DefaultQuantity:        m.DefaultQuantity,
		ShortageThreshold:      m.ShortageThreshold,
		PictureID:              m.PictureID,
		PictureHash:            pictureHash,
		Items:                  items,
		Tags:                   tags,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
	}
}

func ItemFromModel(m *models.Item) *Item {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	tagIDs := make([]uint, len(m.Tags))
	for i, tag := range m.Tags {
		tagIDs[i] = tag.ID
	}

	return &Item{
		ID:                     m.ID,
		Description:            m.Description,
		Quantity:               m.Quantity,
		ReservedQuantity:       m.ReservedQuantity,
		ExpirationDate:         m.ExpirationDate,
		DisplayMeasurementUnit: m.DisplayMeasurementUnit,
		PurchasePrice:          m.PurchasePrice,
		ItemTypeID:             m.ItemTypeID,
		TagIDs:                 tagIDs,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
		DeletedAt:              deletedAt,
	}
}

func ItemFullFromModel(m models.Item) ItemFull {
	tags := make([]Tag, len(m.Tags))
	for i, t := range m.Tags {
		tags[i] = TagFromModel(t)
	}

	return ItemFull{
		ID:                     m.ID,
		Description:            m.Description,
		Quantity:               m.Quantity,
		ReservedQuantity:       m.ReservedQuantity,
		ExpirationDate:         m.ExpirationDate,
		DisplayMeasurementUnit: m.DisplayMeasurementUnit,
		PurchasePrice:          m.PurchasePrice,
		Tags:                   tags,
		CreatedAt:              m.CreatedAt,
		UpdatedAt:              m.UpdatedAt,
	}
}

func ResourceReservationDTOFromModel(m models.ResourceReservation) ResourceReservation {
	dto := ResourceReservation{
		ID:                      m.ID,
		ItemID:                  m.ItemID,
		ResourceSpecificationID: m.ResourceSpecificationID,
		ReservedQuantity:        m.ReservedQuantity,
		UsedQuantity:            m.UsedQuantity,
	}

	if m.Item.ID != 0 {
		dto.ItemDescription = m.Item.Description
	}

	return dto
}

func ResourceSpecificationFromModel(m models.ResourceSpecification) ResourceSpecification {
	return ResourceSpecification{
		ID:              m.ID,
		ItemTypeID:      m.ItemTypeID,
		ItemTypeName:    m.ItemType.Name,
		ResourceType:    m.ResourceType,
		PlannedQuantity: m.PlannedQuantity,
	}
}

func ResourceSpecificationFullFromModel(m models.ResourceSpecification) ResourceSpecificationFull {
	reservations := make([]ResourceReservation, len(m.ResourceReservations))
	for i, res := range m.ResourceReservations {
		reservations[i] = ResourceReservationDTOFromModel(res)
	}

	return ResourceSpecificationFull{
		ResourceSpecification: ResourceSpecificationFromModel(m),
		Reservations:          reservations,
	}
}

func ProjectFullFromModel(m *models.Project) *ProjectFull {
	specs := make([]ResourceSpecificationFull, len(m.ResourceSpecifications))
	for i, spec := range m.ResourceSpecifications {
		specs[i] = ResourceSpecificationFullFromModel(spec)
	}

	return &ProjectFull{
		Project: Project{
			ID:              m.ID,
			Name:            m.Name,
			Description:     m.Description,
			State:           m.State,
			PlannedDeadline: m.PlannedDeadline,
			PlannedIncome:   m.PlannedIncome,
			PlannedWorkTime: m.PlannedWorkTime,
			ActualDeadline:  m.ActualDeadline,
			ActualIncome:    m.ActualIncome,
			ActualWorkTime:  m.ActualWorkTime,
			StartedAt:       m.StartedAt,
			FinishedAt:      m.FinishedAt,
			CreatedAt:       m.CreatedAt,
			UpdatedAt:       m.UpdatedAt,
		},
		Specifications: specs,
	}
}

func ProjectFromModel(m *models.Project) *Project {
	return &Project{
		ID:              m.ID,
		Name:            m.Name,
		Description:     m.Description,
		State:           m.State,
		PlannedDeadline: m.PlannedDeadline,
		PlannedIncome:   m.PlannedIncome,
		PlannedWorkTime: m.PlannedWorkTime,
		ActualDeadline:  m.ActualDeadline,
		ActualIncome:    m.ActualIncome,
		ActualWorkTime:  m.ActualWorkTime,
		StartedAt:       m.StartedAt,
		FinishedAt:      m.FinishedAt,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}
