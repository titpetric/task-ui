//go:build integration

package phpscripts_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
	"github.com/titpetric/task-ui/phpscripts"
	"github.com/titpetric/task-ui/storage"
	_ "modernc.org/sqlite"
)

// refinementProject is what mig records the module's migrations under:
// the plugin cut to seven characters and eight hex of the name's digest,
// because the column holds sixteen.
var refinementProject = (&phpscript.Module{Vendor: "titpetric", Plugin: "refinement"}).Project()

// fakeDispatcher stands in for the runs: it records what a comment asked
// for and answers with a note identifier, so the thread's pages can be
// read without a Claude session.
type fakeDispatcher struct {
	comments []string
	actions  []string
}

func (f *fakeDispatcher) Dispatch(_ context.Context, m *phpscript.Module, comment, action string) (string, error) {
	f.comments = append(f.comments, comment)
	f.actions = append(f.actions, m.Name()+":"+action)
	return "01NOTE" + strings.ToUpper(action[:1]), nil
}

// instance extracts the shipped modules, wires them to a database of
// their own and starts the manager over them, the way app does.
func instance(t *testing.T) (*phpscript.Manager, storage.DB, *fakeDispatcher) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "modules")
	assert.NoError(t, phpscripts.Extract(root))

	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	dispatcher := &fakeDispatcher{}
	var reported []error
	opts := phpscript.NewOptions(root)
	opts.Database = phpscript.NewDatabase(db)
	opts.Migrator = phpscript.NewSchema(db)
	opts.Dispatcher = dispatcher
	opts.OnError = func(err error) { reported = append(reported, err) }

	m := phpscript.New(opts)
	assert.NoError(t, m.Start(context.Background()))
	assert.Len(t, reported, 0, "every shipped module registers")
	return m, db, dispatcher
}

// get drives one request through a module.
func get(t *testing.T, m *phpscript.Manager, target string) phpscript.Result {
	t.Helper()
	out, err := m.Render(context.Background(),
		httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, target, nil))
	assert.NoError(t, err, target)
	return out
}

// post drives one form through a module.
func post(t *testing.T, m *phpscript.Manager, target string, form url.Values) phpscript.Result {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	out, err := m.Render(context.Background(), httptest.NewRecorder(), r)
	assert.NoError(t, err, target)
	return out
}

// The module's schema is applied at startup, by migrate.php, which runs
// because it carries @startup and not because anything requires it.
func TestRefinementMigrates(t *testing.T) {
	t.Parallel()
	_, db, _ := instance(t)

	for _, table := range []string{"refinement_thread", "refinement_comment", "refinement_artifact", "refinement_correction"} {
		var name string
		assert.NoError(t, db.GetContext(context.Background(), &name,
			"SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", table), table)
		assert.Equal(t, table, name)
	}

	// Under a project of its own name, so a module added later starts
	// from its own first migration.
	var count int
	assert.NoError(t, db.GetContext(context.Background(), &count,
		"SELECT COUNT(*) FROM migrations WHERE project = ?", refinementProject))
	assert.Equal(t, 4, count)
}

// The first page is the form that starts a thread, and the thread list
// is empty until one does.
func TestRefinementIndex(t *testing.T) {
	t.Parallel()
	m, _, _ := instance(t)

	out := get(t, m, "/m/titpetric/refinement")
	body := string(out.Body)
	assert.Equal(t, http.StatusOK, out.Status)
	assert.Contains(t, body, `action="/m/titpetric/refinement/threads"`)
	assert.Contains(t, body, "Start a thread")
	assert.Contains(t, body, "No threads yet.")
}

// Starting a thread writes the thread and its first comment, dispatches
// the run, and answers with a redirect so a reload starts no second one.
func TestRefinementStart(t *testing.T) {
	t.Parallel()
	m, db, dispatcher := instance(t)
	ctx := context.Background()

	want := "A database schema with strict separation of user, user_login_method and user_login_method_password."
	out := post(t, m, "/m/titpetric/refinement/threads", url.Values{"want": {want}})
	assert.Equal(t, http.StatusSeeOther, out.Status, "a form that changed something redirects")

	to := out.Header.Get("Location")
	assert.True(t, strings.HasPrefix(to, "/m/titpetric/refinement/threads/"), to)

	assert.Equal(t, []string{want}, dispatcher.comments, "the comment reaches the run as it was written")
	assert.Equal(t, []string{"titpetric/refinement:action/refinement.md"}, dispatcher.actions)

	type thread struct {
		ID    string `db:"id"`
		Title string `db:"title"`
	}
	var row thread
	assert.NoError(t, db.GetContext(ctx, &row, "SELECT id, title FROM refinement_thread"))
	assert.Equal(t, 26, len(row.ID), "a ULID primary key")
	assert.Contains(t, row.Title, "strict separation")

	var note string
	assert.NoError(t, db.GetContext(ctx, &note, "SELECT note_id FROM refinement_comment"))
	assert.NotEqual(t, "", note, "the comment holds the note the dispatch opened")

	// The first comment is no correction: it has nothing to correct.
	var corrections int
	assert.NoError(t, db.GetContext(ctx, &corrections, "SELECT COUNT(*) FROM refinement_correction"))
	assert.Equal(t, 0, corrections)

	// The thread lists, and its page reads.
	assert.Contains(t, string(get(t, m, "/m/titpetric/refinement").Body), row.Title)
	assert.Contains(t, string(get(t, m, to).Body), "Refine it")
}

// Every comment after the first is a correction over what the run before
// it produced, and each one starts a run of its own.
func TestRefinementCorrections(t *testing.T) {
	t.Parallel()
	m, db, dispatcher := instance(t)
	ctx := context.Background()

	out := post(t, m, "/m/titpetric/refinement/threads", url.Values{"want": {"a schema"}})
	to := out.Header.Get("Location")

	corrections := []string{
		"use ULID as primary keys",
		"do not use foreign keys",
		"run mig lint on the schema and fix violations",
	}
	for _, body := range corrections {
		out = post(t, m, to+"/comments", url.Values{"comment": {body}})
		assert.Equal(t, http.StatusSeeOther, out.Status, body)
	}

	assert.Equal(t, append([]string{"a schema"}, corrections...), dispatcher.comments,
		"each correction starts a run of its own")

	var count int
	assert.NoError(t, db.GetContext(ctx, &count, "SELECT COUNT(*) FROM refinement_comment"))
	assert.Equal(t, 4, count)

	assert.NoError(t, db.GetContext(ctx, &count, "SELECT COUNT(*) FROM refinement_correction"))
	assert.Equal(t, 3, count, "three corrections over the first comment")

	// Each correction names the comment before it.
	var corrects []string
	assert.NoError(t, db.SelectContext(ctx, &corrects,
		"SELECT corrects_id FROM refinement_correction ORDER BY created_at, id"))
	for _, id := range corrects {
		assert.Equal(t, 26, len(id), "a correction names the comment it changed")
	}

	// The page reads every comment, in order, the corrections marked.
	body := string(get(t, m, to).Body)
	last := -1
	for _, want := range append([]string{"a schema"}, corrections...) {
		at := strings.Index(body, want)
		assert.True(t, at > last, want)
		last = at
	}
	assert.Equal(t, 3, strings.Count(body, "correction</span>"))
}

// The report a run wrote is the index the page shows, rendered as the
// instance's own markdown: a link resolves into the artifact tree and an
// image is embedded, so a diagram is read on the page.
func TestRefinementReport(t *testing.T) {
	t.Parallel()
	m, db, _ := instance(t)
	ctx := context.Background()

	out := post(t, m, "/m/titpetric/refinement/threads", url.Values{"want": {"a schema"}})
	to := out.Header.Get("Location")
	id := filepath.Base(to)

	report := "The user and its login methods, kept apart.\n\n" +
		"- [schema/user.up.sql](schema/user.up.sql)\n\n" +
		"![The schema](diagrams/schema.svg)\n"
	_, err := db.ExecContext(ctx, "UPDATE refinement_thread SET report_markdown = ? WHERE id = ?", report, id)
	assert.NoError(t, err)
	_, err = db.ExecContext(ctx,
		"INSERT INTO refinement_artifact (id, thread_id, comment_id, path, kind, created_at) VALUES (?, ?, '', ?, ?, ?)",
		"01ARTIFACT0000000000000001", id, "diagrams/schema.svg", "image", "2026-10-30 00:00:00")
	assert.NoError(t, err)

	body := string(get(t, m, to).Body)
	assert.Contains(t, body, "The user and its login methods, kept apart.")
	assert.Contains(t, body, `href="/artifacts/files/schema/user.up.sql"`,
		"a relative link resolves into the artifact tree")
	assert.Contains(t, body, `<img src="/artifacts/files/diagrams/schema.svg"`,
		"an image is embedded, so a diagram is read on the page")
	assert.Contains(t, body, "diagrams/schema.svg</a>", "the artifact is listed beside the report")
}

// A thread that is not there is a 404, and a form with nothing in it is
// a 400 that says what it needs.
func TestRefinementRefuses(t *testing.T) {
	t.Parallel()
	m, _, dispatcher := instance(t)

	out := get(t, m, "/m/titpetric/refinement/threads/01NOTATHREAD00000000000000")
	assert.Equal(t, http.StatusNotFound, out.Status)
	assert.Contains(t, string(out.Body), "No thread there.")

	out = post(t, m, "/m/titpetric/refinement/threads", url.Values{"want": {"   "}})
	assert.Equal(t, http.StatusBadRequest, out.Status)
	assert.Contains(t, string(out.Body), "one paragraph")
	assert.Len(t, dispatcher.comments, 0, "nothing was dispatched")

	out = post(t, m, "/m/titpetric/refinement/threads", url.Values{"want": {"a schema"}})
	to := out.Header.Get("Location")
	out = post(t, m, to+"/comments", url.Values{"comment": {""}})
	assert.Equal(t, http.StatusBadRequest, out.Status)
	assert.Contains(t, string(out.Body), "one thing to change")
	assert.Len(t, dispatcher.comments, 1, "the empty correction started no run")
}

// A module's public tree is served as bytes, with its content type.
func TestRefinementAssets(t *testing.T) {
	t.Parallel()
	m, _, _ := instance(t)

	root := filepath.Dir(filepath.Dir(mustRoot(t, m)))
	entries, err := os.ReadDir(filepath.Join(root, "titpetric", "refinement", "public", "assets"))
	assert.NoError(t, err)
	assert.Greater(t, len(entries), 0, "the module ships at least one asset")

	for _, entry := range entries {
		out := get(t, m, "/m/titpetric/refinement/assets/"+entry.Name())
		assert.True(t, out.File, entry.Name())
		assert.NotEqual(t, "", out.Header.Get("Content-Type"), entry.Name())
	}
}

// mustRoot is the directory the refinement module was found in.
func mustRoot(t *testing.T, m *phpscript.Manager) string {
	t.Helper()
	module, ok := m.Module("titpetric/refinement")
	assert.True(t, ok)
	return module.Root
}
