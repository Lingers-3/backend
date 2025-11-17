package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"

	"pocketeer/internal/platform/database"
	"pocketeer/internal/platform/database/models"

	"gorm.io/gorm"
)

// QUESTION(noatu): move that to Config or nah?
const (
	MaxImageSize   = 10 * 1024 * 1024 // 10MiB
	ImageDirectory = "/var/pocketeer/img"
)

var allowedMimeTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

type PictureService struct {
	db *database.DB
}

func NewPictureService(db *database.DB) *PictureService {
	return &PictureService{db}
}

type PictureInfo struct {
	ID               uint   `json:"id"`
	OriginalFilename string `json:"original_filename"`
	MimeType         string `json:"mime_type"`
	Size             int64  `json:"size"`
}

func pictureInfoFromModel(m *models.Picture) *PictureInfo {
	return &PictureInfo{
		ID:               m.ID,
		OriginalFilename: m.OriginalFilename,
		MimeType:         m.MimeType,
		Size:             m.Size,
	}
}

func (s *PictureService) Upload(ctx context.Context, auth0ID string, fileHeader *multipart.FileHeader) (*PictureInfo, error) {
	// Need to be authenticated to upload images, and then be authorized to delete them
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return nil, err
	}

	if fileHeader.Size > MaxImageSize {
		return nil, ErrImageTooLarge
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("ERROR: opening uploaded file: %v", err)
		return nil, ErrFileSystemError
	}
	defer file.Close()

	// Reading to get MIME and hash
	content, err := io.ReadAll(file)
	if err != nil {
		log.Printf("ERROR: reading file content: %v", err)
		return nil, ErrFileSystemError
	}

	mimeType := detectMimeType(content)
	ext, ok := allowedMimeTypes[mimeType]
	if !ok {
		return nil, ErrInvalidImageFormat
	}

	hash := sha256.Sum256(content)
	hashStr := hex.EncodeToString(hash[:])

	if err := os.MkdirAll(ImageDirectory, 0755); err != nil {
		log.Printf("ERROR: creating image directory: %v", err)
		return nil, ErrFileSystemError
	}

	// NOTE(noatu): better to write the file first and then fail at database
	// than write to db and fail to write the file. Besides, most of the time
	// user would try to upload the same file again, so checking if it exists:
	filepath := filepath.Join(ImageDirectory, hashStr+ext)
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		if err := os.WriteFile(filepath, content, 0644); err != nil {
			log.Printf("ERROR: writing file to disk: %v", err)
			return nil, ErrFileSystemError
		}
	}

	// Check if db record exists for the user
	var picture models.Picture
	err = s.db.WithContext(ctx).
		Where("hash = ? AND user_id = ?", hashStr, userID).
		First(&picture).Error
	if err == nil {
		// Record already exists for the user
		return pictureInfoFromModel(&picture), nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("ERROR: checking for existing picture: %v", err)
		return nil, ErrDatabaseError
	}

	// Create new database record
	picture = models.Picture{
		Hash:             hashStr,
		OriginalFilename: fileHeader.Filename,
		MimeType:         mimeType,
		Size:             fileHeader.Size,
		UserID:           userID,
	}

	if err := s.db.WithContext(ctx).Create(&picture).Error; err != nil {
		log.Printf("ERROR: creating picture record: %v", err)
		return nil, ErrDatabaseError
	}

	return pictureInfoFromModel(&picture), nil
}

// Retrieve picture metadata
func (s *PictureService) Get(ctx context.Context, auth0ID string, pictureID uint) (*PictureInfo, error) {
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return nil, err
	}

	var picture models.Picture
	err = s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", pictureID, userID).
		First(&picture).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPictureNotFound
		}
		log.Printf("ERROR: fetching picture: %v", err)
		return nil, ErrDatabaseError
	}

	return pictureInfoFromModel(&picture), nil
}

type PictureFile struct {
	Filename string
	MimeType string
	Content  []byte
}

// Retrieves the file content with some metadata
func (s *PictureService) GetFile(ctx context.Context, auth0ID string, pictureID uint) (*PictureFile, error) {
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return nil, err
	}

	var picture models.Picture
	err = s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", pictureID, userID).
		First(&picture).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPictureNotFound
		}
		log.Printf("ERROR: fetching picture: %v", err)
		return nil, ErrDatabaseError
	}

	ext := allowedMimeTypes[picture.MimeType]
	filepath := filepath.Join(ImageDirectory, picture.Hash+ext)
	content, err := os.ReadFile(filepath)
	if err != nil {
		log.Printf("ERROR: reading image file: %v", err)
		return nil, ErrFileSystemError
	}

	return &PictureFile{
		Filename: picture.OriginalFilename,
		MimeType: picture.MimeType,
		Content:  content,
	}, nil
}

func (s *PictureService) Delete(ctx context.Context, auth0ID string, pictureID uint) error {
	userID, err := GetUserIDByAuth0ID(ctx, s.db, auth0ID)
	if err != nil {
		return err
	}

	var picture models.Picture
	err = s.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", pictureID, userID).
		First(&picture).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPictureNotFound
		}
		log.Printf("ERROR: fetching picture info for deletion: %v", err)
		return ErrDatabaseError
	}

	// HACK(noatu): I want to keep the Picture model simple and not introduce constraints in it.
	// Because of that, everything that uses Picture should be manually checked here.
	var count int64
	err = s.db.WithContext(ctx).
		Model(&models.ItemType{}).
		Where("user_id = ? AND picture_id = ?", userID, pictureID).
		Count(&count).Error
	if err != nil {
		log.Printf("ERROR: checking picture references: %v", err)
		return ErrDatabaseError
	}
	if count > 0 {
		return ErrForeignKeyViolated
	}

	// TODO: Add checks for other models that reference Picture

	err = s.db.WithContext(ctx).Unscoped().Delete(&picture).Error
	if err != nil {
		log.Printf("ERROR: deleting picture record: %v", err)
		return ErrDatabaseError
	}

	// Only delete the file if no other users reference it's hash
	err = s.db.WithContext(ctx).
		Model(&models.Picture{}).
		Where("hash = ?", picture.Hash).
		Count(&count).Error
	if err != nil {
		log.Printf("WARNING: failed to check hash references: %v", err)
		// Continue anyway, DB record already deleted
	}
	if count == 0 {
		ext := allowedMimeTypes[picture.MimeType]
		filepath := filepath.Join(ImageDirectory, picture.Hash+ext)
		if err := os.Remove(filepath); err != nil {
			log.Printf("WARNING: failed to delete image file %s: %v", filepath, err)
			// IGNORE error, DB record is already deleted
		}
	}

	return nil
}

// Detects MIME type from file content by checking magic bytes for common image formats
func detectMimeType(content []byte) string {
	if len(content) < 12 {
		return ""
	}

	// JPEG
	if content[0] == 0xFF && content[1] == 0xD8 && content[2] == 0xFF {
		return "image/jpeg"
	}

	// PNG
	if content[0] == 0x89 && content[1] == 0x50 && content[2] == 0x4E && content[3] == 0x47 {
		return "image/png"
	}

	// WebP
	if string(content[0:4]) == "RIFF" && string(content[8:12]) == "WEBP" {
		return "image/webp"
	}

	// GIF
	if string(content[0:3]) == "GIF" {
		return "image/gif"
	}

	return ""
}
