package errors

import (
	"context"
	"errors"
	"testing"
)

func TestWhenCheckingIfErrorIsContextError(t *testing.T) {
	t.Run("and error is context.Canceled", func(t *testing.T) {
		t.Run("it should return true", func(t *testing.T) {
			if !IsContextError(context.Canceled) {
				t.Error("expected IsContextError to return true for context.Canceled")
			}
		})
	})

	t.Run("and error is context.DeadlineExceeded", func(t *testing.T) {
		t.Run("it should return true", func(t *testing.T) {
			if !IsContextError(context.DeadlineExceeded) {
				t.Error("expected IsContextError to return true for context.DeadlineExceeded")
			}
		})
	})

	t.Run("and error is ErrContextCancelled", func(t *testing.T) {
		t.Run("it should return true", func(t *testing.T) {
			if !IsContextError(ErrContextCancelled) {
				t.Error("expected IsContextError to return true for ErrContextCancelled")
			}
		})
	})

	t.Run("and error is other error", func(t *testing.T) {
		t.Run("it should return false", func(t *testing.T) {
			otherError := errors.New("some other error")
			if IsContextError(otherError) {
				t.Error("expected IsContextError to return false for other errors")
			}
		})
	})

	t.Run("and error is nil", func(t *testing.T) {
		t.Run("it should return false", func(t *testing.T) {
			if IsContextError(nil) {
				t.Error("expected IsContextError to return false for nil")
			}
		})
	})
}

func TestWhenCheckingContext(t *testing.T) {
	t.Run("and context is cancelled", func(t *testing.T) {
		t.Run("it should return ErrContextCancelled", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := CheckContext(ctx)
			if !errors.Is(err, ErrContextCancelled) {
				t.Errorf("expected ErrContextCancelled, got %v", err)
			}
		})
	})

	t.Run("and context is not cancelled", func(t *testing.T) {
		t.Run("it should return nil", func(t *testing.T) {
			ctx := context.Background()

			err := CheckContext(ctx)
			if err != nil {
				t.Errorf("expected nil, got %v", err)
			}
		})
	})
}
