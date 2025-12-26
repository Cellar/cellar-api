package validators

import (
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	t.Run("when filename is clean", func(t *testing.T) {
		t.Run("it should return the filename unchanged", func(t *testing.T) {
			result := SanitizeFilename("document.pdf")
			expected := "document.pdf"
			if result != expected {
				t.Errorf("expected %q, got %q", expected, result)
			}
		})
	})

	t.Run("when filename contains path separators", func(t *testing.T) {
		t.Run("and contains forward slashes", func(t *testing.T) {
			t.Run("it should strip the directory path", func(t *testing.T) {
				result := SanitizeFilename("../../etc/passwd")
				expected := "passwd"
				if result != expected {
					t.Errorf("expected %q, got %q", expected, result)
				}
			})
		})

		t.Run("and contains backslashes", func(t *testing.T) {
			t.Run("it should strip the directory path", func(t *testing.T) {
				result := SanitizeFilename("..\\..\\Windows\\System32\\config")
				expected := "config"
				if result != expected {
					t.Errorf("expected %q, got %q", expected, result)
				}
			})
		})

		t.Run("and contains mixed separators", func(t *testing.T) {
			t.Run("it should strip the directory path", func(t *testing.T) {
				result := SanitizeFilename("../path\\to/file.txt")
				expected := "file.txt"
				if result != expected {
					t.Errorf("expected %q, got %q", expected, result)
				}
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
				if len(result) > 255 {
					t.Errorf("expected length <= 255, got %d", len(result))
				}
			})

			t.Run("it should preserve the file extension", func(t *testing.T) {
				longName := ""
				for i := 0; i < 300; i++ {
					longName += "a"
				}
				longName += ".pdf"
				result := SanitizeFilename(longName)
				if result[len(result)-4:] != ".pdf" {
					t.Errorf("expected extension .pdf, got %q", result[len(result)-4:])
				}
			})
		})
	})

	t.Run("when filename is empty", func(t *testing.T) {
		t.Run("it should return a default filename", func(t *testing.T) {
			result := SanitizeFilename("")
			expected := "untitled"
			if result != expected {
				t.Errorf("expected %q, got %q", expected, result)
			}
		})
	})

	t.Run("when filename contains only path separators", func(t *testing.T) {
		t.Run("it should return a default filename", func(t *testing.T) {
			result := SanitizeFilename("../../")
			expected := "untitled"
			if result != expected {
				t.Errorf("expected %q, got %q", expected, result)
			}
		})
	})
}
