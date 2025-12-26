package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeFilename(t *testing.T) {
	t.Run("when filename is clean", func(t *testing.T) {
		t.Run("it should return the filename unchanged", func(t *testing.T) {
			result := SanitizeFilename("document.pdf")
			assert.Equal(t, "document.pdf", result)
		})
	})

	t.Run("when filename contains path separators", func(t *testing.T) {
		t.Run("and contains forward slashes", func(t *testing.T) {
			t.Run("it should strip the directory path", func(t *testing.T) {
				result := SanitizeFilename("../../etc/passwd")
				assert.Equal(t, "passwd", result)
			})
		})

		t.Run("and contains backslashes", func(t *testing.T) {
			t.Run("it should strip the directory path", func(t *testing.T) {
				result := SanitizeFilename("..\\..\\Windows\\System32\\config")
				assert.Equal(t, "config", result)
			})
		})

		t.Run("and contains mixed separators", func(t *testing.T) {
			t.Run("it should strip the directory path", func(t *testing.T) {
				result := SanitizeFilename("../path\\to/file.txt")
				assert.Equal(t, "file.txt", result)
			})
		})
	})

	t.Run("when filename is too long", func(t *testing.T) {
		t.Run("and exceeds 255 characters", func(t *testing.T) {
			t.Run("it should truncate to 255 characters", func(t *testing.T) {
				longName := string(make([]byte, 300))
				for i := range longName {
					longName = longName[:i] + "a"
				}
				longName += ".txt"
				result := SanitizeFilename(longName)
				assert.LessOrEqual(t, len(result), 255)
			})

			t.Run("it should preserve the file extension", func(t *testing.T) {
				longName := ""
				for i := 0; i < 300; i++ {
					longName += "a"
				}
				longName += ".pdf"
				result := SanitizeFilename(longName)
				assert.Equal(t, ".pdf", result[len(result)-4:])
			})
		})
	})

	t.Run("when filename is empty", func(t *testing.T) {
		t.Run("it should return a default filename", func(t *testing.T) {
			result := SanitizeFilename("")
			assert.Equal(t, "untitled", result)
		})
	})

	t.Run("when filename contains only path separators", func(t *testing.T) {
		t.Run("it should return a default filename", func(t *testing.T) {
			result := SanitizeFilename("../../")
			assert.Equal(t, "untitled", result)
		})
	})
}
