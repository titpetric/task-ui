package storage

import (
	"context"

	"github.com/titpetric/task-ui/model"
)

// FolderRefStorage finds the rows that name a folder by its ID: the
// schedules, the secrets, the env files and the workflow steps.
type FolderRefStorage struct {
	db DB
}

// NewFolderRefStorage returns the storage over a pool.
func NewFolderRefStorage(db DB) *FolderRefStorage {
	return &FolderRefStorage{db: db}
}

// folderRefQuery selects the rows of a folder and of the folders inside
// it. Each table matches on the ID or on the ID and a slash as a
// prefix with substr, as an ID may hold the _ that LIKE matches any
// character with.
const folderRefQuery = `
SELECT 'schedule' AS kind, id, folder, job AS name FROM schedule
	WHERE folder = ?1 OR substr(folder, 1, ?2) = ?3
UNION ALL
SELECT 'secret', id, scope, name FROM secret
	WHERE scope = ?1 OR substr(scope, 1, ?2) = ?3
UNION ALL
SELECT 'env_file', id, folder, path FROM env_file
	WHERE folder = ?1 OR substr(folder, 1, ?2) = ?3
UNION ALL
SELECT 'workflow_step', w.id, s.folder, w.name || ': ' || s.job FROM workflow_step s JOIN workflow w ON w.id = s.workflow_id
	WHERE s.folder = ?1 OR substr(s.folder, 1, ?2) = ?3
ORDER BY kind, folder, name`

// Under returns the rows that name the folder or one inside it, by kind,
// folder and name.
func (s *FolderRefStorage) Under(ctx context.Context, id string) ([]model.FolderRef, error) {
	prefix := id + "/"
	refs, err := client(s.db).Select[model.FolderRef](ctx, folderRefQuery, id, len(prefix), prefix)
	if err != nil {
		return nil, err
	}
	if refs == nil {
		refs = []model.FolderRef{}
	}
	return refs, nil
}
