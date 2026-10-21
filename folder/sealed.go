package folder

import (
	"path"
	"strings"
)

// sealedNames are the entries task-ui lists but never opens: the git
// repository and the files that hold credentials or a web server's
// access rules. A name that starts with a prefix in sealedPrefixes
// counts too.
var (
	sealedNames    = []string{".git", ".env", ".htpasswd", ".htaccess"}
	sealedPrefixes = []string{".env."}
)

// Sealed reports whether an entry of a checkout is one task-ui lists but
// does not open, show or go into: .git, .env and .env.*, .htpasswd and
// .htaccess.
func Sealed(name string) bool {
	for _, sealed := range sealedNames {
		if name == sealed {
			return true
		}
	}
	for _, prefix := range sealedPrefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

// SealedPath reports whether a slash path names a sealed entry or goes
// through one: .git/config and sub/.env.local are both sealed.
func SealedPath(rel string) bool {
	for _, segment := range strings.Split(path.Clean("/"+rel), "/") {
		if Sealed(segment) {
			return true
		}
	}
	return false
}
