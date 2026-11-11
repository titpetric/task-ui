package envfile_test

import (
	"context"
	"sort"
	"sync"
	"testing"

	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// rows keeps env file rows in memory, as envfile.Rows.
type rows struct {
	mu   sync.Mutex
	byID map[string]model.EnvFile
}

var _ envfile.Rows = (*rows)(nil)

func (r *rows) Get(_ context.Context, ref model.EnvFile) (*model.EnvFile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, row := range r.byID {
		if row.Folder == ref.Folder && row.Path == ref.Path {
			return &row, nil
		}
	}
	return nil, storage.ErrNotFound
}

func (r *rows) List(ctx context.Context, folder string) ([]model.EnvFile, error) {
	all, _ := r.All(ctx)
	var result []model.EnvFile
	for _, row := range all {
		if row.Folder == folder {
			result = append(result, row)
		}
	}
	return result, nil
}

func (r *rows) All(context.Context) ([]model.EnvFile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var result []model.EnvFile
	for _, row := range r.byID {
		result = append(result, row)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Folder+result[i].Path < result[j].Folder+result[j].Path })
	return result, nil
}

func (r *rows) Set(ctx context.Context, file *model.EnvFile) error {
	if old, err := r.Get(ctx, *file); err == nil {
		file.ID = old.ID
	} else {
		file.ID = model.NewID()
	}
	return r.Update(ctx, file)
}

func (r *rows) Update(_ context.Context, file *model.EnvFile) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.byID == nil {
		r.byID = map[string]model.EnvFile{}
	}
	r.byID[file.ID] = *file
	return nil
}

func (r *rows) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.byID, id)
	return nil
}

func TestRows(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var r envfile.Rows = &rows{}
	assert.NoError(t, r.Set(ctx, &model.EnvFile{Folder: "infra", Path: ".env", Content: "A=1\n"}))
	row, err := r.Get(ctx, model.EnvFile{Folder: "infra", Path: ".env"})
	assert.NoError(t, err)
	assert.Equal(t, "A=1\n", row.Content)
}
