package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/titpetric/task-ui/internal/assert"
)

// A whole body is compressed and a range of one is not, because a range of
// a gzip stream arrives under the plain body's Content-Range.
func TestUncompressedRanges(t *testing.T) {
	t.Parallel()

	body := strings.Repeat("a log line that compresses well\n", 600)
	handler := uncompressedRanges(middleware.Compress(5))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.ServeContent(w, r, "build.log", time.Time{}, strings.NewReader(body))
	}))

	ask := func(ranged bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/artifacts/files/build.log", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		if ranged {
			req.Header.Set("Range", "bytes=0-99")
		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}

	whole := ask(false)
	assert.Equal(t, http.StatusOK, whole.Code)
	assert.Equal(t, "gzip", whole.Header().Get("Content-Encoding"))
	assert.True(t, whole.Body.Len() < len(body))

	part := ask(true)
	assert.Equal(t, http.StatusPartialContent, part.Code)
	assert.Equal(t, "", part.Header().Get("Content-Encoding"))
	assert.Equal(t, 100, part.Body.Len())
}
