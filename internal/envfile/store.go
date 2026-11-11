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

// Marker separates the filled content from the secrets task-ui appends.
const Marker = "# --- task-ui secrets, generated before each run ---"

// Store writes the env files users fill into their folders. It never
// reads a folder's sources or files: a folder has the env files a user
// filled for it and nothing else.
type Store struct {
	// Managed holds the content users fill.
	Managed Rows
	// Secrets returns what a written file carries after the content.
	Secrets Secrets
}

// New returns the store over the rows and the secrets it reads.
func New(managed Rows, secrets Secrets) *Store {
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

// File returns a folder's env file with the content a user filled, empty
// for a file nobody filled; an unsafe path or a storage failure errs.
func (s *Store) File(ctx context.Context, f *folder.Folder, path string) (File, error) {
	rel, err := SafeRelative(path)
	if err != nil {
		return File{}, err
	}
	row, err := s.Managed.Get(ctx, model.EnvFile{Folder: f.ID, Path: rel})
	switch {
	case errors.Is(err, storage.ErrNotFound):
		return File{Folder: f.ID, Path: rel}, nil
	case err != nil:
		return File{}, err
	}
	return File{Folder: f.ID, Path: rel, Content: row.Content}, nil
}

// Files returns the env files users filled for a folder, none when
// storage fails.
func (s *Store) Files(ctx context.Context, f *folder.Folder) []File {
	rows, err := s.Managed.List(ctx, f.ID)
	if err != nil {
		return []File{}
	}
	result := make([]File, 0, len(rows))
	for _, row := range rows {
		result = append(result, File{Folder: f.ID, Path: row.Path, Content: row.Content})
	}
	return result
}

// Save keeps the content of a folder's env file and writes the file in
// the folder; no content deletes the row and the written file.
func (s *Store) Save(ctx context.Context, f *folder.Folder, file File) error {
	rel, err := SafeRelative(file.Path)
	if err != nil {
		return err
	}
	content := strings.ReplaceAll(file.Content, "\r\n", "\n")
	if strings.TrimSpace(content) == "" {
		return s.clear(ctx, f, rel)
	}
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if err := s.Managed.Set(ctx, &model.EnvFile{Folder: f.ID, Path: rel, Content: content}); err != nil {
		return err
	}
	return s.write(ctx, f, rel, content)
}

// clear deletes the row of a file and the file task-ui wrote from it.
func (s *Store) clear(ctx context.Context, f *folder.Folder, rel string) error {
	row, err := s.Managed.Get(ctx, model.EnvFile{Folder: f.ID, Path: rel})
	switch {
	case errors.Is(err, storage.ErrNotFound):
		return nil
	case err != nil:
		return err
	}
	if err := s.Managed.Delete(ctx, row.ID); err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(f.Dir, rel)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// Generate writes each env file a user filled for the folder, with the
// secrets appended, ready for a run.
func (s *Store) Generate(ctx context.Context, f *folder.Folder) error {
	rows, err := s.Managed.List(ctx, f.ID)
	if err != nil {
		return err
	}
	var failures []error
	for _, row := range rows {
		rel, err := SafeRelative(row.Path)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		if err := s.write(ctx, f, rel, row.Content); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

func (s *Store) write(ctx context.Context, f *folder.Folder, rel, content string) error {
	secrets, err := s.Secrets.For(ctx, f.ID)
	if err != nil {
		return err
	}
	target := filepath.Join(f.Dir, rel)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, []byte(Render(content, secrets)), 0o600)
}

// Render joins the filled content and the secrets into what the folder
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
