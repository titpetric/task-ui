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

func (m *memPins) MoveFolder(_ context.Context, from, to string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	moved := 0
	for i := range m.pins {
		if folder, ok := model.Moved(m.pins[i].Folder, from, to); ok {
			m.pins[i].Folder = folder
			moved++
		}
	}
	return moved, nil
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

// A store of pins is what the handlers take; a project that moves
// takes the pins of the folders inside it along, and its answer counts
// them.
func TestPinStore(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var store api.PinStore = &memPins{}
	assert.NoError(t, store.Add(ctx, &model.Pin{UserName: "claude", Folder: "malta"}))
	assert.NoError(t, store.Add(ctx, &model.Pin{UserName: "claude", Folder: "malta/apps", Job: "build"}))
	assert.NoError(t, store.Add(ctx, &model.Pin{UserName: "black", Folder: "maltese"}))
	moved, err := store.MoveFolder(ctx, "malta", "acme/malta")
	assert.NoError(t, err)
	assert.Equal(t, 2, moved)
	pins, err := store.List(ctx, "claude")
	assert.NoError(t, err)
	assert.Equal(t, "acme/malta", pins[0].Folder)
	assert.Equal(t, "acme/malta/apps", pins[1].Folder)

	data, err := json.Marshal(api.ProjectMoved{Project: &model.Project{Folder: "acme/malta"}, Pins: moved, Left: []model.FolderRef{}})
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"pins":2`)
	assert.Contains(t, string(data), `"left":[]`)
}

// A pin resolves its folder through the steps' catalogue: a folder the
// project settings hide from the listings takes a pin, on its job or
// on the folder itself.
func TestPins_Hidden(t *testing.T) {
	t.Parallel()
	malta := &folder.Folder{ID: "malta", Dir: t.TempDir()}
	deploy := &folder.Job{Folder: "malta/apps/blog", ID: "deploy"}
	deploy.Name = "deploy"
	blog := &folder.Folder{ID: "malta/apps/blog", Dir: t.TempDir(), Jobs: []*folder.Job{deploy}}
	h := api.New(api.Options{Catalogue: catalogue{"malta": malta}, Steps: catalogue{"malta": malta, "malta/apps/blog": blog}, Pins: &memPins{}}).Routes()
	claude := &model.User{Name: "claude", Role: model.RoleRunner}

	code, out := pinCall(t, h, claude, http.MethodPost, "/pins", `{"folder":"malta/apps/blog","job":"deploy"}`)
	assert.Equal(t, http.StatusCreated, code, out)
	code, out = pinCall(t, h, claude, http.MethodPost, "/pins", `{"folder":"malta/apps/blog"}`)
	assert.Equal(t, http.StatusCreated, code, out)
	assert.Equal(t, "malta/apps/blog", out["folder"])

	limited := &model.User{Name: "guest", Role: model.RoleRunner, Folders: model.EncodeList([]string{"gozo"})}
	code, _ = pinCall(t, h, limited, http.MethodPost, "/pins", `{"folder":"malta/apps/blog","job":"deploy"}`)
	assert.Equal(t, http.StatusNotFound, code, "the user's folders still bound it")
}
