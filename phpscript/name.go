package phpscript

import (
	"fmt"
	"strings"
)

// Prefix is where every module is served from. A path under it names the
// module in its first two segments and belongs to the module in the rest.
const Prefix = "/m"

// ParseName splits a module's name into its vendor and its plugin.
//
// Both halves are lower case, and each one holds letters, digits, a dash, an
// underscore or a dot: the name reaches a URL, a table row and a directory,
// and these are the characters all three read the same. A name with a slash
// too few or too many, an empty half or a dot segment is refused.
func ParseName(name string) (string, string, error) {
	vendor, plugin, found := strings.Cut(strings.TrimSpace(name), "/")
	if !found {
		return "", "", fmt.Errorf("%w: %q holds no slash", ErrName, name)
	}
	for _, half := range []string{vendor, plugin} {
		if err := checkHalf(name, half); err != nil {
			return "", "", err
		}
	}
	return vendor, plugin, nil
}

// checkHalf refuses a half of a name that a URL, a row or a directory would
// read differently.
func checkHalf(name, half string) error {
	switch {
	case half == "":
		return fmt.Errorf("%w: %q has an empty half", ErrName, name)
	case half == "." || half == "..":
		return fmt.Errorf("%w: %q names a directory above itself", ErrName, name)
	case strings.ContainsAny(half, "/"):
		return fmt.Errorf("%w: %q holds more than one slash", ErrName, name)
	}
	for _, r := range half {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
		case r == '-', r == '_', r == '.':
		default:
			return fmt.Errorf("%w: %q holds %q, which is not a lower case letter, a digit, a dash, an underscore or a dot", ErrName, name, r)
		}
	}
	return nil
}
