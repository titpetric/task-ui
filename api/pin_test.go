package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// memPins are pins in memory, one per user and target.
type memPins struct {
	mu   sync.Mutex
	pins []model.Pin
}

func (m *memPins) List(_ context.Context, user string) ([]model.Pin, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []model.Pin{}
	for _, pin := range m.pins {
		if pin.UserName == user {
			out = append(out, pin)
		}
	}
	return out, nil
}

func (m *memPins) Add(_ context.Context, pin *model.Pin) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, had := range m.pins {
		if had.UserName == pin.UserName && had.Folder == pin.Folder && had.Job == pin.Job && had.Dir == pin.Dir && had.Args == pin.Args {
			*pin = had
			return nil
		}
	}
	pin.ID = model.NewID()
	m.pins = append(m.pins, *pin)
	return nil
}

func (m *memPins) Delete(_ context.Context, user, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, pin := range m.pins {
		if pin.ID == id && pin.UserName == user {
			m.pins = append(m.pins[:i], m.pins[i+1:]...)
			return nil
		}
	}
	return storage.ErrNotFound
}

// pinCall sends a request to the pins routes as a user.
func pinCall(t *testing.T, h http.Handler, user *model.User, method, path, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if user != nil {
		req = req.WithContext(access.WithUser(req.Context(), user))
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	out := map[string]any{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

// A user pins a folder, a directory inside one or a job, sees the own
// pins alone, and unpins them; what is not there, or not seen, does not
// pin.
func TestPins(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Join(dir, "apps", "blog"), 0o755))
	malta := &folder.Folder{ID: "malta", Dir: dir}
	build := &folder.Job{Folder: "malta/dockerfiles", ID: "build"}
	build.Name = "build"
	push := &folder.Job{Folder: "malta/dockerfiles", ID: "push", TakesArgs: true}
	push.Name = "push"
	docker := &folder.Folder{ID: "malta/dockerfiles", Dir: t.TempDir(), Jobs: []*folder.Job{build, push}}
	pins := &memPins{}
	h := api.New(api.Options{Catalogue: catalogue{"malta": malta, "malta/dockerfiles": docker}, Pins: pins}).Routes()
	claude := &model.User{Name: "claude", Role: model.RoleRunner}
	black := &model.User{Name: "black", Role: model.RoleAdmin}

	code, out := pinCall(t, h, claude, http.MethodPost, "/pins", `{"folder":"malta/apps/blog"}`)
	assert.Equal(t, http.StatusCreated, code)
	folderPin, _ := out["id"].(string)
	assert.Equal(t, "malta/apps/blog", out["folder"])
	assert.Equal(t, []any{}, out["args"])

	code, out = pinCall(t, h, claude, http.MethodPost, "/pins", `{"folder":"malta/dockerfiles","job":"push","args":["--all"]}`)
	assert.Equal(t, http.StatusCreated, code)
	assert.Equal(t, []any{"--all"}, out["args"])
	code, out = pinCall(t, h, claude, http.MethodPost, "/pins", `{"folder":"malta/dockerfiles","job":"build"}`)
	assert.Equal(t, http.StatusCreated, code, out)

	for body, want := range map[string]int{
		`{"folder":"malta/nope"}`:                                       http.StatusNotFound,
		`{"folder":"gozo"}`:                                             http.StatusNotFound,
		`{"folder":"malta/dockerfiles","job":"nope"}`:                   http.StatusNotFound,
		`{"folder":"malta","dir":"apps"}`:                               http.StatusNotFound,
		`{"folder":"malta/dockerfiles","job":"build","args":["--all"]}`: http.StatusBadRequest,
		`nope`: http.StatusBadRequest,
	} {
		code, _ := pinCall(t, h, claude, http.MethodPost, "/pins", body)
		assert.Equal(t, want, code, body)
	}

	// A user who may not see the folder does not pin it.
	limited := &model.User{Name: "ops", Role: model.RoleRunner, Folders: `["malta"]`}
	code, _ = pinCall(t, h, limited, http.MethodPost, "/pins", `{"folder":"malta/dockerfiles","job":"build"}`)
	assert.Equal(t, http.StatusNotFound, code)

	code, out = pinCall(t, h, claude, http.MethodGet, "/pins", "")
	assert.Equal(t, http.StatusOK, code)
	assert.Len(t, out["pins"], 3)
	_, out = pinCall(t, h, black, http.MethodGet, "/pins", "")
	assert.Len(t, out["pins"], 0)

	// A pin is its user's to remove.
	code, _ = pinCall(t, h, black, http.MethodDelete, "/pins/"+folderPin, "")
	assert.Equal(t, http.StatusNotFound, code)
	code, _ = pinCall(t, h, claude, http.MethodDelete, "/pins/"+folderPin, "")
	assert.Equal(t, http.StatusNoContent, code)
	_, out = pinCall(t, h, claude, http.MethodGet, "/pins", "")
	assert.Len(t, out["pins"], 2)

	code, _ = pinCall(t, h, nil, http.MethodGet, "/pins", "")
	assert.Equal(t, http.StatusUnauthorized, code)
	code, _ = pinCall(t, api.New(api.Options{}).Routes(), claude, http.MethodGet, "/pins", "")
	assert.Equal(t, http.StatusNotFound, code)
}

// A pin request reads args as a run's.
func TestPinRequest(t *testing.T) {
	t.Parallel()
	var req api.PinRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"folder":"malta","job":"push","dir":"task","args":["-v"]}`), &req))
	assert.Equal(t, api.RunArgs{"-v"}, req.Args)
}

// A store of pins is what the handlers take.
func TestPinStore(t *testing.T) {
	t.Parallel()
	var store api.PinStore = &memPins{}
	assert.NoError(t, store.Add(context.Background(), &model.Pin{UserName: "claude", Folder: "malta"}))
}
