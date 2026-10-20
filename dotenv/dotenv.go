// Package dotenv reads and writes KEY=VALUE files the way atkins and
// Taskfiles read them.
package dotenv

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var validKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Read reads a dotenv file: comments and blank lines skipped, quotes around
// a value removed, `export` prefixes ignored.
func Read(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return Parse(file)
}

// Parse reads dotenv content from a reader.
func Parse(r io.Reader) (map[string]string, error) {
	vars := map[string]string{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if !validKey.MatchString(key) {
			continue
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
			value = value[1 : len(value)-1]
		}
		vars[key] = value
	}
	return vars, scanner.Err()
}

// Write writes the variables as a dotenv file, keys sorted, readable by
// the owner alone. Write creates the directory.
func Write(path string, vars map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(Format(vars)), 0o600)
}

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
