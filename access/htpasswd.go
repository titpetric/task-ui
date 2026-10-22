package access

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ParseHtpasswd reads a password file in the htpasswd format, one
// "name:hash" per line, blank lines and # comments skipped. It accepts only
// bcrypt hashes, as `htpasswd -B` and `task-ui hash` make them:
// the other htpasswd formats are too weak to keep.
func ParseHtpasswd(r io.Reader) ([]Credential, error) {
	var result []Credential
	seen := map[string]bool{}
	scanner := bufio.NewScanner(r)
	for n := 1; scanner.Scan(); n++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, hash, found := strings.Cut(line, ":")
		name, hash = strings.TrimSpace(name), strings.TrimSpace(hash)
		if !found || name == "" || hash == "" {
			return nil, fmt.Errorf("line %d: expected name:hash", n)
		}
		if !Hashed(hash) {
			return nil, fmt.Errorf("line %d: user %s: only bcrypt hashes are read, make one with htpasswd -B or task-ui hash", n, name)
		}
		if seen[name] {
			return nil, fmt.Errorf("line %d: user %s is listed twice", n, name)
		}
		seen[name] = true
		result = append(result, Credential{Name: name, Hash: hash})
	}
	return result, scanner.Err()
}

// LoadHtpasswd reads a password file.
func LoadHtpasswd(path string) ([]Credential, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	creds, err := ParseHtpasswd(file)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return creds, nil
}

// SetHtpasswd writes a user's hash into a password file: it replaces
// the user's line when there, adds one when not, and creates a missing
// file. Other lines stay as they are.
func SetHtpasswd(path, name, hash string) error {
	name = strings.TrimSpace(name)
	if name == "" || strings.ContainsAny(name, ":\n") {
		return errors.New("a user name may not be empty or hold a colon")
	}
	if !Hashed(hash) {
		return errors.New("the hash is not a bcrypt hash")
	}
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	var lines []string
	replaced := false
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if existing, _, _ := strings.Cut(strings.TrimSpace(line), ":"); strings.TrimSpace(existing) == name && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			if !replaced {
				lines = append(lines, Credential{Name: name, Hash: hash}.Line())
				replaced = true
			}
			continue
		}
		if line != "" || len(data) > 0 {
			lines = append(lines, line)
		}
	}
	if !replaced {
		lines = append(lines, Credential{Name: name, Hash: hash}.Line())
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600)
}

// RemoveHtpasswd removes a user's line from a password file, and
// reports whether the file held one. Other lines stay as they are; a
// missing file holds nobody.
func RemoveHtpasswd(path, name string) (bool, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return false, errors.New("a user name may not be empty")
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var lines []string
	removed := false
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		trimmed := strings.TrimSpace(line)
		if existing, _, _ := strings.Cut(trimmed, ":"); strings.TrimSpace(existing) == name && !strings.HasPrefix(trimmed, "#") {
			removed = true
			continue
		}
		lines = append(lines, line)
	}
	if !removed {
		return false, nil
	}
	text := strings.Join(lines, "\n")
	if text != "" {
		text += "\n"
	}
	return true, os.WriteFile(path, []byte(text), 0o600)
}
