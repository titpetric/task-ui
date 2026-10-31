package phpscript_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/phpscript/runner"
	"github.com/titpetric/phpscript/stdlib"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

// fakeMigrator records the directory Task::migrate named.
type fakeMigrator struct {
	dirs []string
	err  error
}

func (f *fakeMigrator) Migrate(_ context.Context, m *phpscript.Module, dir string) error {
	f.dirs = append(f.dirs, m.Name()+":"+dir)
	return f.err
}

// fakeDispatcher records the comment, the action and the folder
// Task::dispatch named.
type fakeDispatcher struct {
	calls   []string
	folders []string
	id      string
	err     error
}

func (f *fakeDispatcher) Dispatch(_ context.Context, m *phpscript.Module, comment, action, folder string) (string, error) {
	f.calls = append(f.calls, m.Name()+":"+action+":"+comment)
	f.folders = append(f.folders, folder)
	return f.id, f.err
}

// run builds a runtime over dir, binds the host symbols and runs src,
// returning what it echoed. The module on the context is named, as a serving
// runtime's is, so a binding reports which module called it.
func run(t *testing.T, dir, src string, opts phpscript.Options, registered func(*phpscript.Module)) (string, error) {
	t.Helper()
	var out stringWriter
	rt := runner.New(&out, runner.Options{RootFS: os.DirFS(dir), WorkDir: "/", SAPI: "test"})
	stdlib.Register(rt)
	stdlib.RegisterFS(rt, dir)
	phpscript.Bind(rt, opts, registered)
	own := &phpscript.Module{Vendor: "titpetric", Plugin: "test", Root: dir}
	rt.SetContext(phpscript.WithModule(context.Background(), own))
	prog, err := rt.Load(src)
	if err != nil {
		return out.String(), err
	}
	err = rt.Run(prog)
	return out.String(), err
}

// stringWriter collects what a script echoed.
type stringWriter struct{ body []byte }

func (s *stringWriter) Write(p []byte) (int, error) {
	s.body = append(s.body, p...)
	return len(p), nil
}

func (s *stringWriter) String() string { return string(s.body) }

var _ io.Writer = (*stringWriter)(nil)

func TestBindRegister(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	var got *phpscript.Module
	_, err := run(t, dir, `<?php
$module = new Task\Module;
$module->setName("titpetric/refinement");
$module->setTitle("Refinement");
$module->setDescription("Refine a design through a thread of comments");
$module->setIcon("flow");
$module->set_version("2.0");
Task::register($module);
`, phpscript.Options{}, func(m *phpscript.Module) { got = m })
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "titpetric/refinement", got.Name())
	assert.Equal(t, "Refinement", got.Title)
	assert.Equal(t, "flow", got.Icon)
	assert.Equal(t, "2.0", got.Version, "a snake case method reaches the Go name")
	assert.Equal(t, dir, got.Root, "the root is where the file was found, not what the script says")
}

func TestBindRegisterRefuses(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	for _, src := range []string{
		`<?php $m = new Task\Module; $m->setName("nope"); Task::register($m);`,
		`<?php Task::register(new Task\Module);`,
	} {
		var got *phpscript.Module
		_, err := run(t, dir, src, phpscript.Options{}, func(m *phpscript.Module) { got = m })
		assert.Error(t, err, src)
		assert.Nil(t, got, src)
	}
}

func TestBindConstants(t *testing.T) {
	t.Parallel()
	body, err := run(t, t.TempDir(), `<?php echo Task::VERSION, " ", Task::PREFIX;`,
		phpscript.Options{Version: "1.4.0"}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "1.4.0 /m", body)
}

// fakeRuns answers the run of one note and nothing of any other.
type fakeRuns struct{ note, run string }

func (f *fakeRuns) RunOf(_ context.Context, note string) string {
	if note == f.note {
		return f.run
	}
	return ""
}

// Task::run is what a module draws a status label from: the run a
// dispatch turned into, by the note it answered with. A note with no run
// yet, an empty note and an instance with no lookup are all "", because
// a label with no run is a run that has not started and not an error.
func TestBindRun(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	opts := phpscript.Options{Runs: &fakeRuns{note: "01NOTE", run: "06grxxqjec8jjt19vd65wfdg8r"}}

	body, err := run(t, dir, `<?php echo "[", Task::run("01NOTE"), "]";`, opts, nil)
	assert.NoError(t, err)
	assert.Equal(t, "[06grxxqjec8jjt19vd65wfdg8r]", body)

	body, err = run(t, dir, `<?php echo "[", Task::run("01OTHER"), "]";`, opts, nil)
	assert.NoError(t, err)
	assert.Equal(t, "[]", body, "a note whose run has not started")

	body, err = run(t, dir, `<?php echo "[", Task::run("  "), "]";`, opts, nil)
	assert.NoError(t, err)
	assert.Equal(t, "[]", body, "no note is no run")

	body, err = run(t, dir, `<?php echo "[", Task::run("01NOTE"), "]";`, phpscript.Options{}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "[]", body, "an instance with no lookup answers rather than failing")
}

func TestBindMigrate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Join(dir, "schema"), 0o755))
	migrator := &fakeMigrator{}

	_, err := run(t, dir, `<?php Task::migrate("schema/");`,
		phpscript.Options{Migrator: migrator}, nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"titpetric/test:schema"}, migrator.dirs, "the module of the context is the one that migrates")

	_, err = run(t, dir, `<?php Task::migrate("../elsewhere");`,
		phpscript.Options{Migrator: migrator}, nil)
	assert.ErrorContains(t, err, "leaves the module root")

	_, err = run(t, dir, `<?php Task::migrate("schema/");`, phpscript.Options{}, nil)
	assert.ErrorContains(t, err, "applies no module schema")
}

func TestBindDispatch(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	write(t, filepath.Join(dir, "action", "refinement.md"), "Refine the artifacts.\n")
	dispatcher := &fakeDispatcher{id: "01JRUN"}

	body, err := run(t, dir, `<?php echo Task::dispatch("use ULID as primary keys", "action/refinement.md");`,
		phpscript.Options{Dispatcher: dispatcher}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "01JRUN", body)
	assert.Equal(t, []string{"titpetric/test:action/refinement.md:use ULID as primary keys"}, dispatcher.calls)
	assert.Equal(t, []string{""}, dispatcher.folders, "a call with no folder works at the top of the tree")

	// The folder is the third argument, and a module that keeps a scratch
	// space per thread passes the thread there.
	_, err = run(t, dir, `<?php Task::dispatch("again", "action/refinement.md", "01THREAD");`,
		phpscript.Options{Dispatcher: dispatcher}, nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"", "01THREAD"}, dispatcher.folders)

	_, err = run(t, dir, `<?php Task::dispatch("x", "action/gone.md");`,
		phpscript.Options{Dispatcher: dispatcher}, nil)
	assert.Error(t, err, "an action file that is not there is an error at the call")

	_, err = run(t, dir, `<?php Task::dispatch("  ", "action/refinement.md");`,
		phpscript.Options{Dispatcher: dispatcher}, nil)
	assert.ErrorContains(t, err, "takes the comment")

	_, err = run(t, dir, `<?php Task::dispatch("x", "action/refinement.md");`, phpscript.Options{}, nil)
	assert.ErrorContains(t, err, "starts no module runs")

	dispatcher.err = errors.New("the backlog is full")
	_, err = run(t, dir, `<?php Task::dispatch("x", "action/refinement.md");`,
		phpscript.Options{Dispatcher: dispatcher}, nil)
	assert.ErrorContains(t, err, "backlog")
}

func TestInside(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, filepath.Join(root, "schema", "one.up.sql"), "")

	for path, want := range map[string]string{
		"schema":            "schema",
		"schema/":           "schema",
		"/schema":           "schema",
		"schema/one.up.sql": filepath.Join("schema", "one.up.sql"),
		"./schema":          "schema",
		"":                  ".",
		"/":                 ".",
		"a/../schema":       "schema",
	} {
		got, err := phpscript.Inside(root, path)
		assert.NoError(t, err, path)
		assert.Equal(t, want, got, path)
	}

	for _, bad := range []string{"../escape", "schema/../../escape", "/../escape"} {
		_, err := phpscript.Inside(root, bad)
		assert.ErrorIs(t, err, phpscript.ErrEscapes, bad)
	}

	_, err := phpscript.Inside("", "schema")
	assert.ErrorIs(t, err, phpscript.ErrNoModule)
}

func TestInsideSymlink(t *testing.T) {
	t.Parallel()
	root, outside := t.TempDir(), t.TempDir()
	write(t, filepath.Join(outside, "secret.sql"), "")
	assert.NoError(t, os.Symlink(outside, filepath.Join(root, "link")))

	_, err := phpscript.Inside(root, "link")
	assert.ErrorIs(t, err, phpscript.ErrEscapes, "a symlink out of the root is out of the root")
}

func TestBindNoModule(t *testing.T) {
	t.Parallel()
	rt := runner.New(io.Discard, runner.Options{RootFS: os.DirFS(t.TempDir())})
	stdlib.Register(rt)
	phpscript.Bind(rt, phpscript.Options{Migrator: &fakeMigrator{}, Dispatcher: &fakeDispatcher{}}, nil)

	for _, src := range []string{
		`<?php Task::register(new Task\Module);`,
		`<?php Task::migrate("schema/");`,
		`<?php Task::dispatch("x", "action/a.md");`,
	} {
		prog, err := rt.Load(src)
		assert.NoError(t, err, src)
		assert.ErrorIs(t, rt.Run(prog), phpscript.ErrNoModule, src)
	}
}
