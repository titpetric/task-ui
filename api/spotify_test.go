package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/spotify"
)

const playingTrack = "4uLU6hMCjMI75M1A2tKUQC"

type linkMemory struct {
	mu    sync.Mutex
	links map[string]spotify.Link
}

func (m *linkMemory) Load(_ context.Context, user string) (*spotify.Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if link, ok := m.links[user]; ok {
		return &link, nil
	}
	return nil, nil
}

func (m *linkMemory) Save(_ context.Context, user string, link spotify.Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.links[user] = link
	return nil
}

func (m *linkMemory) Delete(_ context.Context, user string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.links, user)
	return nil
}

// spotifyRoutes serves the API over a fake Spotify.
func spotifyRoutes(t *testing.T) (http.Handler, *linkMemory) {
	t.Helper()
	liked := false
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/token":
			_, _ = w.Write([]byte(`{"access_token":"access-token-123","refresh_token":"refresh-token-123","expires_in":3600,"scope":"x"}`))
		case "/v1/me":
			_, _ = w.Write([]byte(`{"id":"u1","display_name":"Tit"}`))
		case "/v1/me/player":
			_, _ = w.Write([]byte(`{"is_playing":true,"device":{"name":"Kitchen","type":"Speaker","volume_percent":20},"item":{"id":"` + playingTrack + `","uri":"spotify:track:` + playingTrack + `","name":"Song","artists":[{"name":"Artist"}],"album":{"name":"Album"}}}`))
		case "/v1/me/tracks/contains":
			_, _ = w.Write([]byte(map[bool]string{true: `[true]`, false: `[false]`}[liked]))
		case "/v1/me/tracks":
			liked = r.Method == http.MethodPut
		}
	}))
	t.Cleanup(fake.Close)
	memory := &linkMemory{links: map[string]spotify.Link{}}
	svc := &spotify.Service{
		Settings: func(context.Context) (map[string]string, error) {
			return map[string]string{"client_id": "cid", "client_secret": "secret-secret", "redirect_uri": "https://task.example.com" + spotify.CallbackPath, "scopes": "write"}, nil
		},
		Tokens: memory,
		Client: &spotify.Client{API: fake.URL + "/v1"},
		OAuth:  &spotify.OAuth{Accounts: fake.URL},
		States: &spotify.States{},
	}
	return api.New(api.Options{Spotify: svc, Now: time.Now}).Routes(), memory
}

func spotifyDo(h http.Handler, user *model.User, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req = req.WithContext(access.WithUser(req.Context(), user))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// The whole flow over the routes: authorize, finish, read, like twice, and
// disconnect. No answer holds a token.
func TestSpotify_Routes(t *testing.T) {
	t.Parallel()
	h, memory := spotifyRoutes(t)
	admin := &model.User{ID: "1", Name: "root", Admin: true, Role: model.RoleRunner}

	rec := spotifyDo(h, admin, http.MethodGet, "/integrations/spotify/authorize", "")
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	consent, err := url.Parse(rec.Header().Get("Location"))
	assert.NoError(t, err)
	state := consent.Query().Get("state")
	assert.NotEqual(t, "", state)

	rec = spotifyDo(h, admin, http.MethodGet, "/integrations/spotify/finish?code=c&state="+state, "")
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	// The same state again is refused.
	rec = spotifyDo(h, admin, http.MethodGet, "/integrations/spotify/finish?code=c&state="+state, "")
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	rec = spotifyDo(h, admin, http.MethodGet, "/spotify/now-playing", "")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"track":"Song"`)

	var out spotify.ChangeView
	rec = spotifyDo(h, admin, http.MethodPost, "/spotify/like", `{}`)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.True(t, out.Changed)
	rec = spotifyDo(h, admin, http.MethodPost, "/spotify/like", `{}`)
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))
	assert.False(t, out.Changed)

	rec = spotifyDo(h, admin, http.MethodGet, "/integrations/spotify/link", "")
	assert.Contains(t, rec.Body.String(), `"linked":true`)
	for _, secret := range []string{"access-token-123", "refresh-token-123", "secret-secret"} {
		assert.NotContains(t, rec.Body.String(), secret)
	}

	rec = spotifyDo(h, admin, http.MethodDelete, "/integrations/spotify/link", "")
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Len(t, memory.links, 0)
	rec = spotifyDo(h, admin, http.MethodGet, "/spotify/now-playing", "")
	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "link")
}

// Without the rules a viewer links and controls nothing, names the rule it
// lacks, and cannot disconnect someone else.
func TestSpotify_Permissions(t *testing.T) {
	t.Parallel()
	h, memory := spotifyRoutes(t)
	viewer := &model.User{ID: "2", Name: "view", Role: model.RoleViewer}
	memory.links["root"] = spotify.Link{Token: spotify.Token{Refresh: "r"}}

	for path, rule := range map[string]string{
		"/integrations/spotify/authorize": "spotify_link",
		"/spotify/now-playing":            "spotify_link",
		"/spotify/playlists":              "spotify_link",
	} {
		rec := spotifyDo(h, viewer, http.MethodGet, path, "")
		assert.Equal(t, http.StatusForbidden, rec.Code, path)
		assert.Contains(t, rec.Body.String(), rule)
	}
	for _, path := range []string{"/spotify/playback", "/spotify/like", "/spotify/playlist"} {
		rec := spotifyDo(h, viewer, http.MethodPost, path, `{}`)
		assert.Equal(t, http.StatusForbidden, rec.Code, path)
		assert.Contains(t, rec.Body.String(), "spotify_control")
	}
	rec := spotifyDo(h, viewer, http.MethodDelete, "/integrations/spotify/link?user=root", "")
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Len(t, memory.links, 1, "nothing was deleted")
}

// The callback holds no session and no state: it hands the browser to the
// finish route from a page of ours, with the code and state alone, escaped.
func TestSpotifyCallback(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	api.SpotifyCallback(rec, httptest.NewRequest(http.MethodGet, `/cb?code=a%22%3E%3Cscript%3E&state=s1&other=x`, nil))
	body := rec.Body.String()
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, body, `http-equiv="refresh"`)
	assert.Contains(t, body, spotify.FinishPath+"?code=")
	assert.Contains(t, body, "state=s1")
	assert.NotContains(t, body, "<script>")
	assert.NotContains(t, body, "other=")
	assert.Equal(t, "no-referrer", rec.Header().Get("Referrer-Policy"))
}
