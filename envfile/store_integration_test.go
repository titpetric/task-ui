//go:build integration

package envfile_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
	_ "modernc.org/sqlite"
)

func newStore(t *testing.T) *envfile.Store {
	t.Helper()
	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	keys := &sshkey.Store{Dir: t.TempDir()}
	_, err = keys.Generate("ed25519", "test")
	assert.NoError(t, err)
	return envfile.New(storage.NewEnvFileStorage(db), secret.New(storage.NewSecretStorage(db), secret.NewVault(secret.DirKeys{Store: keys})))
}

func TestNew(t *testing.T) {
	t.Parallel()
	s := newStore(t)
	assert.NotNil(t, s.Managed)
	assert.NotNil(t, s.Secrets)
}

func TestStore_Files(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newStore(t)
	dir := t.TempDir()
	f := &folder.Folder{ID: "infra", Dir: dir, EnvFiles: []string{".env", "conf/deploy.env", "../outside.env"}}
	assert.NoError(t, os.WriteFile(filepath.Join(dir, ".env"), []byte("PORT=8080\n# note\nNAME=\"my app\"\n"), 0o644))

	files := s.Files(ctx, f)
	assert.Len(t, files, 2)
	assert.True(t, files[0].Exists)
	assert.False(t, files[0].Managed)
	assert.Equal(t, []string{"NAME", "PORT"}, files[0].Keys)
	assert.False(t, files[1].Exists)
	_, err := s.File(ctx, f, "../outside.env")
	assert.Error(t, err)
}

func TestStore_Save(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newStore(t)
	dir := t.TempDir()
	f := &folder.Folder{ID: "infra", Dir: dir, EnvFiles: []string{".env"}}
	assert.NoError(t, s.Secrets.Set(ctx, &model.Secret{Scope: storage.ScopeAll, Name: "TOKEN", Content: "s3cret"}))
	assert.NoError(t, s.Secrets.Set(ctx, &model.Secret{Scope: "infra", Name: "PORT", Content: "9090"}))

	assert.NoError(t, s.Save(ctx, f, envfile.File{Path: ".env", Content: "PORT=8080\r\nNAME=app"}))
	data, err := os.ReadFile(filepath.Join(dir, ".env"))
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(string(data), "PORT=8080\nNAME=app\n\n"+envfile.Marker+"\nPORT=9090\nTOKEN=s3cret\n"), string(data))

	// Reading it back shows the managed content alone.
	file, err := s.File(ctx, f, ".env")
	assert.NoError(t, err)
	assert.True(t, file.Managed)
	assert.Equal(t, "PORT=8080\nNAME=app\n", file.Content)

	assert.Error(t, s.Save(ctx, f, envfile.File{Path: "../x.env", Content: "A=1"}))
}

func TestStore_Generate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	s := newStore(t)
	dir := t.TempDir()
	f := &folder.Folder{ID: "infra", Dir: dir, EnvFiles: []string{"conf/deploy.env", "../outside.env"}}
	assert.NoError(t, s.Secrets.Set(ctx, &model.Secret{Scope: storage.ScopeAll, Name: "TOKEN", Content: "s3cret"}))

	assert.Error(t, s.Generate(ctx, f), "the outside path fails")
	data, err := os.ReadFile(filepath.Join(dir, "conf", "deploy.env"))
	assert.NoError(t, err)
	assert.Contains(t, string(data), "TOKEN=s3cret")
	assert.NoFileExists(t, filepath.Join(filepath.Dir(dir), "outside.env"))

	// A folder file generated before reads back without the secrets.
	unmanaged := &folder.Folder{ID: "web", Dir: t.TempDir(), EnvFiles: []string{".env"}}
	assert.NoError(t, os.WriteFile(filepath.Join(unmanaged.Dir, ".env"), []byte("A=1\n\n"+envfile.Marker+"\nTOKEN=old\n"), 0o600))
	file, err := s.File(ctx, unmanaged, ".env")
	assert.NoError(t, err)
	assert.Equal(t, "A=1\n", file.Content)
	assert.Equal(t, []string{"A"}, file.Keys)
}
