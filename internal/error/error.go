package error

import (
	"errors"
)

var (
	ErrInternal          = errors.New("internal error")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrPageNotFound      = errors.New("page not found")
	ErrMethodNotAllowed  = errors.New("method not allowed")
	ErrValidationProcess = errors.New("validation error")
	ErrTooManyRequests   = errors.New("too many requests")
	ErrBadRequest        = errors.New("bad request")

	// User Domain error
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")

	// API-key Domain error
	ErrAPIKeyAlreadyExists = errors.New("API-key already exists")
	ErrAPIKeyNotFound      = errors.New("API key not found")

	// Paste Domain error
	ErrPasteNotFound      = errors.New("paste not found")
	ErrPasteAlreadyExists = errors.New("paste already exists")
)

type BaseError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (r BaseError) Error() string {
	return r.Message
}

func (r BaseError) Code() int {
	return r.Status
}
