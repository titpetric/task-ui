package recording_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/recording"
)

func TestStore_Path(t *testing.T) {
	t.Parallel()
	s := &recording.Store{Root: "/data"}
	id := model.NewID()
	path, err := s.Path(&model.Run{ID: id, Folder: "infra", Job: "make:build"})
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join("/data", "infra", "make:build", id+".cast"), path)
	assert.Equal(t, filepath.Join("infra", "make:build", id+".cast"), s.Relative(path))

	_, err = s.Path(&model.Run{ID: "../../etc/passwd", Folder: "infra", Job: "x"})
	assert.Error(t, err)
}

func TestStore_Open(t *testing.T) {
	t.Parallel()
	s := &recording.Store{Root: t.TempDir()}
	run := &model.Run{ID: model.NewID(), Folder: "infra", Job: "deploy", TermCols: 80, TermRows: 24}
	rec, err := s.Open(run)
	assert.NoError(t, err)
	assert.FileExists(t, rec.Path())
	assert.NoError(t, rec.Close())

	data, err := os.ReadFile(rec.Path())
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"version":2`)
	assert.Contains(t, string(data), `"width":80`)

	_, err = s.Open(&model.Run{ID: "bad"})
	assert.Error(t, err)
}

func TestStore_Relative(t *testing.T) {
	t.Parallel()
	s := &recording.Store{Root: "/data"}
	assert.Equal(t, "a/b.cast", s.Relative("/data/a/b.cast"))
	assert.Equal(t, "../x", s.Relative("/x"), "a path outside the root is relative to it")
}

// Read opens the recording a run row names; a row without one, or whose
// file is gone, has none.
func TestStore_Read(t *testing.T) {
	t.Parallel()
	s := &recording.Store{Root: t.TempDir()}
	run := &model.Run{ID: model.NewID(), Folder: "infra", Job: "plan"}
	path, err := s.Path(run)
	assert.NoError(t, err)
	assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	assert.NoError(t, os.WriteFile(path, []byte("{\"version\":2}\n"), 0o644))
	run.CastPath = s.Relative(path)

	file, err := s.Read(run)
	assert.NoError(t, err)
	data, _ := os.ReadFile(file.Name())
	file.Close()
	assert.Equal(t, "{\"version\":2}\n", string(data))

	_, err = s.Read(&model.Run{})
	assert.ErrorIs(t, err, os.ErrNotExist)
	_, err = s.Read(&model.Run{CastPath: "infra/plan/gone.cast"})
	assert.ErrorIs(t, err, os.ErrNotExist)
}
