package services

import (
	"context"
	"errors"
	"log"
	"pocketeer/internal/platform/database"
	"pocketeer/internal/platform/database/models"
	"time"

	"gorm.io/gorm"
)

type TagService struct {
	db          *database.DB
	userService *UserService
}

func NewTagService(db *database.DB, userService *UserService) *TagService {
	return &TagService{db, userService}
}

type Tag struct {
	ID    uint    `json:"id"`
	Name  string  `json:"name"`
	Color *string `json:"color,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func TagFromModel(m models.Tag) Tag {
	return Tag{
		ID:        m.ID,
		Name:      m.Name,
		Color:     m.Color,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type TagFull struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	Color       *string   `json:"color"`
	Name        string    `json:"name"`
	ItemIDs     []uint    `json:"item_ids"`
	ItemTypeIDs []uint    `json:"item_type_ids"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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

type CreateTagRequest struct {
	Name       string  `json:"name" validate:"required,max=256"`
	Color      *string `json:"color" validate:"omitempty,len=6,hexadecimal"`
	TargetType *string `json:"target_type" validate:"omitempty,oneof=item item_type"`
	TargetID   *uint   `json:"target_id" validate:"omitempty,gt=0"`
}

func (s *TagService) Create(ctx context.Context, auth0ID string, req CreateTagRequest) (*TagFull, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, ErrUnauthenticated
	}

	tag := models.Tag{
		Name:   req.Name,
		Color:  req.Color,
		UserID: userID,
	}

	var itemIDs, itemTypeIDs []uint

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&tag).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrTagAlreadyExists
			}
			return err
		}

		if req.TargetType != nil && req.TargetID != nil {
			switch *req.TargetType {
			case "item":
				if err := tx.Model(&tag).Association("Items").
					Append(&models.Item{Model: gorm.Model{ID: *req.TargetID}}); err != nil {
					return err
				}
				itemIDs = []uint{*req.TargetID}
			case "item_type":
				if err := tx.Model(&tag).Association("ItemTypes").
					Append(&models.ItemType{Model: gorm.Model{ID: *req.TargetID}}); err != nil {
					return err
				}
				itemTypeIDs = []uint{*req.TargetID}
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("ERROR creating tag: %v", err)
		return nil, ErrDatabaseError
	}

	return TagFullFromModel(&tag, itemIDs, itemTypeIDs), nil
}

func (s *TagService) Get(ctx context.Context, auth0ID string, tagID uint) (*TagFull, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, ErrUnauthenticated
	}

	var tag models.Tag
	err = s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", tagID, userID).
		First(&tag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, ErrDatabaseError
	}

	var itemIDs []uint
	err = s.db.Model(&models.Item{}).
		Select(`"Items".id`).
		Joins(`JOIN "item_tags" itt ON itt.item_id = "Items".id`).
		Where("itt.tag_id = ?", tag.ID).
		Pluck(`"Items".id`, &itemIDs).Error
	if err != nil {
		return nil, ErrDatabaseError
	}

	var itemTypeIDs []uint
	err = s.db.Model(&models.ItemType{}).
		Select(`"ItemTypes".id`).
		Joins(`JOIN "item_type_tags" ittt ON ittt.item_type_id = "ItemTypes".id`).
		Where("ittt.tag_id = ?", tag.ID).
		Pluck(`"ItemTypes".id`, &itemTypeIDs).Error
	if err != nil {
		return nil, ErrDatabaseError
	}

	return TagFullFromModel(&tag, itemIDs, itemTypeIDs), nil
}

func (s *TagService) GetAll(ctx context.Context, auth0ID string) ([]*TagFull, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, ErrUnauthenticated
	}

	var tags []models.Tag
	err = s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&tags).Error

	if err != nil {
		return nil, ErrDatabaseError
	}

	result := make([]*TagFull, 0, len(tags))
	for i := range tags {
		result = append(result, TagFullFromModel(&tags[i], nil, nil))
	}

	return result, nil
}

type UpdateTagRequest struct {
	Name  *string `json:"name" validate:"omitempty,max=256"`
	Color *string `json:"color" validate:"omitempty,len=6,hexadecimal"`
}

func (s *TagService) Update(ctx context.Context, auth0ID string, tagID uint, req UpdateTagRequest) (*TagFull, error) {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return nil, ErrUnauthenticated
	}

	updates := map[string]any{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Color != nil {
		updates["color"] = req.Color
	}

	if len(updates) == 0 {
		var tag models.Tag
		if err := s.db.WithContext(ctx).
			Where("id = ? AND user_id = ?", tagID, userID).
			First(&tag).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrTagNotFound
			}
			return nil, ErrDatabaseError
		}
		return TagFullFromModel(&tag, nil, nil), nil
	}

	err = s.db.WithContext(ctx).
		Model(&models.Tag{}).
		Where("id = ? AND user_id = ?", tagID, userID).
		Updates(updates).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrTagAlreadyExists
		}
		return nil, ErrDatabaseError
	}

	var updatedTag models.Tag
	err = s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", tagID, userID).
		First(&updatedTag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTagNotFound
		}
		return nil, ErrDatabaseError
	}

	return TagFullFromModel(&updatedTag, nil, nil), nil
}

func (s *TagService) Delete(ctx context.Context, auth0ID string, tagID uint) error {
	userID, err := s.userService.GetUserIDByAuth0ID(ctx, auth0ID)
	if err != nil {
		return ErrUnauthenticated
	}

	result := s.db.WithContext(ctx).
		Unscoped().
		Where("id = ? AND user_id = ?", tagID, userID).
		Delete(&models.Tag{})

	if result.Error != nil {
		return ErrDatabaseError
	}

	if result.RowsAffected == 0 {
		return ErrTagNotFound
	}

	return nil
}
