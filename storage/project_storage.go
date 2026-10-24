package storage

import (
	"context"
	"errors"

	"github.com/titpetric/task-ui/model"
)

// ProjectStorage keeps what the configuration says about folders.
type ProjectStorage struct {
	db DB
}

// NewProjectStorage returns the storage over a pool.
func NewProjectStorage(db DB) *ProjectStorage {
	return &ProjectStorage{db: db}
}

// Create inserts a project. The folder is its identity for the catalogue.
func (s *ProjectStorage) Create(ctx context.Context, project *model.Project) error {
	if project.Folder == "" {
		return errors.New("a project needs a folder")
	}
	if project.ID == "" {
		project.ID = model.NewID()
	}
	fillProjectLists(project)
	project.CreatedAt = stamp()
	project.UpdatedAt = project.CreatedAt
	return client(s.db).Insert(ctx, model.ProjectTable, project)
}

// fillProjectLists sets the JSON columns to JSON, so a row made
// from a literal reads back the same as one made from a form.
func fillProjectLists(project *model.Project) {
	if project.Allow == "" {
		project.Allow = model.EncodeList(nil)
	}
	if project.Sources == "" {
		project.Sources = model.EncodeList(nil)
	}
	if project.Commands == "" {
		project.Commands = model.EncodeStrings(nil)
	}
	if project.JobAllow == "" {
		project.JobAllow = model.EncodeMap(nil)
	}
}

// Get returns a project by ID.
func (s *ProjectStorage) Get(ctx context.Context, id string) (*model.Project, error) {
	project, err := client(s.db).Get[model.Project](ctx, "SELECT * FROM project WHERE id = ?", id)
	return project, notFound(err)
}

// GetByFolder returns the project of a folder.
func (s *ProjectStorage) GetByFolder(ctx context.Context, folder string) (*model.Project, error) {
	project, err := client(s.db).Get[model.Project](ctx, "SELECT * FROM project WHERE folder = ?", folder)
	return project, notFound(err)
}

// List returns every project, by folder.
func (s *ProjectStorage) List(ctx context.Context) ([]model.Project, error) {
	return client(s.db).Select[model.Project](ctx, "SELECT * FROM project ORDER BY folder")
}

// Update writes a project back, by ID.
func (s *ProjectStorage) Update(ctx context.Context, project *model.Project) error {
	fillProjectLists(project)
	project.UpdatedAt = stamp()
	return client(s.db).Update(ctx, model.ProjectTable, project, "id")
}

// Upsert creates the project of a folder, or updates the one there is,
// keeping its ID.
func (s *ProjectStorage) Upsert(ctx context.Context, project *model.Project) error {
	existing, err := s.GetByFolder(ctx, project.Folder)
	if errors.Is(err, ErrNotFound) {
		return s.Create(ctx, project)
	}
	if err != nil {
		return err
	}
	project.ID = existing.ID
	project.CreatedAt = existing.CreatedAt
	return s.Update(ctx, project)
}

// Delete removes a project by ID, unsetting it as a workspace's home project.
func (s *ProjectStorage) Delete(ctx context.Context, id string) error {
	return inTx(ctx, s.db, func(c conn) error {
		if err := c.Exec(ctx, "UPDATE workspace SET home_project_id = '' WHERE home_project_id = ?", id); err != nil {
			return err
		}
		return c.Exec(ctx, "DELETE FROM project WHERE id = ?", id)
	})
}
