package services

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
