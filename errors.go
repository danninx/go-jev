package jev

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthorized        = errors.New("unauthorized: missing or invalid API key")
	ErrUnprocessableEntity = errors.New("unprocessable entity: request validation failed")
	ErrRateLimited         = errors.New("rate limited: too many requests")
	ErrOverloaded          = errors.New("overloaded: service temporarily unavailable")
)

type APIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message,omitempty"`
	Detail     any    `json:"detail,omitempty"`
	Err        error  `json:"-"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("jev api error (status %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("jev api error (status %d)", e.StatusCode)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

func NewAPIError(statusCode int, message string, detail any) *APIError {
	var sentinel error
	switch statusCode {
	case 401:
		sentinel = ErrUnauthorized
	case 422:
		sentinel = ErrUnprocessableEntity
	case 429:
		sentinel = ErrRateLimited
	case 529:
		sentinel = ErrOverloaded
	}

	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Detail:     detail,
		Err:        sentinel,
	}
}
