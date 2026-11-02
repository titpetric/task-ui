package phpfront_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpfront"
)

// host answers what the real one answers off the instance.
type host struct {
	published []string
}

func (h *host) Whoami(context.Context) string { return "ops" }

func (h *host) Folders(context.Context) []phpfront.Folder {
	return []phpfront.Folder{{ID: "infra", Title: "Infra", Dir: "/src/infra", Jobs: []string{"build", "test"}}}
}

func (h *host) Running(context.Context) []phpfront.Running {
	return []phpfront.Running{{Key: "infra/build", ID: "01hz", State: "running"}}
}

func (h *host) Dirty(_ context.Context, dir string) int {
	if dir == "/src/infra" {
		return 3
	}
	return 0
}

func (h *host) Rules(context.Context) phpfront.Rules {
	return phpfront.Rules{
		Rules:       []phpfront.Rule{{Key: "view", Title: "Allow View"}},
		Sections:    []phpfront.Rule{{Key: "note", Title: "Notes"}},
		Permissions: []phpfront.Rule{{Key: "run", Title: "Run Jobs"}},
		Paths:       []string{"note/all", "note/view", "run"},
		Checks:      []string{"note/view", "run"},
	}
}

func (h *host) Publish(_ context.Context, topic string, _ map[string]any) error {
	h.published = append(h.published, topic)
	return nil
}

// Every host call answers across the boundary, and a Go struct arrives as a
// map keyed by its json tags.
func TestBind(t *testing.T) {
	t.Parallel()

	own := &host{}
	body := serve(t, own, "/host")

	// A Go struct arrives as an associative array keyed by its json tags, in
	// declaration order, which is what a template indexes by.
	for _, want := range []string{
		"base=/development",
		"version=test",
		"whoami=ops",
		"dirty=3",
		`folders=[{"id":"infra","title":"Infra","dir":"/src/infra","jobs":["build","test"]}]`,
		`running=[{"key":"infra/build","id":"01hz","state":"running"}]`,
		`"checks":["note/view","run"]`,
		`"sections":[{"key":"note","title":"Notes"}]`,
		"published=true",
		// Indexing by key is what a template and the DAOs do, and what a
		// host object silently answers nothing for.
		"indexed=true",
		`keys=["rules","sections","permissions","paths","checks","extras"]`,
		"folder=infra jobs=build,test",
		"section=note:Notes",
		"checks=note/view,run",
	} {
		assert.Contains(t, body, want)
	}
	assert.Contains(t, body, `<h1 id="title">Title</h1>`, "markdown renders through the instance's renderer")
	assert.Equal(t, []string{"runs"}, own.published)
}

// A nil Host answers every call empty, which is an instance with no chrome and
// no panic.
func TestBind_NoHost(t *testing.T) {
	t.Parallel()

	body := serve(t, nil, "/host")
	assert.Contains(t, body, "whoami=\n")
	assert.Contains(t, body, "dirty=0")
	assert.Contains(t, body, "folders=[]", "an empty array, so a foreach over it draws the empty line")
	assert.Contains(t, body, `rules={"rules":[],"sections":[],"permissions":[],"paths":[],"checks":[],"extras":[]}`)
	assert.Contains(t, body, "published=false")
}

// serve runs one request against the host fixture tree.
func serve(t *testing.T, own phpfront.Host, path string) string {
	t.Helper()

	opts := phpfront.NewOptions("testdata/host")
	opts.Base = "/development"
	opts.Version = "test"
	opts.Host = own
	opts.OnError = func(err error) { t.Errorf("router: %v", err) }

	front, err := phpfront.New(opts)
	assert.NoError(t, err)

	router, err := front.Routes(context.Background())
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	return strings.ReplaceAll(w.Body.String(), "\r\n", "\n")
}
