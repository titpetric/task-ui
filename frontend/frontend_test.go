package frontend_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/killswitch"
	"github.com/titpetric/task-ui/markdown"
)

// TestMain bounds the binary. Every test here serves a whole front end
// over a database of its own, and a leak in one of them is paid for by
// every test after it: the run that noticed was an OOM kill with no
// stack and no test name on it.
//
// killswitch panics the first reading over 8GiB instead, which prints
// the goroutines and so says which test was running. The limit and the
// per-second trace are environment knobs; the package doc names them.
func TestMain(m *testing.M) {
	stop := killswitch.Watch()
	code := m.Run()
	stop()
	os.Exit(code)
}

// Every response of the pages carries the content security policy, the
// sign in screen and the assets included.
func TestRoutes_Policy(t *testing.T) {
	t.Parallel()
	h, err := frontend.New(frontend.Options{Guard: &access.Guard{}})
	assert.NoError(t, err)
	routes := h.Routes()
	for _, path := range []string{"/healthz", "/login", "/static/task-ui.css"} {
		rec := httptest.NewRecorder()
		routes.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		assert.Equal(t, http.StatusOK, rec.Code, path)
		assert.Equal(t, markdown.CSP, rec.Header().Get("Content-Security-Policy"), path)
	}
}

// catalogue is a fixed set of folders.
type catalogue map[string]*folder.Folder

func (c catalogue) Find(id string) *folder.Folder { return c[id] }

func (c catalogue) List() []*folder.Folder {
	ids := make([]string, 0, len(c))
	for id := range c {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]*folder.Folder, 0, len(ids))
	for _, id := range ids {
		result = append(result, c[id])
	}
	return result
}
