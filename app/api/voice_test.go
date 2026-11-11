package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/app/api"
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/orb/stt"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/internal/assert"
)

// voiceRoutes serves the API over a transcriber that writes one line.
func voiceRoutes(t *testing.T, said string) http.Handler {
	t.Helper()
	dir := t.TempDir()
	bin, model := filepath.Join(dir, "whisper-cli"), filepath.Join(dir, "m.bin")
	for _, path := range []string{bin, model} {
		assert.NoError(t, os.WriteFile(path, nil, 0o600))
	}
	svc := &stt.Service{
		Settings: func(context.Context) (map[string]string, error) {
			return map[string]string{"bin": bin, "model": model}, nil
		},
		Run: func(_ context.Context, _ string, args, _ []string) error {
			return os.WriteFile(args[5]+".txt", []byte(said), 0o600)
		},
	}
	return api.New(api.Options{STT: svc}).Routes()
}

func voiceDo(h http.Handler, user *model.User, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req = req.WithContext(access.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// A clip posted while dictating comes back as words, and the server says
// it transcribes before the orb posts anything.
func TestVoice_Transcribe(t *testing.T) {
	t.Parallel()
	h := voiceRoutes(t, "[BLANK_AUDIO]\n Deploy the api.\n")
	admin := &model.User{ID: "1", Name: "root", Admin: true, Role: model.RoleRunner}

	rec := voiceDo(h, admin, http.MethodGet, "/voice/ready", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"ready":true`)

	rec = voiceDo(h, admin, http.MethodPost, "/voice/transcribe", "RIFFclip")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"text":"Deploy the api."`)

	// A clip with nothing in it is the caller's mistake.
	rec = voiceDo(h, admin, http.MethodPost, "/voice/transcribe", "")
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// A server with no transcriber answers 503, and the orb stops posting on
// it. A reader who may not ask Claude transcribes nothing.
func TestVoice_Off(t *testing.T) {
	t.Parallel()
	off := api.New(api.Options{}).Routes()
	admin := &model.User{ID: "1", Name: "root", Admin: true, Role: model.RoleRunner}
	rec := voiceDo(off, admin, http.MethodPost, "/voice/transcribe", "RIFFclip")
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), "not set up")
	rec = voiceDo(off, admin, http.MethodGet, "/voice/ready", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"ready":false`)

	viewer := &model.User{ID: "2", Name: "view", Role: model.RoleViewer}
	for _, call := range [][2]string{{http.MethodPost, "/voice/transcribe"}, {http.MethodGet, "/voice/ready"}} {
		rec = voiceDo(voiceRoutes(t, "x"), viewer, call[0], call[1], "RIFFclip")
		assert.Equal(t, http.StatusForbidden, rec.Code, call[1])
		assert.Contains(t, rec.Body.String(), "prompt/edit", call[1])
	}
}
