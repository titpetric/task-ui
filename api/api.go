// Package api implements the JSON endpoints of task-ui: runs and their
// frames, artifacts, schedules, projects, users and secrets, and the
// websocket a window attaches to a run with.
//
// Handlers decode, check permissions, call storage and the runner, map
// errors to status codes and encode. The user comes from the context,
// put there by access.Guard, which the composition mounts in front.
package api

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"sort"
	"strconv"

	chi "github.com/go-chi/chi/v5"

	"github.com/titpetric/task-ui/hours"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// Handlers serves the API.
type Handlers struct {
	opts Options
}

// New returns the handlers over the options.
func New(opts Options) *Handlers {
	return &Handlers{opts: opts}
}

// Routes returns the router, to mount under /api. Every route expects
// the user on the context.
func (h *Handlers) Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/runs", h.listRuns)
	r.Post("/runs", h.startRun)
	r.Post("/runs/all", h.startAll)
	r.Get("/runs/{id}", h.getRun)
	r.Post("/runs/{id}/stop", h.stopRun)
	r.Post("/runs/{id}/kill", h.killRun)
	r.Get("/runs/{id}/frame.svg", h.frameSVG)
	r.Get("/runs/{id}/frame.txt", h.frameText)
	r.Get("/runs/{id}/cast", h.cast)
	r.Handle("/runs/{id}/attach", h.attach())

	r.Get("/artifacts", h.listArtifacts)
	r.Get("/artifacts/{id}", h.downloadArtifact)
	r.Delete("/artifacts/{id}", h.deleteArtifact)

	r.Get("/schedules", h.listSchedules)
	r.Post("/schedules", h.createSchedule)
	r.Get("/schedules/{id}", h.getSchedule)
	r.Put("/schedules/{id}", h.updateSchedule)
	r.Delete("/schedules/{id}", h.deleteSchedule)

	r.Get("/projects", h.listProjects)
	r.Post("/projects", h.cloneProject)
	r.Get("/projects/{folder}", h.getProject)
	r.Put("/projects/{folder}", h.updateProject)
	r.Delete("/projects/{folder}", h.deleteProject)

	r.Get("/users", h.listUsers)
	r.Post("/users", h.createUser)
	r.Put("/users/{id}", h.updateUser)
	r.Delete("/users/{id}", h.deleteUser)

	r.Get("/secrets", h.listSecrets)
	r.Post("/secrets", h.setSecret)
	r.Delete("/secrets/{id}", h.deleteSecret)

	r.Get("/keys", h.listKeys)
	r.Post("/keys/generate", h.generateKey)
	r.Post("/keys/upload", h.uploadKey)
	r.Delete("/keys/{name}", h.deleteKey)

	r.Get("/env/{folder}", h.listEnv)
	r.Get("/env/{folder}/*", h.getEnv)
	r.Put("/env/{folder}/*", h.putEnv)

	return r
}

// param returns what the route bound under name.
func param(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

// maxBody bounds a request body.
const maxBody = 1 << 20

// readJSON decodes the body into v.
func readJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	data, err := io.ReadAll(io.LimitReader(r.Body, maxBody))
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return errors.New("a JSON body is needed")
	}
	return json.Unmarshal(data, v)
}

// writeJSON encodes v with a status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// fail answers with an error message.
func fail(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// failErr answers with the status an error maps to.
func failErr(w http.ResponseWriter, err error) {
	fail(w, statusOf(err), err.Error())
}

// statusOf maps the errors the packages return to HTTP statuses.
func statusOf(err error) int {
	switch {
	case errors.Is(err, storage.ErrNotFound), errors.Is(err, runner.ErrNotFound), errors.Is(err, fs.ErrNotExist):
		return http.StatusNotFound
	case errors.Is(err, runner.ErrBusy):
		return http.StatusConflict
	case errors.Is(err, runner.ErrOutOfHours):
		return http.StatusForbidden
	case errors.Is(err, runner.ErrReasonNeeded), errors.Is(err, hours.ErrInvalid):
		return http.StatusBadRequest
	case errors.Is(err, runner.ErrNotInteractive), errors.Is(err, runner.ErrNotRunning), errors.Is(err, storage.ErrInvalidName):
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

// storageNotFound is the not found error, named for the handlers that
// find nothing themselves.
var storageNotFound = storage.ErrNotFound

// sorted returns the list in order.
func sorted(list []string) []string {
	sort.Strings(list)
	return list
}

// queryInt reads an integer query parameter, or the fallback.
func queryInt(r *http.Request, name string, fallback int) int {
	if v, err := strconv.Atoi(r.URL.Query().Get(name)); err == nil {
		return v
	}
	return fallback
}
