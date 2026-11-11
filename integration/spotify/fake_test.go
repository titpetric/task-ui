package spotify_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/titpetric/task-ui/integration/spotify"
)

// A real ID is 22 characters of base62.
const (
	trackA = "4uLU6hMCjMI75M1A2tKUQC"
	trackB = "7ouMYWpwJ422jRcDASZB7P"
)

// memTokens keeps links in memory.
type memTokens struct {
	mu    sync.Mutex
	links map[string]spotify.Link
}

func (m *memTokens) Load(_ context.Context, user string) (*spotify.Link, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	link, ok := m.links[user]
	if !ok {
		return nil, nil
	}
	return &link, nil
}

func (m *memTokens) Save(_ context.Context, user string, link spotify.Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.links == nil {
		m.links = map[string]spotify.Link{}
	}
	m.links[user] = link
	return nil
}

func (m *memTokens) Delete(_ context.Context, user string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.links, user)
	return nil
}

// fake is Spotify: the account service and the Web API in one server.
type fake struct {
	*httptest.Server
	mu sync.Mutex
	// expect is the access token the API accepts.
	access  string
	refresh string
	// grant fails token requests with invalid_grant.
	grant bool
	// issued counts the access tokens handed out.
	issued int
	// verifier and code seen on the exchange.
	verifier string
	liked    map[string]bool
	items    map[string][]string
	playing  string
	// calls logs "METHOD path".
	calls []string
	// limited answers the next n requests with 429.
	limited int
	noDev   bool
	premium bool
}

func newFake(t *testing.T) *fake {
	t.Helper()
	f := &fake{liked: map[string]bool{}, items: map[string][]string{"pl1": {trackB}}, playing: trackA, access: "access-1-aaaaaaaa", refresh: "refresh-1-bbbbbbbb"}
	f.Server = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.Close)
	return f
}

func (f *fake) log() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string{}, f.calls...)
}

func jsonOut(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func apiError(w http.ResponseWriter, status int, reason, msg string) {
	w.WriteHeader(status)
	jsonOut(w, map[string]any{"error": map[string]any{"status": status, "message": msg, "reason": reason}})
}

func track(id, name string) map[string]any {
	return map[string]any{"id": id, "uri": "spotify:track:" + id, "name": name,
		"artists": []map[string]string{{"name": "Artist"}}, "album": map[string]string{"name": "Album"}}
}

func (f *fake) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.URL.Path == "/api/token" {
		_ = r.ParseForm()
		if f.grant {
			w.WriteHeader(http.StatusBadRequest)
			jsonOut(w, map[string]string{"error": "invalid_grant", "error_description": "Refresh token revoked " + r.PostForm.Get("refresh_token")})
			return
		}
		if r.PostForm.Get("grant_type") == "authorization_code" {
			f.verifier = r.PostForm.Get("code_verifier")
		}
		f.issued++
		f.access = fmt.Sprintf("access-%d-aaaaaaaa", f.issued+1)
		jsonOut(w, map[string]any{"access_token": f.access, "refresh_token": f.refresh, "expires_in": 3600, "scope": "user-read-playback-state user-modify-playback-state"})
		return
	}
	f.calls = append(f.calls, r.Method+" "+r.URL.Path)
	if f.limited > 0 {
		f.limited--
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}
	if r.Header.Get("Authorization") != "Bearer "+f.access {
		apiError(w, http.StatusUnauthorized, "", "The access token expired")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/v1")
	switch {
	case path == "/me":
		jsonOut(w, map[string]string{"id": "u1", "display_name": "Tit", "product": "premium"})
	case path == "/me/player" && r.Method == http.MethodGet:
		if f.playing == "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		jsonOut(w, map[string]any{"is_playing": true, "device": map[string]any{"name": "Kitchen", "type": "Speaker", "volume_percent": 30}, "item": track(f.playing, "Song")})
	case strings.HasPrefix(path, "/me/player/") && f.premium:
		apiError(w, http.StatusForbidden, "PREMIUM_REQUIRED", "Player command failed: Premium required")
	case strings.HasPrefix(path, "/me/player/") && f.noDev:
		apiError(w, http.StatusNotFound, "NO_ACTIVE_DEVICE", "Player command failed: No active device found")
	case strings.HasPrefix(path, "/me/player/"):
		w.WriteHeader(http.StatusNoContent)
	case path == "/search":
		jsonOut(w, map[string]any{"tracks": map[string]any{"items": []any{track(trackB, "Found")}}})
	case path == "/me/playlists":
		jsonOut(w, map[string]any{"items": []map[string]any{
			{"id": "pl1", "name": "Road trip", "owner": map[string]string{"id": "u1", "display_name": "Tit"}},
			{"id": "pl2", "name": "Road trip", "owner": map[string]string{"id": "u1", "display_name": "Tit"}},
			{"id": "pl3", "name": "Focus", "owner": map[string]string{"id": "u1", "display_name": "Tit"}},
			{"id": "pl4", "name": "Theirs", "owner": map[string]string{"id": "other", "display_name": "Other"}},
		}})
	case strings.HasPrefix(path, "/playlists/") && strings.HasSuffix(path, "/items"):
		id := strings.Split(path, "/")[2]
		switch r.Method {
		case http.MethodGet:
			var entries []map[string]any
			for _, t := range f.items[id] {
				entries = append(entries, map[string]any{"item": track(t, "T")})
			}
			jsonOut(w, map[string]any{"items": entries})
		case http.MethodPost:
			var body struct{ URIs []string }
			_ = json.NewDecoder(r.Body).Decode(&body)
			var raw map[string][]string
			_ = raw
			f.items[id] = append(f.items[id], strings.TrimPrefix(first(r, body.URIs), "spotify:track:"))
			jsonOut(w, map[string]string{"snapshot_id": "s"})
		case http.MethodDelete:
			f.items[id] = nil
			jsonOut(w, map[string]string{"snapshot_id": "s"})
		}
	case path == "/me/tracks/contains":
		ids := strings.Split(r.URL.Query().Get("ids"), ",")
		out := make([]bool, len(ids))
		for i, id := range ids {
			out[i] = f.liked[id]
		}
		jsonOut(w, out)
	case path == "/me/tracks":
		var body struct {
			IDs []string `json:"ids"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		for _, id := range body.IDs {
			f.liked[id] = r.Method == http.MethodPut
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.NotFound(w, r)
	}
}

// first returns the first URI the request body carried; the body is read
// once by the caller, so the handler passes what it decoded.
func first(_ *http.Request, uris []string) string {
	if len(uris) == 0 {
		return ""
	}
	return uris[0]
}

// service returns a Service over the fake, with the integration set up
// for the settings given on top of the defaults.
func (f *fake) service(settings map[string]string) (*spotify.Service, *memTokens) {
	values := map[string]string{"client_id": "cid", "client_secret": "client-secret-xyz", "redirect_uri": "https://task.example.com" + spotify.CallbackPath, "scopes": "write"}
	for k, v := range settings {
		values[k] = v
	}
	tokens := &memTokens{}
	return &spotify.Service{
		Settings: func(context.Context) (map[string]string, error) { return values, nil },
		Tokens:   tokens,
		Client:   &spotify.Client{API: f.URL + "/v1", Sleep: func(context.Context, time.Duration) error { return nil }},
		OAuth:    &spotify.OAuth{Accounts: f.URL},
		States:   &spotify.States{},
	}, tokens
}

var tit = spotify.Actor{Name: "tit", Role: "runner"}

// link walks the whole authorization code flow for tit.
func link(t *testing.T, svc *spotify.Service) {
	t.Helper()
	ctx := context.Background()
	authorize, err := svc.Begin(ctx, tit, "session-1")
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(authorize)
	if err := svc.Complete(ctx, tit, "session-1", u.Query().Get("state"), "the-code"); err != nil {
		t.Fatal(err)
	}
}
