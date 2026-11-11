package api

import (
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/runner"
)

// errRunSecretMethod refuses a method other than GET.
var errRunSecretMethod = errors.New("GET a secret")

// RunSecret answers GET /api/run/secrets/{name} for task-ui get-secret
// in a run: the run's own token, TASK_UI_RUN_TOKEN, signs it in, and it
// returns a secret that reaches the run's folder, the same set the
// run's environment gets, recorded read against the run. The token
// signs in nothing else, and dies with its run. A token of no running
// run answers 401, a name the folder's secrets lack 404.
type RunSecret struct {
	Source RunSecretSource
}

// ServeHTTP answers {"name","value"}, else the error: 401 no run, 404 no secret.
func (h *RunSecret) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.serveHTTP(w, r); err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, errRunSecretMethod):
			status = http.StatusMethodNotAllowed
		case errors.Is(err, runner.ErrRunToken):
			status = http.StatusUnauthorized
		case errors.Is(err, runner.ErrNoSecret):
			status = http.StatusNotFound
		}
		fail(w, status, err.Error())
	}
}

// serveHTTP writes the secret, or returns why it does not.
func (h *RunSecret) serveHTTP(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return errRunSecretMethod
	}
	token := access.BearerToken(r)
	if token == "" || h.Source == nil {
		return fmt.Errorf("%w: TASK_UI_RUN_TOKEN, the run's own token, signs this in", runner.ErrRunToken)
	}
	name := path.Base(r.URL.Path)
	value, err := h.Source.Secret(r.Context(), token, name)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	writeJSON(w, http.StatusOK, map[string]string{"name": name, "value": value})
	return nil
}
