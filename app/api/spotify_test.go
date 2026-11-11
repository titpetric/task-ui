package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/app/api"
	"github.com/titpetric/task-ui/integration/spotify"
	"github.com/titpetric/task-ui/internal/assert"
)

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
