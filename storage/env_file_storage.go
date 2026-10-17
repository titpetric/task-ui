package storage

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/titpetric/task-ui/model"
)

// EnvFileStorage keeps the managed content of the dotenv files folders
// read.
type EnvFileStorage struct {
	db *sqlx.DB
}

// NewEnvFileStorage returns the storage over a pool.
func NewEnvFileStorage(db *sqlx.DB) *EnvFileStorage {
	return &EnvFileStorage{db: db}
}

// Create inserts an env file.
func (s *EnvFileStorage) Create(ctx context.Context, file *model.EnvFile) error {
	if file.Folder == "" || file.Path == "" {
		return errors.New("an env file needs a folder and a path")
	}
	if file.ID == "" {
		file.ID = model.NewID()
	}
	file.CreatedAt = stamp()
	file.UpdatedAt = file.CreatedAt
	return client(s.db).Insert(ctx, model.EnvFileTable, file)
}

// Get returns the env file of a folder at a path, as the ref names them.
func (s *EnvFileStorage) Get(ctx context.Context, ref model.EnvFile) (*model.EnvFile, error) {
	file, err := client(s.db).Get[model.EnvFile](ctx, "SELECT * FROM env_file WHERE folder = ? AND path = ?", ref.Folder, ref.Path)
	return file, notFound(err)
}

// List returns the env files of a folder, by path.
func (s *EnvFileStorage) List(ctx context.Context, folder string) ([]model.EnvFile, error) {
	return client(s.db).Select[model.EnvFile](ctx, "SELECT * FROM env_file WHERE folder = ? ORDER BY path", folder)
}

// Set creates an env file, or replaces the content of the one at the
// same folder and path.
func (s *EnvFileStorage) Set(ctx context.Context, file *model.EnvFile) error {
	existing, err := s.Get(ctx, *file)
	if errors.Is(err, ErrNotFound) {
		return s.Create(ctx, file)
	}
	if err != nil {
		return err
	}
	file.ID = existing.ID
	file.CreatedAt = existing.CreatedAt
	return s.Update(ctx, file)
}

// Update writes an env file back, by ID.
func (s *EnvFileStorage) Update(ctx context.Context, file *model.EnvFile) error {
	file.UpdatedAt = stamp()
	return client(s.db).Update(ctx, model.EnvFileTable, file, "id")
}

// Delete removes an env file by ID.
func (s *EnvFileStorage) Delete(ctx context.Context, id string) error {
	return client(s.db).Exec(ctx, "DELETE FROM env_file WHERE id = ?", id)
}
