package services

import "errors"

var (
	ErrDatabaseError = errors.New("database error")

	ErrUnauthenticated  = errors.New("unauthenticated, please log in (again)")
	ErrItemTypeNotFound = errors.New("item type not found")
	ErrItemNotFound     = errors.New("item not found")
	ErrTagNotFound      = errors.New("tag not found")
)

type ErrNotImplemented struct {
	MethodName string
}
func (e *ErrNotImplemented) Error() string {
	if e.MethodName != "" {
		return e.MethodName + " is not yet implemented"
	}
	return "Method is not yet implemented"
}
func NewErrNotImplemented(method string) *ErrNotImplemented {
	return &ErrNotImplemented{MethodName: method}
}
