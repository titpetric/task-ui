package repository

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/server/model"
)

// Env is the environment store: the secrets task-ui adds to runs, and the
// dotenv files it manages for the folders.
//
//	<root>/secrets/all.env           secrets for every folder
//	<root>/secrets/<folder>.env      secrets for one folder
//	<root>/files/<folder>/<path>     the managed content of a folder's env file
//
// Before a run, each env file the pipeline reads is written into the folder:
// the managed content (or the file already there), then the secrets. The
// pipeline file itself is never written.
type Env struct {
	Root string
}

// ScopeAll is the secret scope that applies to every folder.
const ScopeAll = "all"

var validKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ErrInvalidKey is returned for a variable name a shell would not accept.
var ErrInvalidKey = errors.New("a variable name is letters, digits and underscores, and does not start with a digit")

// ValidKey reports whether key is a variable name.
func ValidKey(key string) bool {
	return validKey.MatchString(key)
}

// ValidScope reports whether scope is "all" or a folder ID.
func ValidScope(scope string) bool {
	return scope == ScopeAll || (scope != "" && SafeID(scope) == scope)
}

func (e *Env) secretsPath(scope string) string {
	return filepath.Join(e.Root, "secrets", scope+".env")
}

// Secrets lists the secrets of every scope, all first, then by folder and
// key. Values are filled in.
func (e *Env) Secrets() []model.Secret {
	files, _ := filepath.Glob(filepath.Join(e.Root, "secrets", "*.env"))
	result := []model.Secret{}
	for _, file := range files {
		scope := strings.TrimSuffix(filepath.Base(file), ".env")
		vars, _ := ReadDotenv(file)
		for _, key := range sortedKeys(vars) {
			result = append(result, model.Secret{Scope: scope, Key: key, Value: vars[key]})
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

// Secret returns one secret.
func (e *Env) Secret(scope, key string) (model.Secret, bool) {
	if !ValidScope(scope) {
		return model.Secret{}, false
	}
	vars, _ := ReadDotenv(e.secretsPath(scope))
	value, ok := vars[key]
	return model.Secret{Scope: scope, Key: key, Value: value}, ok
}

// SetSecret stores a secret, replacing a value under the same key.
func (e *Env) SetSecret(secret model.Secret) error {
	if !ValidScope(secret.Scope) {
		return fmt.Errorf("invalid scope %q", secret.Scope)
	}
	if !ValidKey(secret.Key) {
		return ErrInvalidKey
	}
	path := e.secretsPath(secret.Scope)
	vars, _ := ReadDotenv(path)
	if vars == nil {
		vars = map[string]string{}
	}
	vars[secret.Key] = secret.Value
	return writeDotenv(path, vars)
}

// DeleteSecret removes a secret.
func (e *Env) DeleteSecret(scope, key string) error {
	if !ValidScope(scope) {
		return fmt.Errorf("invalid scope %q", scope)
	}
	path := e.secretsPath(scope)
	vars, err := ReadDotenv(path)
	if err != nil {
		return err
	}
	if _, ok := vars[key]; !ok {
		return fmt.Errorf("no secret %s in %s", key, scope)
	}
	delete(vars, key)
	if len(vars) == 0 {
		return os.Remove(path)
	}
	return writeDotenv(path, vars)
}

// SecretsFor returns the secrets a folder's runs receive: the shared ones,
// then the folder's own, which win.
func (e *Env) SecretsFor(project string) map[string]string {
	result := map[string]string{}
	for _, scope := range []string{ScopeAll, project} {
		vars, _ := ReadDotenv(e.secretsPath(scope))
		for key, value := range vars {
			result[key] = value
		}
	}
	return result
}

// safeRelative keeps an env file path inside its folder: relative, no
// parent references, no variables.
func safeRelative(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" || filepath.IsAbs(path) || strings.Contains(path, "$") {
		return "", fmt.Errorf("env file path %q must be relative to the folder", path)
	}
	clean := filepath.Clean(path)
	if clean == "." || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("env file path %q leaves the folder", path)
	}
	return clean, nil
}

func (e *Env) managedPath(project, rel string) string {
	return filepath.Join(e.Root, "files", project, rel)
}

// File returns a folder's env file as task-ui manages it.
func (e *Env) File(p *model.Project, path string) (model.EnvFile, error) {
	rel, err := safeRelative(path)
	if err != nil {
		return model.EnvFile{}, err
	}
	file := model.EnvFile{Project: p.ID, Path: rel}
	if data, err := os.ReadFile(e.managedPath(p.ID, rel)); err == nil {
		file.Managed = true
		file.Content = string(data)
	}
	if data, err := os.ReadFile(filepath.Join(p.Dir, rel)); err == nil {
		file.Exists = true
		if !file.Managed {
			file.Content = stripGenerated(string(data))
		}
	}
	vars, _ := parseDotenv(strings.NewReader(file.Content))
	file.Keys = sortedKeys(vars)
	return file, nil
}

// Files returns the env files a folder reads.
func (e *Env) Files(p *model.Project) []model.EnvFile {
	result := []model.EnvFile{}
	for _, path := range p.EnvFiles {
		file, err := e.File(p, path)
		if err != nil {
			continue
		}
		result = append(result, file)
	}
	return result
}

// Save stores the managed content of a folder's env file and generates
// the file in the folder.
func (e *Env) Save(p *model.Project, path, content string) error {
	rel, err := safeRelative(path)
	if err != nil {
		return err
	}
	content = strings.ReplaceAll(content, "\r\n", "\n")
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	managed := e.managedPath(p.ID, rel)
	if err := os.MkdirAll(filepath.Dir(managed), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(managed, []byte(content), 0o600); err != nil {
		return err
	}
	return e.generate(p, rel)
}

// generatedMarker separates the managed content from the secrets task-ui
// appends. Reading a generated file back drops everything after it.
const generatedMarker = "# --- task-ui secrets, generated before each run ---"

func stripGenerated(content string) string {
	if i := strings.Index(content, generatedMarker); i >= 0 {
		return strings.TrimRight(content[:i], "\n") + "\n"
	}
	return content
}

// Generate writes every env file a folder reads into the folder, with the
// secrets appended, ready for a run. A file the UI never managed and that
// is not in the folder is created from the secrets alone.
func (e *Env) Generate(p *model.Project) error {
	var failures []error
	for _, path := range p.EnvFiles {
		rel, err := safeRelative(path)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		if err := e.generate(p, rel); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func (e *Env) generate(p *model.Project, rel string) error {
	file, err := e.File(p, rel)
	if err != nil {
		return err
	}
	secrets := e.SecretsFor(p.ID)
	content := strings.TrimRight(file.Content, "\n")
	if len(secrets) > 0 {
		if content != "" {
			content += "\n\n"
		}
		content += generatedMarker + "\n"
		for _, key := range sortedKeys(secrets) {
			content += key + "=" + quote(secrets[key]) + "\n"
		}
	} else if content != "" {
		content += "\n"
	}
	target := filepath.Join(p.Dir, rel)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, []byte(content), 0o600)
}

// ReadDotenv reads a KEY=VALUE file the way atkins does: comments and
// blank lines skipped, quotes around a value removed.
func ReadDotenv(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return parseDotenv(file)
}

func parseDotenv(r interface{ Read([]byte) (int, error) }) (map[string]string, error) {
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
		if !ValidKey(key) {
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

func writeDotenv(path string, vars map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	var b strings.Builder
	for _, key := range sortedKeys(vars) {
		b.WriteString(key + "=" + quote(vars[key]) + "\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o600)
}

// quote wraps a value in double quotes when a bare one would be read
// differently: spaces, a leading quote, or a comment character.
func quote(value string) string {
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

func sortedKeys(vars map[string]string) []string {
	keys := make([]string, 0, len(vars))
	for key := range vars {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
