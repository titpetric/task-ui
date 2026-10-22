package access_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestSameOrigin(t *testing.T) {
	t.Parallel()
	handler := access.SameOrigin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	post := func(origin, referer string) int {
		r := httptest.NewRequest(http.MethodPost, "http://app.example/x", nil)
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		if referer != "" {
			r.Header.Set("Referer", referer)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w.Code
	}
	assert.Equal(t, http.StatusNoContent, post("http://app.example", ""))
	assert.Equal(t, http.StatusNoContent, post("", "http://app.example/page"))
	assert.Equal(t, http.StatusNoContent, post("", ""), "no origin at all is a non-browser client")
	assert.Equal(t, http.StatusForbidden, post("http://evil.example", ""))
	assert.Equal(t, http.StatusForbidden, post("::bad::", ""))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "http://app.example/x", nil))
	assert.Equal(t, http.StatusNoContent, w.Code, "reads are never refused")
}

func TestSameHost(t *testing.T) {
	t.Parallel()
	r := httptest.NewRequest(http.MethodPost, "http://app.example/x", nil)
	r.Header.Set("X-Forwarded-Host", "public.example, other.example")
	assert.True(t, access.SameHost(r, "app.example"))
	assert.True(t, access.SameHost(r, "public.example"))
	assert.True(t, access.SameHost(r, "other.example"))
	assert.False(t, access.SameHost(r, "evil.example"))
	assert.False(t, access.SameHost(r, ""))
}

func TestSafeBack(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{
		"":              "/",
		"/f/echo":       "/f/echo",
		"//evil":        "/",
		"/\\evil":       "/",
		"http://x":      "/",
		"/settings?x=1": "/settings?x=1",
	} {
		assert.Equal(t, want, access.SafeBack(in), in)
	}
}
