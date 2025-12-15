package services

import "errors"

// WARN(pencelheimer): use *NotFound instead of Forbidden
var (
	// NOTE(pencelheimer): Remember to sort this list
	ErrDatabaseError                 = errors.New("database error")
	ErrFileSystemError               = errors.New("file system error")
	ErrForeignKeyViolated            = errors.New("resource is still in use")
	ErrImageTooLarge                 = errors.New("image is too large")
	ErrInsufficientResources         = errors.New("insufficient resources for specific item")
	ErrInvalidImageFormat            = errors.New("invalid image format")
	ErrInvalidQuantity               = errors.New("used quantity cannot exceed reserved quantity")
	ErrItemMismatch                  = errors.New("item does not belong to the specified item type")
	ErrItemNotFound                  = errors.New("item not found")
	ErrItemTypeNotFound              = errors.New("item type not found")
	ErrNotImplemented                = errors.New("method is not yet implemented")
	ErrPictureNotFound               = errors.New("picture not found")
	ErrProjectAlreadyActive          = errors.New("project is already active or completed")
	ErrProjectAlreadyExists          = errors.New("project with this name already exists")
	ErrProjectNotActive              = errors.New("project is not active")
	ErrProjectNotFound               = errors.New("project not found")
	ErrProjectNotPlanning            = errors.New("project is not in planning state")
	ErrResourceReservationNotFound   = errors.New("resource reservation not found")
	ErrResourceSpecificationNotFound = errors.New("resource specification not found")
	ErrTagAlreadyExists              = errors.New("tag already exists")
	ErrTagNotFound                   = errors.New("tag not found")
	ErrTemplateNotFound              = errors.New("template not found")
	ErrTemplateNameAlreadyExists     = errors.New("template name already exists")
	ErrTemplateResourceAlreadyExists = errors.New("resource specification for this item type already exists in the template")
	ErrTemplateServiceInvalidSort    = errors.New("invalid sorting attribute or direction")
	ErrUnauthenticated               = errors.New("unauthenticated, please log in (again)")
	ErrUserNotFound                  = errors.New("user not found")
)
