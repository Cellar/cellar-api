package errors

import (
	"context"
	"errors"
)

// ErrContextCancelled is returned when an operation is cancelled due to context cancellation
var ErrContextCancelled = errors.New("operation cancelled due to context cancellation")

// IsContextError checks if an error is due to context cancellation or timeout
func IsContextError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, ErrContextCancelled)
}

// CheckContext checks if context is done and returns appropriate error
func CheckContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ErrContextCancelled
	default:
		return nil
	}
}

// ValidationError represents an error caused by invalid input or business rule violation
type ValidationError struct {
	message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return e.message
}

// NewValidationError creates a new validation error with the given message
func NewValidationError(msg string) error {
	return &ValidationError{message: msg}
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	if err == nil {
		return false
	}
	var ve *ValidationError
	return errors.As(err, &ve)
}

// FileTooLargeError represents an error caused by a file that exceeds the maximum allowed size
type FileTooLargeError struct {
	message string
}

// Error implements the error interface
func (e *FileTooLargeError) Error() string {
	return e.message
}

// NewFileTooLargeError creates a new file too large error with the given message
func NewFileTooLargeError(msg string) error {
	return &FileTooLargeError{message: msg}
}

// IsFileTooLargeError checks if an error is a file too large error
func IsFileTooLargeError(err error) bool {
	if err == nil {
		return false
	}
	var fe *FileTooLargeError
	return errors.As(err, &fe)
}

// RateLimitError represents an error caused by exceeding rate limits
type RateLimitError struct {
	message    string
	retryAfter int
}

// Error implements the error interface
func (e *RateLimitError) Error() string {
	return e.message
}

// RetryAfter returns the number of seconds until the rate limit resets
func (e *RateLimitError) RetryAfter() int {
	return e.retryAfter
}

// NewRateLimitError creates a new rate limit error with the given message and retry-after value
func NewRateLimitError(msg string, retryAfter int) error {
	return &RateLimitError{
		message:    msg,
		retryAfter: retryAfter,
	}
}

// IsRateLimitError checks if an error is a rate limit error
func IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	var re *RateLimitError
	return errors.As(err, &re)
}

// GetRateLimitError attempts to extract a RateLimitError from an error chain
func GetRateLimitError(err error) *RateLimitError {
	if err == nil {
		return nil
	}
	var re *RateLimitError
	if errors.As(err, &re) {
		return re
	}
	return nil
}
