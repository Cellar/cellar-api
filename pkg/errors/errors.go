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
