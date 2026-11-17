package assets_test

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/assets"
)

// revision is the ULID a test tree answers under.
const revision = "01jx8z5j6k7m8n9p0q1r2s3t4v"

// tree is a theme of two files and a module of one.
func tree() *assets.Tree {
	return assets.New(map[string]fs.FS{
		assets.Theme: fstest.MapFS{
			assets.Stylesheet: {Data: []byte("@import url(\"css/base.css\");\n@import url(\"css/forms.css\");\n")},
			"css/base.css":    {Data: []byte("body{}")},
			"css/forms.css":   {Data: []byte("form{}")},
			"task-ui.js":      {Data: []byte("//")},
		},
		"notes": fstest.MapFS{"notes.js": {Data: []byte("// notes")}},
	}, revision)
}

// TestTree_Revision returns the identifier the tree was built with.
func TestTree_Revision(t *testing.T) {
	t.Parallel()

	assert.Equal(t, revision, tree().Revision())
	var none *assets.Tree
	assert.Equal(t, "", none.Revision())
}

// TestTree_Href returns the tree's URL under its revision.
func TestTree_Href(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "/assets/r_"+revision, tree().Href())
	var none *assets.Tree
	assert.Equal(t, "/assets", none.Href())
}

// The version is a short hash of every file, the same every call.
func TestVersion(t *testing.T) {
	t.Parallel()

	v := tree().Version()
	assert.Len(t, v, 12)
	assert.Equal(t, v, tree().Version())
}

// A changed file changes the version, which is the whole point of it.
func TestVersion_Changes(t *testing.T) {
	t.Parallel()

	other := assets.New(map[string]fs.FS{
		assets.Theme: fstest.MapFS{"task-ui.js": {Data: []byte("// other")}},
	}, revision)
	assert.NotEqual(t, tree().Version(), other.Version())
}

// A nil tree answers as one with nothing in it.
func TestNil(t *testing.T) {
	t.Parallel()

	var none *assets.Tree
	assert.Equal(t, "", none.Version())
	assert.Empty(t, none.CSS())
}

// Every import of the entry stylesheet carries the version.
func TestCSS(t *testing.T) {
	t.Parallel()

	tr := tree()
	css := string(tr.CSS())
	assert.Contains(t, css, `@import url("css/base.css?v=`+tr.Version()+`")`)
	assert.Contains(t, css, `@import url("css/forms.css?v=`+tr.Version()+`")`)
	assert.NotContains(t, css, `"css/base.css"`)
}

// A theme without a stylesheet is a tree without CSS, not a panic.
func TestCSS_Missing(t *testing.T) {
	t.Parallel()

	assert.Empty(t, assets.New(nil, revision).CSS())
}

func TestServeHTTP(t *testing.T) {
	t.Parallel()

	tr := tree()
	for _, tc := range []struct {
		name   string
		path   string
		status int
		want   string
	}{
		{"a theme file", assets.Prefix + "theme/task-ui.js", http.StatusOK, "//"},
		{"a nested theme file", assets.Prefix + "theme/css/base.css", http.StatusOK, "body{}"},
		{"a module file", assets.Prefix + "notes/notes.js", http.StatusOK, "// notes"},
		{"the stylesheet is the versioned one", assets.Prefix + "theme/" + assets.Stylesheet, http.StatusOK, "?v=" + tr.Version()},
		{"a revisioned theme file", assets.Prefix + "r_" + revision + "/theme/task-ui.js", http.StatusOK, "//"},
		{"a revisioned nested theme file", assets.Prefix + "r_" + revision + "/theme/css/base.css", http.StatusOK, "body{}"},
		{"a revisioned module file", assets.Prefix + "r_" + revision + "/modules/notes/notes.js", http.StatusOK, "// notes"},
		{"the revisioned stylesheet", assets.Prefix + "r_" + revision + "/theme/" + assets.Stylesheet, http.StatusOK, "?v=" + tr.Version()},
		{"another revision", assets.Prefix + "r_other/theme/task-ui.js", http.StatusNotFound, ""},
		{"an unknown segment", assets.Prefix + "nope/task-ui.js", http.StatusNotFound, ""},
		{"a segment naming no file", assets.Prefix + "theme/", http.StatusNotFound, ""},
		{"a file that is not there", assets.Prefix + "theme/missing.js", http.StatusNotFound, ""},
		{"a path leaving the root", assets.Prefix + "theme/../../go.mod", http.StatusNotFound, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			tr.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tc.path, nil))
			assert.Equal(t, tc.status, w.Code)
			if tc.want != "" {
				assert.Contains(t, w.Body.String(), tc.want)
			}
		})
	}
}

// Open reads one file of a segment, and refuses a path leaving the root.
func TestTree_Open(t *testing.T) {
	t.Parallel()

	tr := tree()
	data, err := tr.Open(assets.Theme, "css/base.css")
	assert.NoError(t, err)
	assert.Equal(t, "body{}", string(data))

	_, err = tr.Open("nope", "css/base.css")
	assert.Error(t, err)
	_, err = tr.Open(assets.Theme, "../go.mod")
	assert.Error(t, err)
	_, err = tr.Open(assets.Theme, "missing.css")
	assert.Error(t, err)
}

// A served file is cacheable for an hour and carries its type.
func TestServeHTTP_Headers(t *testing.T) {
	t.Parallel()

	w := httptest.NewRecorder()
	tree().ServeHTTP(w, httptest.NewRequest(http.MethodGet, assets.Prefix+"theme/css/base.css", nil))
	assert.Equal(t, assets.MaxAge, w.Header().Get("Cache-Control"))
	assert.Contains(t, w.Header().Get("Content-Type"), "text/css")
}

// Dirs maps the theme and every module that ships assets, and nothing else.
func TestDirs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	for _, dir := range []string{
		filepath.Join(root, "theme", "assets"),
		filepath.Join(root, "modules", "notes", "assets"),
		filepath.Join(root, "modules", "runs", "templates"),
	} {
		assert.NoError(t, os.MkdirAll(dir, 0o755))
	}

	segments := assets.Dirs(root)
	assert.Contains(t, segments, assets.Theme)
	assert.Contains(t, segments, "notes")
	assert.NotContains(t, segments, "runs")
	assert.Len(t, segments, 2)
}

// A root holding nothing is a tree without segments, not a failure.
func TestDirs_Empty(t *testing.T) {
	t.Parallel()

	assert.Empty(t, assets.Dirs(t.TempDir()))
}
