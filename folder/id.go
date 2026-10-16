package folder

import (
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

// ValidID reports whether id is a folder ID as SafeID would produce it.
func ValidID(id string) bool {
	return id != "" && SafeID(id) == id
}
