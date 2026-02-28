package error

import (
	"errors"
)

var (
	InternalError          = errors.New("internal error")
	Unauthorized           = errors.New("unauthorized")
	PageNotFound           = errors.New("page not found")
	MethodNotAllowed       = errors.New("method not allowed")
	ValidationProcessError = errors.New("validation error")

	// User Domain error
	UserAlreadyExists = errors.New("user already exists")
	UserNotFound      = errors.New("user not found")
	APIKeyNotFound    = errors.New("API key not found")
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
