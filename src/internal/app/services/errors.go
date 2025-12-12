package services

import "errors"

// WARN(pencelheimer): use *NotFound instead of Forbidden
var (
	// NOTE(pencelheimer): Remember to sort this list
	ErrDatabaseError                 = errors.New("database error")
	ErrFileSystemError               = errors.New("file system error")
	ErrForeignKeyViolated            = errors.New("resource is still in use")
	ErrImageTooLarge                 = errors.New("image is too large")
	ErrInvalidImageFormat            = errors.New("invalid image format")
	ErrInvalidQuantity               = errors.New("used quantity cannot exceed reserved quantity")
	ErrItemNotFound                  = errors.New("item not found")
	ErrItemTypeNotFound              = errors.New("item type not found")
	ErrNotImplemented                = errors.New("method is not yet implemented")
	ErrPictureNotFound               = errors.New("picture not found")
	ErrProjectAlreadyActive          = errors.New("project is already active or completed")
	ErrProjectAlreadyExists          = errors.New("project with this name already exists")
	ErrProjectNotActive              = errors.New("project is not active")
	ErrProjectNotFound               = errors.New("project not found")
	ErrProjectNotPlanning            = errors.New("project is not in planning state")
	ErrResourceSpecificationNotFound = errors.New("resource specification not found")
	ErrTagAlreadyExists              = errors.New("tag already exists")
	ErrTagNotFound                   = errors.New("tag not found")
	ErrUnauthenticated               = errors.New("unauthenticated, please log in (again)")
	ErrUserNotFound                  = errors.New("user not found")
)
