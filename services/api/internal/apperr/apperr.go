package apperr

import "errors"

// Sentinel errors mapped to HTTP by handlers.
var (
	ErrNotFound   = errors.New("resource not found")
	ErrConflict   = errors.New("resource conflict")
	ErrValidation = errors.New("validation failed")
	ErrInvalid    = errors.New("invalid request")
	ErrForbidden  = errors.New("forbidden")
)
