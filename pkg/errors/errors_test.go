package errors

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhenCheckingIfErrorIsContextError(t *testing.T) {
	t.Run("and error is context.Canceled", func(t *testing.T) {
		t.Run("it should return true", func(t *testing.T) {
			assert.True(t, IsContextError(context.Canceled))
		})
	})

	t.Run("and error is context.DeadlineExceeded", func(t *testing.T) {
		t.Run("it should return true", func(t *testing.T) {
			assert.True(t, IsContextError(context.DeadlineExceeded))
		})
	})

	t.Run("and error is ErrContextCancelled", func(t *testing.T) {
		t.Run("it should return true", func(t *testing.T) {
			assert.True(t, IsContextError(ErrContextCancelled))
		})
	})

	t.Run("and error is other error", func(t *testing.T) {
		t.Run("it should return false", func(t *testing.T) {
			otherError := errors.New("some other error")
			assert.False(t, IsContextError(otherError))
		})
	})

	t.Run("and error is nil", func(t *testing.T) {
		t.Run("it should return false", func(t *testing.T) {
			assert.False(t, IsContextError(nil))
		})
	})
}

func TestWhenCheckingContext(t *testing.T) {
	t.Run("and context is cancelled", func(t *testing.T) {
		t.Run("it should return ErrContextCancelled", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := CheckContext(ctx)
			assert.ErrorIs(t, err, ErrContextCancelled)
		})
	})

	t.Run("and context is not cancelled", func(t *testing.T) {
		t.Run("it should return nil", func(t *testing.T) {
			ctx := context.Background()

			err := CheckContext(ctx)
			assert.NoError(t, err)
		})
	})
}
