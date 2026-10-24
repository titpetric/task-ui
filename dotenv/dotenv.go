// Package dotenv writes KEY=VALUE lines the way atkins and Taskfiles
// read them.
package dotenv

import (
	"sort"
	"strings"
)

// Format renders the variables as dotenv lines, keys sorted.
func Format(vars map[string]string) string {
	var b strings.Builder
	for _, key := range Keys(vars) {
		b.WriteString(key + "=" + Quote(vars[key]) + "\n")
	}
	return b.String()
}

// Quote wraps a value in quotes when a shell would read a bare one differently:
// spaces, a quote, a comment character, or nothing at all.
func Quote(value string) string {
	switch {
	case value != "" && !strings.ContainsAny(value, " \t#\"'"):
		return value
	case !strings.Contains(value, `"`):
		return `"` + value + `"`
	case !strings.Contains(value, `'`):
		return `'` + value + `'`
	}
	return value
}

// Keys returns the variable names, sorted.
func Keys(vars map[string]string) []string {
	keys := make([]string, 0, len(vars))
	for key := range vars {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
