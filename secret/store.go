package secret

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/dotenv"
)

// Store keeps the secrets as dotenv files, one per scope:
// <root>/all.env for every folder, <root>/<folder>.env for one.
type Store struct {
	Root string
}

func (s *Store) path(scope string) string {
	return filepath.Join(s.Root, scope+".env")
}

// List returns the secrets of every scope, all first, then by folder and
// key. Values are filled in.
func (s *Store) List() []Secret {
	files, _ := filepath.Glob(filepath.Join(s.Root, "*.env"))
	result := []Secret{}
	for _, file := range files {
		scope := strings.TrimSuffix(filepath.Base(file), ".env")
		vars, _ := dotenv.Read(file)
		for _, key := range dotenv.Keys(vars) {
			result = append(result, Secret{Scope: scope, Key: key, Value: vars[key]})
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Scope != result[j].Scope {
			if result[i].Scope == ScopeAll {
				return true
			}
			if result[j].Scope == ScopeAll {
				return false
			}
			return result[i].Scope < result[j].Scope
		}
		return result[i].Key < result[j].Key
	})
	return result
}

// Get returns one secret, named by scope and key.
func (s *Store) Get(ref Secret) (Secret, bool) {
	if !ValidScope(ref.Scope) {
		return Secret{}, false
	}
	vars, _ := dotenv.Read(s.path(ref.Scope))
	value, ok := vars[ref.Key]
	return Secret{Scope: ref.Scope, Key: ref.Key, Value: value}, ok
}

// Set stores a secret, replacing a value under the same key.
func (s *Store) Set(secret Secret) error {
	if !ValidScope(secret.Scope) {
		return fmt.Errorf("invalid scope %q", secret.Scope)
	}
	if !ValidKey(secret.Key) {
		return ErrInvalidKey
	}
	path := s.path(secret.Scope)
	vars, _ := dotenv.Read(path)
	if vars == nil {
		vars = map[string]string{}
	}
	vars[secret.Key] = secret.Value
	return dotenv.Write(path, vars)
}

// Delete removes a secret, named by scope and key.
func (s *Store) Delete(ref Secret) error {
	if !ValidScope(ref.Scope) {
		return fmt.Errorf("invalid scope %q", ref.Scope)
	}
	path := s.path(ref.Scope)
	vars, err := dotenv.Read(path)
	if err != nil {
		return err
	}
	if _, ok := vars[ref.Key]; !ok {
		return fmt.Errorf("no secret %s in %s", ref.Key, ref.Scope)
	}
	delete(vars, ref.Key)
	if len(vars) == 0 {
		return os.Remove(path)
	}
	return dotenv.Write(path, vars)
}

// For returns the secrets a folder's runs receive: the shared ones, then
// the folder's own, which win.
func (s *Store) For(folderID string) map[string]string {
	result := map[string]string{}
	for _, scope := range []string{ScopeAll, folderID} {
		vars, _ := dotenv.Read(s.path(scope))
		for key, value := range vars {
			result[key] = value
		}
	}
	return result
}

// Scoped returns the secrets that reach a folder, as the list shows them.
func (s *Store) Scoped(folderID string) []Secret {
	result := []Secret{}
	for _, secret := range s.List() {
		if secret.Scope == ScopeAll || secret.Scope == folderID {
			result = append(result, secret)
		}
	}
	return result
}
