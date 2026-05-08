package service

import "errors"

var (
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrUnauthenticated    = errors.New("unauthenticated")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrNotFound           = errors.New("not found")
	ErrInternal           = errors.New("internal")
	ErrUnavailable        = errors.New("unavailable")
	ErrFailedPrecondition = errors.New("failed precondition")
)

type Error struct {
	Kind    error
	Message string
}

func (e Error) Error() string {
	return e.Message
}

func (e Error) Unwrap() error {
	return e.Kind
}

func serviceError(kind error, message string) error {
	return Error{Kind: kind, Message: message}
}
