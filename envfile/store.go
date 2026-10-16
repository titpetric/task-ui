package envfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/dotenv"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/secret"
)

// Store keeps the managed content under <root>/<folder>/<path> and
// generates the file in the folder: the managed content (or the file
// already there), then the secrets. A source file is never written.
type Store struct {
	Root    string
	Secrets *secret.Store
}

// marker separates the managed content from the secrets task-ui appends.
// Reading a generated file back drops everything after it.
const marker = "# --- task-ui secrets, generated before each run ---"

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

func (s *Store) managedPath(folderID, rel string) string {
	return filepath.Join(s.Root, folderID, rel)
}

// File returns a folder's env file as task-ui manages it.
func (s *Store) File(f *folder.Folder, path string) (File, error) {
	rel, err := safeRelative(path)
	if err != nil {
		return File{}, err
	}
	file := File{Folder: f.ID, Path: rel}
	if data, err := os.ReadFile(s.managedPath(f.ID, rel)); err == nil {
		file.Managed = true
		file.Content = string(data)
	}
	if data, err := os.ReadFile(filepath.Join(f.Dir, rel)); err == nil {
		file.Exists = true
		if !file.Managed {
			file.Content = stripGenerated(string(data))
		}
	}
	vars, _ := dotenv.Parse(strings.NewReader(file.Content))
	file.Keys = dotenv.Keys(vars)
	return file, nil
}

// Files returns the env files a folder reads.
func (s *Store) Files(f *folder.Folder) []File {
	result := []File{}
	for _, path := range f.EnvFiles {
		file, err := s.File(f, path)
		if err != nil {
			continue
		}
		result = append(result, file)
	}
	return result
}

// Save stores the managed content of a folder's env file and generates
// the file in the folder.
func (s *Store) Save(f *folder.Folder, file File) error {
	rel, err := safeRelative(file.Path)
	if err != nil {
		return err
	}
	content := strings.ReplaceAll(file.Content, "\r\n", "\n")
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	managed := s.managedPath(f.ID, rel)
	if err := os.MkdirAll(filepath.Dir(managed), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(managed, []byte(content), 0o600); err != nil {
		return err
	}
	return s.generate(f, rel)
}

func stripGenerated(content string) string {
	if i := strings.Index(content, marker); i >= 0 {
		return strings.TrimRight(content[:i], "\n") + "\n"
	}
	return content
}

// Generate writes every env file a folder reads into the folder, with the
// secrets appended, ready for a run. A file the UI never managed and that
// is not in the folder is created from the secrets alone.
func (s *Store) Generate(f *folder.Folder) error {
	var failures []error
	for _, path := range f.EnvFiles {
		rel, err := safeRelative(path)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		if err := s.generate(f, rel); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func (s *Store) generate(f *folder.Folder, rel string) error {
	file, err := s.File(f, rel)
	if err != nil {
		return err
	}
	secrets := s.Secrets.For(f.ID)
	content := strings.TrimRight(file.Content, "\n")
	if len(secrets) > 0 {
		if content != "" {
			content += "\n\n"
		}
		content += marker + "\n" + dotenv.Format(secrets)
	} else if content != "" {
		content += "\n"
	}
	target := filepath.Join(f.Dir, rel)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, []byte(content), 0o600)
}
