package gater

import (
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// hasInvalidPath checks if the given path contains "." or ".." as path segments.
func hasInvalidPath(path string) bool {
	path = filepath.ToSlash(strings.TrimSpace(path))
	for _, p := range strings.Split(path, "/") {
		// Check for special characters "." and ".." which have special meaning in file systems.
		// In object storage systems, these characters should not be used as they can cause confusion.
		switch strings.TrimSpace(p) {
		case ".":
			return true
		case "..":
			return true
		}
	}
	return false
}

// checkValidObjectPrefix checks if the given object prefix is valid:
// - does not have invalid path segments
// - is a valid UTF-8 string
// - does not contain double slashes "//"
func checkValidObjectPrefix(prefix string) bool {
	if hasInvalidPath(prefix) {
		return false
	}
	if !utf8.ValidString(prefix) {
		return false
	}
	if strings.Contains(prefix, `//`) {
		return false
	}
	return true
}
