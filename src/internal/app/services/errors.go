package services

import "errors"

// WARN(noatu): dot not add ErrForbidden, use *NotFound
var ( // Remember to sort that!
	ErrDatabaseError      = errors.New("database error")
	ErrFileSystemError    = errors.New("file system error")
	ErrForeignKeyViolated = errors.New("resource is still in use")
	ErrImageTooLarge      = errors.New("image is too large")
	ErrInvalidImageFormat = errors.New("invalid image format")
	ErrItemNotFound       = errors.New("item not found")
	ErrItemTypeNotFound   = errors.New("item type not found")
	ErrNotImplemented     = errors.New("method is not yet implemented")
	ErrPictureNotFound    = errors.New("picture not found")
	ErrTagAlreadyExists   = errors.New("tag already exists")
	ErrTagNotFound        = errors.New("tag not found")
	ErrUnauthenticated    = errors.New("unauthenticated, please log in (again)")
	ErrUserNotFound       = errors.New("user not found")
)
