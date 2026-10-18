package storage

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/titpetric/task-ui/model"
)

// ArtifactStorage keeps the files runs left behind.
type ArtifactStorage struct {
	db DB
}

// NewArtifactStorage returns the storage over a pool.
func NewArtifactStorage(db DB) *ArtifactStorage {
	return &ArtifactStorage{db: db}
}

// ArtifactFilter narrows a listing. Zero values match everything.
type ArtifactFilter struct {
	RunID  string
	Folder string
	Job    string
	Limit  int
}

// Create inserts an artifact.
func (s *ArtifactStorage) Create(ctx context.Context, artifact *model.Artifact) error {
	if artifact.RunID == "" || artifact.Name == "" {
		return errors.New("an artifact needs a run and a name")
	}
	if artifact.ID == "" {
		artifact.ID = model.NewID()
	}
	if artifact.CreatedAt == nil {
		artifact.CreatedAt = stamp()
	}
	return client(s.db).Insert(ctx, model.ArtifactTable, artifact)
}

// Get returns an artifact by ID.
func (s *ArtifactStorage) Get(ctx context.Context, id string) (*model.Artifact, error) {
	artifact, err := client(s.db).Get[model.Artifact](ctx, "SELECT * FROM artifact WHERE id = ?", id)
	return artifact, notFound(err)
}

// List returns artifacts matching the filter, newest first.
func (s *ArtifactStorage) List(ctx context.Context, filter ArtifactFilter) ([]model.Artifact, error) {
	var where []string
	var args []any
	add := func(column, value string) {
		if value != "" {
			where = append(where, column+" = ?")
			args = append(args, value)
		}
	}
	add("run_id", filter.RunID)
	add("folder", filter.Folder)
	add("job", filter.Job)
	query := "SELECT * FROM artifact"
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY created_at DESC, id DESC"
	if filter.Limit > 0 {
		query += " LIMIT " + strconv.Itoa(filter.Limit)
	}
	return client(s.db).Select[model.Artifact](ctx, query, args...)
}

// Delete removes an artifact by ID.
func (s *ArtifactStorage) Delete(ctx context.Context, id string) error {
	return client(s.db).Exec(ctx, "DELETE FROM artifact WHERE id = ?", id)
}

// DeleteByRun removes the artifacts of a run.
func (s *ArtifactStorage) DeleteByRun(ctx context.Context, runID string) error {
	return client(s.db).Exec(ctx, "DELETE FROM artifact WHERE run_id = ?", runID)
}
