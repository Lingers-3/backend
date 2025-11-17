package services

import "errors"

var (
	ErrUnauthenticated = errors.New("unauthenticated, please log in (again)")

	ErrForeignKeyViolated = errors.New("resource is still in use")
	ErrTagAlreadyExists   = errors.New("tag already exists")

	ErrItemTypeNotFound = errors.New("item type not found")
	ErrItemNotFound     = errors.New("item not found")
	ErrTagNotFound      = errors.New("tag not found")
	ErrPictureNotFound  = errors.New("picture not found")

	ErrInvalidImageFormat = errors.New("invalid image format")
	ErrImageTooLarge      = errors.New("image is too large")

	ErrNotImplemented = errors.New("method is not yet implemented")

	// NOTE(noatu): errors below are internal details
	ErrFileSystemError = errors.New("file system error")
	ErrDatabaseError   = errors.New("database error")

	// WARN(noatu): claims that something exists, use *NotFound
	// ErrForbidden     = errors.New("forbidden")
)
