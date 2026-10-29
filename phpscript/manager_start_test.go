package phpscript_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

func TestStart(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	module(t, root, "titpetric/refinement", "<?php echo 'one';")
	registry := &fakeRegistry{}

	opts := phpscript.NewOptions(root)
	opts.Registry = registry
	m := phpscript.New(opts)
	assert.NoError(t, m.Start(context.Background()))

	assert.Equal(t, []string{"titpetric/refinement"}, registry.names)
	assert.Equal(t, []string{"titpetric/refinement"}, registry.forgot, "Forget takes every name a scan found")
}

func TestStartReports(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// No annotation, so the file is never run.
	write(t, filepath.Join(root, "a", "quiet", phpscript.ModInfo), "<?php\n$m = new Task\\Module;\n")
	// Annotated, and it registers nothing.
	write(t, filepath.Join(root, "a", "silent", phpscript.ModInfo), "<?php\n// @startup\n$m = new Task\\Module;\n")
	// Annotated, and it throws.
	write(t, filepath.Join(root, "a", "broken", phpscript.ModInfo), "<?php\n// @startup\nthrow new Exception(\"no\");\n")
	// Annotated, names itself wrongly.
	write(t, filepath.Join(root, "a", "unnamed", phpscript.ModInfo), "<?php\n// @startup\n$m = new Task\\Module;\n$m->setName(\"nope\");\nTask::register($m);\n")
	module(t, root, "titpetric/good", "<?php echo 'fine';")

	var reported []string
	opts := phpscript.NewOptions(root)
	opts.OnError = func(err error) { reported = append(reported, err.Error()) }
	m := phpscript.New(opts)
	assert.NoError(t, m.Start(context.Background()))

	assert.Len(t, m.Modules(), 1, "the module that registered serves, the four that did not are reported")
	assert.Len(t, reported, 4)
	for _, want := range []string{"no @startup annotation", "called no Task::register", "quiet", "silent", "broken", "unnamed"} {
		assert.True(t, said(reported, want), want)
	}
}

// said reports whether any line holds want.
func said(lines []string, want string) bool {
	for _, line := range lines {
		if strings.Contains(line, want) {
			return true
		}
	}
	return false
}

func TestStartRegistryFails(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	module(t, root, "titpetric/refinement", "<?php echo 'one';")

	var reported []error
	opts := phpscript.NewOptions(root)
	opts.Registry = &fakeRegistry{err: errors.New("the database is locked")}
	opts.OnError = func(err error) { reported = append(reported, err) }
	m := phpscript.New(opts)

	assert.NoError(t, m.Start(context.Background()), "a registry that refuses does not stop the instance")
	assert.Len(t, reported, 2, "the register and the forget both report")
	assert.Len(t, m.Modules(), 1, "the module serves from memory either way")
}

func TestStartUnreadable(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dir := filepath.Join(root, "a", "dangling")
	assert.NoError(t, os.MkdirAll(dir, 0o755))
	assert.NoError(t, os.Symlink(filepath.Join(root, "gone"), filepath.Join(dir, phpscript.ModInfo)))

	m := phpscript.New(phpscript.NewOptions(root))
	assert.Error(t, m.Start(context.Background()), "a modinfo.php the server cannot read is a misconfigured instance")

	file := filepath.Join(t.TempDir(), "notadir")
	assert.NoError(t, os.WriteFile(file, []byte("x"), 0o644))
	plain := phpscript.New(phpscript.NewOptions(file))
	assert.NoError(t, plain.Start(context.Background()))
	assert.Len(t, plain.Modules(), 0, "a root that is a plain file holds no module")
}

func TestReload(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	module(t, root, "titpetric/one", "<?php echo 'one';")

	registry := &fakeRegistry{}
	opts := phpscript.NewOptions(root)
	opts.Registry = registry
	m := phpscript.New(opts)
	assert.NoError(t, m.Start(context.Background()))
	assert.Len(t, m.Modules(), 1)

	module(t, root, "titpetric/two", "<?php echo 'two';")
	assert.NoError(t, m.Reload(context.Background()))
	assert.Len(t, m.Modules(), 2, "a module added to a root appears on a reload")

	assert.NoError(t, os.RemoveAll(filepath.Join(root, "titpetric", "one")))
	assert.NoError(t, m.Reload(context.Background()))
	assert.Len(t, m.Modules(), 1, "a module whose directory is gone goes with it")
	assert.Equal(t, []string{"titpetric/two"}, registry.forgot)
}

// A module's templates compile into its own templates_c, which it ships and
// MiniTPL's compiler creates when it has to. Start touches neither, so a
// shipped templates_c is exactly what a request reads.
func TestStartLeavesTemplates(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dir := module(t, root, "titpetric/refinement", "<?php echo 'one';")
	compiled := filepath.Join(dir, "templates_c", "templates", "thread.tpl")
	write(t, compiled, "compiled")

	m := phpscript.New(phpscript.NewOptions(root))
	assert.NoError(t, m.Start(context.Background()))

	shipped, err := os.ReadFile(compiled)
	assert.NoError(t, err)
	assert.Equal(t, "compiled", string(shipped))

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	for _, entry := range entries {
		assert.NotEqual(t, "tmp", entry.Name(), "Start writes no directory into a module")
	}
}
