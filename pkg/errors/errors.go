// Package errors provides custom error types and utilities for FleetTracker Pro.
// It implements a standardized error handling approach across all services.
package errors

import (
	"fmt"
	"net/http"
)

// AppError represents a standardized application error with HTTP status code and error code.
type AppError struct {
	Code       string `json:"code"`                 // Machine-readable error code
	Message    string `json:"message"`              // Human-readable error message
	Status     int    `json:"-"`                    // HTTP status code
	InternalErr error  `json:"-"`                    // Internal error (not exposed to client)
	Details    map[string]interface{} `json:"details,omitempty"` // Additional error details
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.InternalErr != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.InternalErr)
	}
	return e.Message
}

// Unwrap returns the internal error for error wrapping.
func (e *AppError) Unwrap() error {
	return e.InternalErr
}

// WithDetails adds additional details to the error.
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// WithInternal sets the internal error.
func (e *AppError) WithInternal(err error) *AppError {
	e.InternalErr = err
	return e
}

// Common error constructors

// NewNotFoundError creates a new not found error.
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s not found", resource),
		Status:  http.StatusNotFound,
	}
}

// NewUnauthorizedError creates a new unauthorized error.
func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "Unauthorized access"
	}
	return &AppError{
		Code:    "UNAUTHORIZED",
		Message: message,
		Status:  http.StatusUnauthorized,
	}
}

// NewForbiddenError creates a new forbidden error.
func NewForbiddenError(message string) *AppError {
	if message == "" {
		message = "Access forbidden"
	}
	return &AppError{
		Code:    "FORBIDDEN",
		Message: message,
		Status:  http.StatusForbidden,
	}
}

// NewValidationError creates a new validation error.
func NewValidationError(message string) *AppError {
	if message == "" {
		message = "Validation failed"
	}
	return &AppError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

// NewBadRequestError creates a new bad request error.
func NewBadRequestError(message string) *AppError {
	if message == "" {
		message = "Bad request"
	}
	return &AppError{
		Code:    "BAD_REQUEST",
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

// NewConflictError creates a new conflict error.
func NewConflictError(message string) *AppError {
	if message == "" {
		message = "Resource conflict"
	}
	return &AppError{
		Code:    "CONFLICT",
		Message: message,
		Status:  http.StatusConflict,
	}
}

// NewInternalError creates a new internal server error.
func NewInternalError(message string) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return &AppError{
		Code:    "INTERNAL_ERROR",
		Message: message,
		Status:  http.StatusInternalServerError,
	}
}

// NewTooManyRequestsError creates a new rate limit error.
func NewTooManyRequestsError(message string) *AppError {
	if message == "" {
		message = "Too many requests"
	}
	return &AppError{
		Code:    "TOO_MANY_REQUESTS",
		Message: message,
		Status:  http.StatusTooManyRequests,
	}
}

// NewServiceUnavailableError creates a new service unavailable error.
func NewServiceUnavailableError(message string) *AppError {
	if message == "" {
		message = "Service temporarily unavailable"
	}
	return &AppError{
		Code:    "SERVICE_UNAVAILABLE",
		Message: message,
		Status:  http.StatusServiceUnavailable,
	}
}

// Predefined common errors
var (
	// ErrNotFound is a generic not found error
	ErrNotFound = &AppError{
		Code:    "NOT_FOUND",
		Message: "Resource not found",
		Status:  http.StatusNotFound,
	}

	// ErrUnauthorized is a generic unauthorized error
	ErrUnauthorized = &AppError{
		Code:    "UNAUTHORIZED",
		Message: "Unauthorized access",
		Status:  http.StatusUnauthorized,
	}

	// ErrForbidden is a generic forbidden error
	ErrForbidden = &AppError{
		Code:    "FORBIDDEN",
		Message: "Access forbidden",
		Status:  http.StatusForbidden,
	}

	// ErrValidation is a generic validation error
	ErrValidation = &AppError{
		Code:    "VALIDATION_ERROR",
		Message: "Validation failed",
		Status:  http.StatusBadRequest,
	}

	// ErrBadRequest is a generic bad request error
	ErrBadRequest = &AppError{
		Code:    "BAD_REQUEST",
		Message: "Bad request",
		Status:  http.StatusBadRequest,
	}

	// ErrConflict is a generic conflict error
	ErrConflict = &AppError{
		Code:    "CONFLICT",
		Message: "Resource conflict",
		Status:  http.StatusConflict,
	}

	// ErrInternal is a generic internal server error
	ErrInternal = &AppError{
		Code:    "INTERNAL_ERROR",
		Message: "Internal server error",
		Status:  http.StatusInternalServerError,
	}

	// ErrTooManyRequests is a generic rate limit error
	ErrTooManyRequests = &AppError{
		Code:    "TOO_MANY_REQUESTS",
		Message: "Too many requests",
		Status:  http.StatusTooManyRequests,
	}

	// ErrServiceUnavailable is a generic service unavailable error
	ErrServiceUnavailable = &AppError{
		Code:    "SERVICE_UNAVAILABLE",
		Message: "Service temporarily unavailable",
		Status:  http.StatusServiceUnavailable,
	}
)

// IsAppError checks if an error is an AppError.
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError extracts AppError from error, or creates a generic internal error.
func GetAppError(err error) *AppError {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	// Wrap unknown errors as internal errors
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "Internal server error",
		Status:     http.StatusInternalServerError,
		InternalErr: err,
	}
}

// Wrap wraps an error with a message and converts it to AppError.
func Wrap(err error, message string) *AppError {
	if err == nil {
		return nil
	}

	if appErr, ok := err.(*AppError); ok {
		appErr.Message = message
		return appErr
	}

	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    message,
		Status:     http.StatusInternalServerError,
		InternalErr: err,
	}
}

// WrapWithCode wraps an error with a custom code and message.
func WrapWithCode(err error, code string, message string, status int) *AppError {
	if err == nil {
		return nil
	}

	return &AppError{
		Code:       code,
		Message:    message,
		Status:     status,
		InternalErr: err,
	}
}

