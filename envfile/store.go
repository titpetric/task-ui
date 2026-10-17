// Package envfile manages the dotenv files a folder's sources read: the
// content an admin edits, kept in storage, and the file generated into
// the folder with the secrets appended before each run. A source file is
// never written.
package envfile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/dotenv"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// Marker separates the managed content from the secrets task-ui appends.
// Reading a generated file back drops everything after it.
const Marker = "# --- task-ui secrets, generated before each run ---"

// Store reads and writes the env files of folders.
type Store struct {
	// Managed holds the content saved from the UI.
	Managed *storage.EnvFileStorage
	// Secrets holds what is appended at generation.
	Secrets *storage.SecretStorage
}

// New returns the store over the two storages it reads.
func New(managed *storage.EnvFileStorage, secrets *storage.SecretStorage) *Store {
	return &Store{Managed: managed, Secrets: secrets}
}

// SafeRelative keeps an env file path inside its folder: relative, no
// parent references, no variables.
func SafeRelative(path string) (string, error) {
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

// File returns a folder's env file as task-ui sees it: the managed
// content when there is one, the folder's own file otherwise.
func (s *Store) File(ctx context.Context, f *folder.Folder, path string) (File, error) {
	rel, err := SafeRelative(path)
	if err != nil {
		return File{}, err
	}
	file := File{Folder: f.ID, Path: rel}
	managed, err := s.Managed.Get(ctx, model.EnvFile{Folder: f.ID, Path: rel})
	if err == nil {
		file.Managed = true
		file.Content = managed.Content
	} else if !errors.Is(err, storage.ErrNotFound) {
		return File{}, err
	}
	if data, err := os.ReadFile(filepath.Join(f.Dir, rel)); err == nil {
		file.Exists = true
		if !file.Managed {
			file.Content = StripGenerated(string(data))
		}
	}
	vars, _ := dotenv.Parse(strings.NewReader(file.Content))
	file.Keys = dotenv.Keys(vars)
	return file, nil
}

// Files returns the env files a folder reads.
func (s *Store) Files(ctx context.Context, f *folder.Folder) []File {
	result := []File{}
	for _, path := range f.EnvFiles {
		file, err := s.File(ctx, f, path)
		if err != nil {
			continue
		}
		result = append(result, file)
	}
	return result
}

// Save stores the managed content of a folder's env file and generates
// the file in the folder.
func (s *Store) Save(ctx context.Context, f *folder.Folder, file File) error {
	rel, err := SafeRelative(file.Path)
	if err != nil {
		return err
	}
	content := strings.ReplaceAll(file.Content, "\r\n", "\n")
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if err := s.Managed.Set(ctx, &model.EnvFile{Folder: f.ID, Path: rel, Content: content}); err != nil {
		return err
	}
	return s.generate(ctx, f, rel)
}

// StripGenerated returns the content before the secrets marker.
func StripGenerated(content string) string {
	if i := strings.Index(content, Marker); i >= 0 {
		return strings.TrimRight(content[:i], "\n") + "\n"
	}
	return content
}

// Generate writes every env file a folder reads into the folder, with the
// secrets appended, ready for a run. A file never managed and not in the
// folder is created from the secrets alone.
func (s *Store) Generate(ctx context.Context, f *folder.Folder) error {
	var failures []error
	for _, path := range f.EnvFiles {
		rel, err := SafeRelative(path)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		if err := s.generate(ctx, f, rel); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func (s *Store) generate(ctx context.Context, f *folder.Folder, rel string) error {
	file, err := s.File(ctx, f, rel)
	if err != nil {
		return err
	}
	secrets, err := s.Secrets.For(ctx, f.ID)
	if err != nil {
		return err
	}
	target := filepath.Join(f.Dir, rel)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, []byte(Render(file.Content, secrets)), 0o600)
}

// Render joins the managed content and the secrets into what the folder
// file holds.
func Render(content string, secrets map[string]string) string {
	content = strings.TrimRight(content, "\n")
	if len(secrets) > 0 {
		if content != "" {
			content += "\n\n"
		}
		return content + Marker + "\n" + dotenv.Format(secrets)
	}
	if content != "" {
		content += "\n"
	}
	return content
}
