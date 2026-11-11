package mcp_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/integration/mcp"
	"github.com/titpetric/task-ui/internal/assert"
)

// callTool posts a tools/call and returns the body.
func callTool(api http.Handler, name string, args map[string]any) string {
	data, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": map[string]any{"name": name, "arguments": args}})
	rec := httptest.NewRecorder()
	mcp.New(api, "").ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(data)))
	return rec.Body.String()
}

// The Spotify tools call the API's routes with the arguments given, and
// say what it answered; a refusal names the permission and points to whoami.
func TestSpotifyTools(t *testing.T) {
	t.Parallel()
	var got []string
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := new(bytes.Buffer)
		_, _ = body.ReadFrom(r.Body)
		got = append(got, r.Method+" "+r.URL.RequestURI()+" "+body.String())
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/spotify/now-playing":
			_, _ = w.Write([]byte(`{"active":true,"playing":true,"track":"Song","artists":"Artist","album":"Album","device":"Kitchen","device_type":"Speaker","volume":30}`))
		case "/spotify/like":
			_, _ = w.Write([]byte(`{"changed":false,"message":"Song is already in your Liked Songs; nothing changed."}`))
		default:
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"you need the spotify_control permission"}`))
		}
	})

	out := callTool(api, "spotify_now_playing", nil)
	assert.Contains(t, out, "Playing: Song - Artist (Album) on Kitchen (Speaker), volume 30%")
	out = callTool(api, "spotify_like", map[string]any{})
	assert.Contains(t, out, "already in your Liked Songs")
	out = callTool(api, "spotify_playlist_remove", map[string]any{"playlist": "Focus"})
	assert.Contains(t, out, "you need the spotify_control permission")
	assert.Contains(t, out, "call whoami")
	assert.Equal(t, `POST /spotify/playlist {"playlist":"Focus","remove":true}`+"\n", got[2])
}
