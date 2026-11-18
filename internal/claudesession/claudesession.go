// Package claudesession names the directory a Claude Code conversation keeps its transcript in.
package claudesession

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/titpetric/task-ui/app/model"
)

// Name returns the directory name of a user's session: the user's name, made safe for a path, then the session ID.
func Name(user, id string) string {
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			return r
		}
		return '_'
	}, user)
	safe = strings.Trim(safe, ".")
	if safe == "" {
		safe = "user"
	}
	return safe + "_" + id
}

// Dir returns the directory of a user's session under root, empty without a root or an ID.
func Dir(root, user, id string) string {
	if root == "" || id == "" {
		return ""
	}
	return filepath.Join(root, Name(user, id))
}

// ID returns the session ID a directory name holds, and whether the name holds one.
func ID(name string) (string, bool) {
	at := strings.LastIndex(name, "_")
	if at < 0 {
		return "", false
	}
	id := name[at+1:]
	if !model.ValidID(id) {
		return "", false
	}
	return id, true
}

// UUID returns the UUID Claude Code names the session by: a name flag takes a UUID, not a ULID.
func UUID(id string) string {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("task-ui/claude/"+id)).String()
}

// Remove deletes a session directory, and answers whether it was there.
func Remove(dir string) error {
	if dir == "" {
		return nil
	}
	err := os.RemoveAll(dir)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// Clean removes every session directory under root whose ID keep does not name.
func Clean(root string, keep map[string]bool) error {
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id, ok := ID(entry.Name())
		if !ok || keep[id] {
			continue
		}
		if err := os.RemoveAll(filepath.Join(root, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}
