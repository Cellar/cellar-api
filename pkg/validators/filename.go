package validators

import (
	"path/filepath"
	"strings"
)

func SanitizeFilename(filename string) string {
	if filename == "" {
		return "untitled"
	}

	filename = filepath.Base(filename)

	filename = strings.ReplaceAll(filename, "\\", "/")
	filename = filepath.Base(filename)

	if filename == "" || filename == "." || filename == ".." {
		return "untitled"
	}

	if len(filename) > 255 {
		ext := filepath.Ext(filename)
		nameWithoutExt := filename[:len(filename)-len(ext)]
		maxNameLen := 255 - len(ext)
		if maxNameLen > 0 {
			filename = nameWithoutExt[:maxNameLen] + ext
		} else {
			filename = filename[:255]
		}
	}

	return filename
}
