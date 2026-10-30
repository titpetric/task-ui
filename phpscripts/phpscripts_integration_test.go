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

	"github.com/titpetric/task-ui/artifactstore"
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
	folders  []string
}

func (f *fakeDispatcher) Dispatch(_ context.Context, m *phpscript.Module, comment, action, folder string) (string, error) {
	f.comments = append(f.comments, comment)
	f.actions = append(f.actions, m.Name()+":"+action)
	f.folders = append(f.folders, folder)
	return "01NOTE" + strings.ToUpper(action[:1]), nil
}

// instance extracts the shipped modules, wires them to a database and an
// artifact tree of their own and starts the manager over them, the way
// app does.
//
// The tree is a real one because a module reconciles against it: a row
// whose file is not there is dropped, so a test that writes rows and no
// files is a test of an empty folder.
func instance(t *testing.T) (*phpscript.Manager, storage.DB, *fakeDispatcher, *artifactstore.Store) {
	t.Helper()
	root := filepath.Join(t.TempDir(), "modules")
	assert.NoError(t, phpscripts.Extract(root))

	db, err := storage.OpenAndMigrate(context.Background(), ":memory:")
	assert.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	store := &artifactstore.Store{Dir: filepath.Join(t.TempDir(), "tree")}
	dispatcher := &fakeDispatcher{}
	var reported []error
	opts := phpscript.NewOptions(root)
	opts.Database = phpscript.NewDatabase(db)
	opts.Migrator = phpscript.NewSchema(db)
	opts.Dispatcher = dispatcher
	opts.Artifacts = phpscript.NewTree(store)
	opts.OnError = func(err error) { reported = append(reported, err) }

	m := phpscript.New(opts)
	assert.NoError(t, m.Start(context.Background()))
	assert.Len(t, reported, 0, "every shipped module registers")
	return m, db, dispatcher, store
}

// artifact writes one file into the tree, as a run's does.
func artifact(t *testing.T, store *artifactstore.Store, rel, body string) {
	t.Helper()
	_, err := store.Write(context.Background(), rel, []byte(body), "claude")
	assert.NoError(t, err)
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
	_, db, _, _ := instance(t)

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
	assert.Equal(t, 5, count)
}

// The first page is the form that starts a thread, and the thread list
// is empty until one does.
func TestRefinementIndex(t *testing.T) {
	t.Parallel()
	m, _, _, _ := instance(t)

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
	m, db, dispatcher, _ := instance(t)
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
	m, db, dispatcher, _ := instance(t)
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
	assert.Equal(t, 3, strings.Count(body, "rule</span>"), "a correction is a rule the next Apply all carries")
}

// The report a run wrote is the index the page shows, rendered as the
// instance's own markdown: a link resolves into the thread's folder of
// the artifact tree and an image is embedded, so a diagram is read on
// the page. The report's heading names the thread.
func TestRefinementReport(t *testing.T) {
	t.Parallel()
	m, _, _, store := instance(t)

	out := post(t, m, "/m/titpetric/refinement/threads", url.Values{"want": {"a schema"}})
	to := out.Header.Get("Location")
	id := filepath.Base(to)

	// A run writes into the thread's folder and names its files from
	// inside it. Nothing writes a row: the page reconciles the tree.
	artifact(t, store, id+"/README.md", "# Users and their login methods\n\n"+
		"The user and its login methods, kept apart.\n\n"+
		"- [schema/user.up.sql](schema/user.up.sql)\n\n"+
		"![The schema](diagrams/schema.svg)\n")
	artifact(t, store, id+"/schema/user.up.sql", "CREATE TABLE \"user\" (id CHAR(26) PRIMARY KEY);\n")
	artifact(t, store, id+"/diagrams/schema.svg", "<svg xmlns=\"http://www.w3.org/2000/svg\"></svg>\n")
	artifact(t, store, "PLAN.md", "An artifact of the instance, no thread's.\n")

	body := string(get(t, m, to).Body)
	assert.Contains(t, body, "The user and its login methods, kept apart.")
	assert.Contains(t, body, `href="/artifacts/files/`+id+`/schema/user.up.sql"`,
		"a relative link resolves into the thread's folder of the artifact tree")
	assert.Contains(t, body, `<img src="/artifacts/files/`+id+`/diagrams/schema.svg"`,
		"an image is embedded, so a diagram is read on the page")
	// The files are one link to the folder and not a listing: the
	// instance's artifacts view is what browses one.
	assert.Contains(t, body, `href="/artifacts?path=`+id+`"`,
		"the thread links the folder that holds its files")
	assert.NotContains(t, body, `class="badge mono"`,
		"and lists none of them itself: the artifacts view does that")
	assert.NotContains(t, body, "PLAN.md", "what the instance holds outside the folder is no file of the thread")

	assert.Contains(t, string(get(t, m, "/m/titpetric/refinement").Body),
		"Users and their login methods", "the report's heading names the thread")
}

// thread opens one with the corrections given, and answers its path and
// its id.
func thread(t *testing.T, m *phpscript.Manager, want string, corrections ...string) (string, string) {
	t.Helper()
	to := post(t, m, "/m/titpetric/refinement/threads", url.Values{"want": {want}}).Header.Get("Location")
	for _, body := range corrections {
		assert.Equal(t, http.StatusSeeOther,
			post(t, m, to+"/comments", url.Values{"comment": {body}}).Status, body)
	}
	return to, filepath.Base(to)
}

// Apply all sends the design and every rule as one request, over the
// artifacts that are already there, rather than one run per rule over
// what the run before it left.
//
// It is composed from the comments and not from its own last answer, so
// a second Apply all after another correction carries all of them again.
// That is the whole of the model: a rule stays a rule until it is
// deleted.
func TestRefinementApplyAll(t *testing.T) {
	t.Parallel()
	m, db, dispatcher, _ := instance(t)
	ctx := context.Background()

	to, id := thread(t, m, "A schema for users and their logins.",
		"Use ULID as primary keys.", "No foreign keys.", "Run mig lint.")

	assert.Equal(t, http.StatusSeeOther, post(t, m, to+"/apply", nil).Status)

	sent := dispatcher.comments[len(dispatcher.comments)-1]
	assert.Contains(t, sent, "## The design\n\nA schema for users and their logins.")
	assert.Contains(t, sent, "1. Use ULID as primary keys.")
	assert.Contains(t, sent, "2. No foreign keys.")
	assert.Contains(t, sent, "3. Run mig lint.")
	assert.Contains(t, sent, "SCRATCHPAD.md", "the run is pointed at the scratch pad before anything else")
	assert.Equal(t, []string{id}, []string{dispatcher.folders[len(dispatcher.folders)-1]},
		"the flattened run works in the thread's own folder")

	// Four comments stood before it and all four are folded into the one
	// it wrote; the apply comment itself stands.
	var folded, standing int
	assert.NoError(t, db.GetContext(ctx, &folded,
		"SELECT COUNT(*) FROM refinement_comment WHERE thread_id = ? AND collapsed_into <> ''", id))
	assert.Equal(t, 4, folded)
	assert.NoError(t, db.GetContext(ctx, &standing,
		"SELECT COUNT(*) FROM refinement_comment WHERE thread_id = ? AND collapsed_into = ''", id))
	assert.Equal(t, 1, standing)

	// An apply is neither a design nor a rule, so the next one composes
	// from the comments this one did, plus whatever arrived since.
	assert.Equal(t, http.StatusSeeOther,
		post(t, m, to+"/comments", url.Values{"comment": {"Comment every column."}}).Status)
	assert.Equal(t, http.StatusSeeOther, post(t, m, to+"/apply", nil).Status)

	sent = dispatcher.comments[len(dispatcher.comments)-1]
	assert.Contains(t, sent, "## The design\n\nA schema for users and their logins.")
	assert.Contains(t, sent, "1. Use ULID as primary keys.")
	assert.Contains(t, sent, "4. Comment every column.")
	assert.NotContains(t, sent, "## The rules\n\nEvery correction this thread has made, as a rule the artifacts have to\nsatisfy at once. A later rule wins where two disagree.\n\n1. ## The design",
		"a flattened request is never flattened into the next one")

	// The page folds them, and says how many.
	body := string(get(t, m, to).Body)
	assert.Contains(t, body, "folded into the request below")
	assert.Contains(t, body, `<span class="badge signal">applied</span>`)

	// A request is drawn as the markdown it was written as, and its
	// headings come down to the smallest one: a request is a paragraph
	// inside a thread, so its sections must not outrank the thread's.
	assert.Contains(t, body, `<h6 id="the-design">The design</h6>`)
	assert.Contains(t, body, `<h6 id="the-rules">The rules</h6>`)
	assert.NotContains(t, body, `<h2 id="the-design">`)
	assert.NotContains(t, body, "## The design", "the hashes are not printed at the reader")
	assert.Contains(t, body, "<li>Use ULID as primary keys.</li>", "and a numbered rule is a list item")
}

// A rule can be retired. Deleting the design retires the thread with it,
// because a rule corrects a design and there is nothing left to correct.
func TestRefinementDelete(t *testing.T) {
	t.Parallel()
	m, db, dispatcher, store := instance(t)
	ctx := context.Background()

	to, id := thread(t, m, "A schema for users and their logins.",
		"Use ULID as primary keys.", "Draw it in Comic Sans.")
	artifact(t, store, id+"/schema/user.up.sql", "CREATE TABLE \"user\" (id CHAR(26));\n")

	var rule string
	assert.NoError(t, db.GetContext(ctx, &rule,
		"SELECT id FROM refinement_comment WHERE thread_id = ? AND body LIKE 'Draw it%'", id))

	assert.Equal(t, http.StatusSeeOther, post(t, m, to+"/comments/"+rule+"/delete", nil).Status)

	var left int
	assert.NoError(t, db.GetContext(ctx, &left,
		"SELECT COUNT(*) FROM refinement_correction WHERE comment_id = ?", rule))
	assert.Equal(t, 0, left, "the correction row goes with the comment")
	assert.NotContains(t, string(get(t, m, to).Body), "Comic Sans")

	// What is gone is not carried.
	assert.Equal(t, http.StatusSeeOther, post(t, m, to+"/apply", nil).Status)
	sent := dispatcher.comments[len(dispatcher.comments)-1]
	assert.Contains(t, sent, "1. Use ULID as primary keys.")
	assert.NotContains(t, sent, "Comic Sans")

	// The design takes the thread with it, and leaves the files alone.
	var design string
	assert.NoError(t, db.GetContext(ctx, &design,
		"SELECT id FROM refinement_comment WHERE thread_id = ? AND body LIKE 'A schema%'", id))

	out := post(t, m, to+"/comments/"+design+"/delete", nil)
	assert.Equal(t, http.StatusSeeOther, out.Status)
	assert.Equal(t, "/m/titpetric/refinement", out.Header.Get("Location"))

	var threads, comments, corrections int
	assert.NoError(t, db.GetContext(ctx, &threads, "SELECT COUNT(*) FROM refinement_thread WHERE id = ?", id))
	assert.NoError(t, db.GetContext(ctx, &comments, "SELECT COUNT(*) FROM refinement_comment WHERE thread_id = ?", id))
	assert.NoError(t, db.GetContext(ctx, &corrections, "SELECT COUNT(*) FROM refinement_correction WHERE thread_id = ?", id))
	assert.Equal(t, 0, threads)
	assert.Equal(t, 0, comments)
	assert.Equal(t, 0, corrections)

	assert.Equal(t, http.StatusNotFound, get(t, m, to).Status)
	assert.Equal(t, "CREATE TABLE \"user\" (id CHAR(26));\n",
		phpscript.NewTree(store).Read(id+"/schema/user.up.sql"),
		"the files stay in the artifact tree")
}

// A thread that is not there is a 404, and a form with nothing in it is
// a 400 that says what it needs.
func TestRefinementRefuses(t *testing.T) {
	t.Parallel()
	m, _, dispatcher, _ := instance(t)

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

	// A correction that reaches the route that opens threads opens none:
	// it would bury the correction in a thread with none of the work it
	// was correcting.
	out = post(t, m, "/m/titpetric/refinement/threads", url.Values{"comment": {"use ULID as primary keys"}})
	assert.Equal(t, http.StatusBadRequest, out.Status)
	assert.Contains(t, string(out.Body), "needs its thread")
	assert.Len(t, dispatcher.comments, 1, "and starts no run")
	assert.Equal(t, 1, strings.Count(string(get(t, m, "/m/titpetric/refinement").Body), `class="key"`),
		"the one thread is the one that was asked for")
}

// A module's public tree is served as bytes, with its content type.
func TestRefinementAssets(t *testing.T) {
	t.Parallel()
	m, _, _, _ := instance(t)

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
