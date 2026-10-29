package phpscript_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

// render drives one GET through the Manager.
func render(t *testing.T, m *phpscript.Manager, target string) (phpscript.Result, error) {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, target, nil)
	w := httptest.NewRecorder()
	return m.Render(context.Background(), w, r)
}

func TestRoute(t *testing.T) {
	t.Parallel()
	for target, want := range map[string][2]string{
		"/m/titpetric/refinement":               {"titpetric/refinement", ""},
		"/m/titpetric/refinement/":              {"titpetric/refinement", ""},
		"/m/titpetric/refinement/threads/01J":   {"titpetric/refinement", "threads/01J"},
		"/m/titpetric/refinement/assets/one.js": {"titpetric/refinement", "assets/one.js"},
		"/m/acme/notes/a/b/c":                   {"acme/notes", "a/b/c"},
	} {
		name, rest, err := phpscript.Route(target)
		assert.NoError(t, err, target)
		assert.Equal(t, want[0], name, target)
		assert.Equal(t, want[1], rest, target)
	}

	for _, bad := range []string{"/m", "/m/", "/m/titpetric", "/m/Titpetric/X", "/m/../../etc"} {
		_, _, err := phpscript.Route(bad)
		assert.ErrorIs(t, err, phpscript.ErrNotFound, bad)
	}
}

func TestRender(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	module(t, root, "titpetric/refinement", `<?php
echo "<p>uri=", $_SERVER["REQUEST_URI"], "</p>";
echo "<p>module=", $_SERVER["TASK_UI_MODULE"], " base=", $_SERVER["TASK_UI_MODULE_BASE"], "</p>";
echo "<p>method=", $_SERVER["REQUEST_METHOD"], " q=", isset($_GET["q"]) ? $_GET["q"] : "", "</p>";
`)

	m := phpscript.New(phpscript.NewOptions(root))
	assert.NoError(t, m.Start(context.Background()))

	out, err := render(t, m, "/m/titpetric/refinement/threads/01J?q=one")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, out.Status)
	assert.False(t, out.File)
	assert.Greater(t, int64(out.Took), int64(0), "the render reports how long it took")

	body := string(out.Body)
	assert.Contains(t, body, "uri=/m/titpetric/refinement/threads/01J?q=one",
		"REQUEST_URI holds the path the browser asked for, prefix and query included")
	assert.Contains(t, body, "module=titpetric/refinement base=/m/titpetric/refinement")
	assert.Contains(t, body, "method=GET q=one")
}

func TestRenderPost(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	module(t, root, "titpetric/refinement", `<?php echo "want=", isset($_POST["want"]) ? $_POST["want"] : "none";`)

	m := phpscript.New(phpscript.NewOptions(root))
	assert.NoError(t, m.Start(context.Background()))

	r := httptest.NewRequest(http.MethodPost, "/m/titpetric/refinement",
		strings.NewReader("want=a+schema"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	out, err := m.Render(context.Background(), httptest.NewRecorder(), r)
	assert.NoError(t, err)
	assert.Equal(t, "want=a schema", string(out.Body))
}

func TestRenderFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dir := module(t, root, "titpetric/refinement", "<?php echo 'the page';")
	write(t, filepath.Join(dir, "public", "assets", "refinement.js"), "window.refinement = 1;\n")
	write(t, filepath.Join(dir, "secret.txt"), "not served")

	m := phpscript.New(phpscript.NewOptions(root))
	assert.NoError(t, m.Start(context.Background()))

	out, err := render(t, m, "/m/titpetric/refinement/assets/refinement.js")
	assert.NoError(t, err)
	assert.True(t, out.File)
	assert.Equal(t, "window.refinement = 1;\n", string(out.Body))
	assert.Contains(t, out.Header.Get("Content-Type"), "javascript")

	// A path outside public reaches index.php, never the file beside it.
	out, err = render(t, m, "/m/titpetric/refinement/secret.txt")
	assert.NoError(t, err)
	assert.False(t, out.File)
	assert.Equal(t, "the page", string(out.Body))

	// A directory under public is not a file either.
	out, err = render(t, m, "/m/titpetric/refinement/assets")
	assert.NoError(t, err)
	assert.False(t, out.File)
}

func TestRenderStatus(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	module(t, root, "titpetric/refinement", `<?php
http_response_code(404);
echo "no thread there";
`)

	m := phpscript.New(phpscript.NewOptions(root))
	assert.NoError(t, m.Start(context.Background()))

	out, err := render(t, m, "/m/titpetric/refinement/threads/gone")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, out.Status, "the script's own status reaches the handler")
	assert.Equal(t, "no thread there", string(out.Body))
}

func TestRenderRefuses(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	module(t, root, "titpetric/thrower", `<?php throw new Exception("the thread is gone");`)
	// A module with no index.php registers and serves nothing.
	module(t, root, "titpetric/empty", "")

	m := phpscript.New(phpscript.NewOptions(root))
	assert.NoError(t, m.Start(context.Background()))

	_, err := render(t, m, "/m/titpetric/nothing")
	assert.ErrorIs(t, err, phpscript.ErrNotFound)

	_, err = render(t, m, "/m/titpetric/empty")
	assert.ErrorContains(t, err, phpscript.Index)

	out, err := render(t, m, "/m/titpetric/thrower")
	assert.ErrorContains(t, err, "the thread is gone")
	assert.Greater(t, int64(out.Took), int64(0), "a render that threw still reports its time")
}

func TestRenderPool(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	module(t, root, "titpetric/counter", `<?php
$n = 0;
for ($i = 0; $i < 3; $i++) { $n = $n + $i; }
echo "n=", $n;
`)

	opts := phpscript.NewOptions(root)
	opts.PoolSize = 1
	m := phpscript.New(opts)
	assert.NoError(t, m.Start(context.Background()))

	// The same runtime answers both, reset between them, so the second
	// request sees nothing the first defined.
	for range 2 {
		out, err := render(t, m, "/m/titpetric/counter")
		assert.NoError(t, err)
		assert.Equal(t, "n=3", string(out.Body))
	}
}
