package phpscript_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

// fakeRegistry records what Start registered and what it forgot.
type fakeRegistry struct {
	names  []string
	forgot []string
	err    error
}

func (f *fakeRegistry) Register(_ context.Context, m *phpscript.Module) error {
	f.names = append(f.names, m.Name())
	return f.err
}

func (f *fakeRegistry) Forget(_ context.Context, found []string) (int, error) {
	f.forgot = found
	return 0, f.err
}

// module writes a module under root and returns its directory. body is the
// whole of its index.php.
func module(t *testing.T, root, name, body string) string {
	t.Helper()
	dir := filepath.Join(root, filepath.FromSlash(name))
	write(t, filepath.Join(dir, phpscript.ModInfo),
		"<?php\n// @startup\n$m = new Task\\Module;\n$m->setName(\""+name+"\");\n$m->setTitle(\"The "+name+"\");\nTask::register($m);\n")
	if body != "" {
		write(t, filepath.Join(dir, phpscript.Index), body)
	}
	return dir
}

func TestNew(t *testing.T) {
	t.Parallel()
	m := phpscript.New(phpscript.Options{})
	assert.Equal(t, phpscript.DefaultTimeout, m.Timeout(), "an unset timeout takes the default")
	assert.Len(t, m.Modules(), 0)

	_, ok := m.Module("titpetric/refinement")
	assert.False(t, ok)

	m = phpscript.New(phpscript.Options{Timeout: time.Second, PoolSize: 1})
	assert.Equal(t, time.Second, m.Timeout())
}

func TestNewOptions(t *testing.T) {
	t.Parallel()
	opts := phpscript.NewOptions("/data/modules")
	assert.Equal(t, []string{"/data/modules"}, opts.Roots)
	assert.Equal(t, phpscript.DefaultTimeout, opts.Timeout)
	assert.GreaterOrEqual(t, opts.PoolSize, 2)
}

func TestManagerNil(t *testing.T) {
	t.Parallel()
	var m *phpscript.Manager
	assert.Len(t, m.Modules(), 0, "a nil Manager is an instance with no modules")
	assert.Equal(t, phpscript.DefaultTimeout, m.Timeout())
	assert.NoError(t, m.Start(context.Background()))
	assert.NoError(t, m.Reload(context.Background()))

	_, ok := m.Module("titpetric/refinement")
	assert.False(t, ok)

	_, err := m.Render(context.Background(), nil, nil)
	assert.ErrorIs(t, err, phpscript.ErrNotFound)
}

func TestManagerModules(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	module(t, root, "titpetric/refinement", "<?php echo 'one';")
	module(t, root, "acme/notes", "<?php echo 'two';")

	m := phpscript.New(phpscript.NewOptions(root))
	assert.NoError(t, m.Start(context.Background()))

	modules := m.Modules()
	assert.Len(t, modules, 2)
	assert.Equal(t, "acme/notes", modules[0].Name(), "the modules come back by name")
	assert.Equal(t, "titpetric/refinement", modules[1].Name())
	assert.Equal(t, "The acme/notes", modules[0].Title)

	found, ok := m.Module("titpetric/refinement")
	assert.True(t, ok)
	assert.Equal(t, filepath.Join(root, "titpetric", "refinement"), found.Root)
}
