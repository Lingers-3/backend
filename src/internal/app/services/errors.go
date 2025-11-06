package services

import "errors"

var (
	ErrNotImplemented = errors.New("method is not yet implemented")

	ErrDatabaseError = errors.New("database error")

	ErrUnauthenticated    = errors.New("unauthenticated, please log in (again)")
	ErrItemTypeNotFound   = errors.New("item type not found")
	ErrItemNotFound       = errors.New("item not found")
	ErrTagNotFound        = errors.New("tag not found")
	ErrTagAlreadyExists   = errors.New("tag already exists")
	ErrForeignKeyViolated = errors.New("resource is still in use")

	// WARN(noatu): claims that something exists, use *NotFound
	// ErrForbidden     = errors.New("forbidden")
)
