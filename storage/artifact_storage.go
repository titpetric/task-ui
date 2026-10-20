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
	// Path is a file of the artifact tree: its versions.
	Path  string
	Limit int
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
	add("path", filter.Path)
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

// Latest returns the newest version of every path in the artifact tree,
// the artifacts with a commit, by path.
func (s *ArtifactStorage) Latest(ctx context.Context) ([]model.Artifact, error) {
	return client(s.db).Select[model.Artifact](ctx, `SELECT * FROM artifact WHERE id IN (
		SELECT id FROM (
			SELECT id, ROW_NUMBER() OVER (PARTITION BY path ORDER BY rowid DESC) AS version
			FROM artifact WHERE commit_id != ''
		) WHERE version = 1
	) ORDER BY path`)
}

// Squash follows the artifact tree flattening its history to one commit:
// the newest version of each path now lives in head, and the artifacts of
// the older versions go, as their files did.
func (s *ArtifactStorage) Squash(ctx context.Context, head string) error {
	db := client(s.db)
	if err := db.Exec(ctx, `DELETE FROM artifact WHERE commit_id != '' AND id NOT IN (
		SELECT id FROM (
			SELECT id, ROW_NUMBER() OVER (PARTITION BY path ORDER BY rowid DESC) AS version
			FROM artifact WHERE commit_id != ''
		) WHERE version = 1
	)`); err != nil {
		return err
	}
	return db.Exec(ctx, "UPDATE artifact SET commit_id = ? WHERE commit_id != ''", head)
}

// DeleteAll removes every artifact.
func (s *ArtifactStorage) DeleteAll(ctx context.Context) error {
	return client(s.db).Exec(ctx, "DELETE FROM artifact")
}
