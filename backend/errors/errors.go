package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError represents a structured API error
type APIError struct {
	Code     int    `json:"-"`       // HTTP status code
	Message  string `json:"message"` // User-friendly error message
	Detail   string `json:"detail"`  // Detailed error message
	Type     string `json:"type"`    // Error type for client handling
	Field    string `json:"field"`   // Field name for validation errors
	Internal error  `json:"-"`       // Internal error (not exposed)
}

func (e *APIError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

// Common error types
const (
	TypeValidation     = "VALIDATION_ERROR"
	TypeAuthentication = "AUTHENTICATION_ERROR"
	TypeAuthorization  = "AUTHORIZATION_ERROR"
	TypeNotFound       = "NOT_FOUND"
	TypeBadRequest     = "BAD_REQUEST"
	TypeInternal       = "INTERNAL_ERROR"
)

// Error constructors
func NewValidationError(field, message string) *APIError {
	return &APIError{
		Code:    http.StatusBadRequest,
		Message: "Validation failed",
		Detail:  message,
		Type:    TypeValidation,
		Field:   field,
	}
}

func NewAuthenticationError(message string) *APIError {
	return &APIError{
		Code:    http.StatusUnauthorized,
		Message: message,
		Type:    TypeAuthentication,
	}
}

func NewAuthorizationError(message string) *APIError {
	return &APIError{
		Code:    http.StatusForbidden,
		Message: message,
		Type:    TypeAuthorization,
	}
}

func NewNotFoundError(resource string) *APIError {
	return &APIError{
		Code:    http.StatusNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Type:    TypeNotFound,
	}
}

func NewInternalError(err error) *APIError {
	return &APIError{
		Code:     http.StatusInternalServerError,
		Message:  "An internal error occurred",
		Type:     TypeInternal,
		Internal: err,
	}
}

// Error checking helpers
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Type == TypeNotFound
	}
	return false
}

func IsValidationError(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Type == TypeValidation
	}
	return false
}
