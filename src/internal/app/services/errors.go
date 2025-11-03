package services

import "errors"

var (
	ErrDatabaseError = errors.New("database error")

	ErrUnauthenticated  = errors.New("unauthenticated, please log in (again)")
	ErrItemTypeNotFound = errors.New("item type not found")

	// WARN(noatu): claims that something exists, use *NotFound
	// ErrForbidden     = errors.New("forbidden")
)

type ErrNotImplemented struct {
	MethodName string
}

func (e *ErrNotImplemented) Error() string {
	if e.MethodName != "" {
		return e.MethodName + " is not yet implemented"
	}
	return "method is not yet implemented"
}

func NewErrNotImplemented(method string) *ErrNotImplemented {
	return &ErrNotImplemented{MethodName: method}
}
