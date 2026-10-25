package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// hookInstance returns a webhook over a Forgejo integration with the
// secret s3, the workspace malta with its home from Malta/malta and the
// projects malta/apps and infra at the root, recording the directories
// pulled.
func hookInstance(root string, values map[string]string) (*api.Webhook, *[]string) {
	var pulled []string
	return &api.Webhook{
		Settings: func(context.Context) (map[string]string, error) { return values, nil },
		Workspaces: func(context.Context) ([]model.Workspace, error) {
			return []model.Workspace{{Folder: model.Default}, {Folder: "malta", HomeURL: "ssh://git@git.cdn.si:2222/Malta/malta.git"}}, nil
		},
		Projects: func(context.Context) ([]model.Project, error) {
			return []model.Project{
				{Folder: "malta/apps", SourceURL: "ssh://git@git.cdn.si:2222/Malta/apps.git"},
				{Folder: "infra", SourceURL: "https://git.cdn.si/ops/infra.git"},
			}, nil
		},
		Root: root,
		Pull: func(_ context.Context, dir string, w io.Writer) error {
			if strings.HasSuffix(dir, "infra") {
				return errors.New("index.lock held")
			}
			pulled = append(pulled, dir)
			io.WriteString(w, "Already up to date.\n")
			return nil
		},
	}, &pulled
}

// hook posts a body as Forgejo does, signed with secret under the event.
func hook(h http.Handler, event, secret, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/forgejo", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if event != "" {
		req.Header.Set("X-Forgejo-Event", event)
	}
	if secret != "" {
		req.Header.Set("X-Forgejo-Signature", forgejo.Sign(secret, []byte(body)))
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// A signed push pulls the checkouts of its repository, the project and
// the home alike; a ping answers; a bad signature is 401, an unknown
// repository 200 with the message, a pull that fails 502.
func TestWebhook(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	h, pulled := hookInstance(root, map[string]string{"url": "https://git.cdn.si", "token": "t", "secret": "s3"})

	rec := hook(h, "push", "s3", `{"ref":"refs/heads/main","repository":{"full_name":"Malta/apps","ssh_url":"ssh://git@git.cdn.si:2222/Malta/apps.git","clone_url":"https://git.cdn.si/Malta/apps.git"}}`)
	assert.Equal(t, http.StatusAccepted, rec.Code)
	var answer struct {
		Pulled []string `json:"pulled"`
		Output string   `json:"output"`
	}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &answer))
	assert.Equal(t, []string{"malta/apps"}, answer.Pulled)
	assert.Contains(t, answer.Output, "Already up to date.")
	assert.Equal(t, []string{filepath.Join(root, "malta", "src", "apps")}, *pulled)

	rec = hook(h, "push", "s3", `{"repository":{"full_name":"malta/malta","clone_url":"https://git.cdn.si/malta/malta.git"}}`)
	assert.Equal(t, http.StatusAccepted, rec.Code)
	assert.Contains(t, rec.Body.String(), `"pulled":["malta/home"]`)
	assert.Equal(t, filepath.Join(root, "malta", "home"), (*pulled)[1])

	rec = hook(h, "ping", "s3", `{"zen":"Forgejo"}`)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "pong")

	rec = hook(h, "push", "wrong", `{"repository":{"full_name":"malta/apps"}}`)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "signature")
	rec = hook(h, "push", "", `{"repository":{"full_name":"malta/apps"}}`)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	rec = hook(h, "push", "s3", `{"repository":{"full_name":"malta/nothing"}}`)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "no such checkout: malta/nothing")

	rec = hook(h, "push", "s3", `{"repository":{"full_name":"ops/infra"}}`)
	assert.Equal(t, http.StatusBadGateway, rec.Code)
	assert.Contains(t, rec.Body.String(), "infra: index.lock held")

	rec = hook(h, "create", "s3", `{"ref":"v1"}`)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "ignored: create")
	rec = hook(h, "push", "s3", `{"zen":"no repository"}`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	rec = hook(h, "push", "s3", `not json`)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Len(t, *pulled, 2)

	req := httptest.NewRequest(http.MethodGet, "/api/webhooks/forgejo", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

// With the integration off, or without a webhook secret, every call is
// 404, the signature unread.
func TestWebhook_Off(t *testing.T) {
	t.Parallel()
	h, pulled := hookInstance(t.TempDir(), map[string]string{})
	rec := hook(h, "push", "s3", `{"repository":{"full_name":"malta/apps"}}`)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "not set up")

	h, _ = hookInstance(t.TempDir(), map[string]string{"url": "https://git.cdn.si", "token": "t"})
	assert.Equal(t, http.StatusNotFound, hook(h, "ping", "", `{}`).Code)
	assert.Len(t, *pulled, 0)
}
