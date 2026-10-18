package folder

import (
	"path/filepath"
	"regexp"
	"strings"
)

var unsafeID = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// SafeID turns a directory name into a URL safe folder ID.
func SafeID(name string) string {
	id := strings.Trim(unsafeID.ReplaceAllString(name, "-"), "-.")
	if id == "" {
		return "default"
	}
	return id
}

// ValidID reports whether id is a folder ID: segments as SafeID makes
// them, joined with slashes for a nested folder.
func ValidID(id string) bool {
	if id == "" {
		return false
	}
	for _, segment := range strings.Split(id, "/") {
		if segment == "" || SafeID(segment) != segment {
			return false
		}
	}
	return true
}

// PathID turns a path relative to the root into a folder ID, segment by
// segment.
func PathID(rel string) string {
	var segments []string
	for _, segment := range strings.Split(filepath.ToSlash(rel), "/") {
		if segment == "" || segment == "." {
			continue
		}
		segments = append(segments, SafeID(segment))
	}
	if len(segments) == 0 {
		return "default"
	}
	return strings.Join(segments, "/")
}
