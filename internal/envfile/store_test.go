package envfile_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func newStore() *envfile.Store {
	return envfile.New(&rows{}, secrets{"TOKEN": "s3cret"})
}

func TestNew(t *testing.T) {
	t.Parallel()
	s := newStore()
	assert.NotNil(t, s.Managed)
	assert.NotNil(t, s.Secrets)
}

func TestFile(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(envfile.File{Folder: "infra", Path: ".env", Content: "secret"})
	assert.NoError(t, err)
	assert.NotContains(t, string(data), "secret")
}

func TestSafeRelative(t *testing.T) {
	t.Parallel()
	rel, err := envfile.SafeRelative(" conf/./x.env ")
	assert.NoError(t, err)
	assert.Equal(t, "conf/x.env", rel)
	for _, bad := range []string{"", "/etc/env", "../x.env", "$HOME/.env", "."} {
		_, err := envfile.SafeRelative(bad)
		assert.Error(t, err, bad)
	}
}

func TestRender(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", envfile.Render("", nil))
	assert.Equal(t, "A=1\n", envfile.Render("A=1", nil))
	assert.Equal(t, envfile.Marker+"\nX=y\n", envfile.Render("", map[string]string{"X": "y"}))
	assert.Equal(t, "A=1\n\n"+envfile.Marker+"\nX=y\n", envfile.Render("A=1\n", map[string]string{"X": "y"}))
}

// A folder whose sources read dotenv files, with one in it, lists none
// and gets none written: the store reads no source and no file.
func TestStore_NoDiscovery(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newStore()
	dir := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(dir, "Taskfile.yml"), []byte("version: '3'\ndotenv: ['.env', 'deploy.env']\n"), 0o644))
	assert.NoError(t, os.WriteFile(filepath.Join(dir, ".env"), []byte("PORT=8080\n"), 0o644))
	f := &folder.Folder{ID: "infra", Dir: dir}

	assert.Len(t, s.Files(ctx, f), 0)
	assert.NoError(t, s.Generate(ctx, f))
	data, err := os.ReadFile(filepath.Join(dir, ".env"))
	assert.NoError(t, err)
	assert.Equal(t, "PORT=8080\n", string(data))
	assert.NoFileExists(t, filepath.Join(dir, "deploy.env"))

	// The folder's own file is not read: unfilled reads empty.
	file, err := s.File(ctx, f, ".env")
	assert.NoError(t, err)
	assert.Equal(t, "", file.Content)
	_, err = s.File(ctx, f, "../outside.env")
	assert.Error(t, err)
}

func TestStore_Save(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newStore()
	dir := t.TempDir()
	f := &folder.Folder{ID: "infra", Dir: dir}

	assert.NoError(t, s.Save(ctx, f, envfile.File{Path: "conf/.env", Content: "PORT=8080\r\nNAME=app"}))
	data, err := os.ReadFile(filepath.Join(dir, "conf", ".env"))
	assert.NoError(t, err)
	assert.Equal(t, "PORT=8080\nNAME=app\n\n"+envfile.Marker+"\nTOKEN=s3cret\n", string(data))

	files := s.Files(ctx, f)
	assert.Len(t, files, 1)
	assert.Equal(t, "conf/.env", files[0].Path)
	file, err := s.File(ctx, f, "conf/.env")
	assert.NoError(t, err)
	assert.Equal(t, "PORT=8080\nNAME=app\n", file.Content)

	assert.Error(t, s.Save(ctx, f, envfile.File{Path: "../x.env", Content: "A=1"}))

	// Saving no content deletes the row and the written file.
	assert.NoError(t, s.Save(ctx, f, envfile.File{Path: "conf/.env", Content: " \n"}))
	assert.Len(t, s.Files(ctx, f), 0)
	assert.NoFileExists(t, filepath.Join(dir, "conf", ".env"))
	assert.NoError(t, s.Save(ctx, f, envfile.File{Path: "never.env"}))
}

func TestStore_Generate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newStore()
	dir := t.TempDir()
	f := &folder.Folder{ID: "infra", Dir: dir}
	assert.NoError(t, s.Managed.Set(ctx, &model.EnvFile{Folder: "infra", Path: "deploy.env", Content: "A=1\n"}))
	assert.NoError(t, s.Managed.Set(ctx, &model.EnvFile{Folder: "infra", Path: "../outside.env", Content: "B=2\n"}))
	assert.NoError(t, s.Managed.Set(ctx, &model.EnvFile{Folder: "web", Path: "web.env", Content: "C=3\n"}))

	assert.Error(t, s.Generate(ctx, f), "the outside path fails")
	data, err := os.ReadFile(filepath.Join(dir, "deploy.env"))
	assert.NoError(t, err)
	assert.Equal(t, "A=1\n\n"+envfile.Marker+"\nTOKEN=s3cret\n", string(data))
	assert.NoFileExists(t, filepath.Join(filepath.Dir(dir), "outside.env"))
	assert.NoFileExists(t, filepath.Join(dir, "web.env"))

	failing := envfile.New(s.Managed, secrets{"FAIL": "vault locked"})
	assert.Error(t, failing.Generate(ctx, f))
}
