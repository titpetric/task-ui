//go:build integration

package app_test

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/titpetric/task-ui/internal/assert"
)

// A terminal window attaches over a websocket under /api, so the attach runs
// through the compressor wire installs on the root router. chi's writer passes
// a Hijack to the connection only while it has selected no encoder, and a
// handler that hijacks writes no header first.
//
// Over a real server, for the reason TestTelemetryMiddlewareHijack states: a
// recorder is no http.Hijacker, so a test through one proves nothing here.
func TestCompressLeavesTheAttachHijackable(t *testing.T) {
	t.Parallel()

	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Get("/api/runs/{id}/attach", func(w http.ResponseWriter, _ *http.Request) {
		hijacker, ok := w.(http.Hijacker)
		assert.True(t, ok, "the writer the compressor passes on is an http.Hijacker")
		if !ok {
			return
		}
		conn, buffered, err := hijacker.Hijack()
		assert.NoError(t, err)
		defer conn.Close()
		_, _ = buffered.WriteString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: terminal\r\n\r\nattached")
		assert.NoError(t, buffered.Flush())
	})

	server := httptest.NewServer(r)
	defer server.Close()

	conn, err := net.Dial("tcp", strings.TrimPrefix(server.URL, "http://"))
	assert.NoError(t, err)
	defer conn.Close()

	// The encoding a browser sends. chi wraps the writer only on this header;
	// without it the handler gets the connection's own.
	_, err = conn.Write([]byte("GET /api/runs/1/attach HTTP/1.1\r\nHost: test\r\nAccept-Encoding: gzip\r\n\r\n"))
	assert.NoError(t, err)
	answer, err := io.ReadAll(conn)
	assert.NoError(t, err)
	assert.Contains(t, string(answer), "101 Switching Protocols")
	assert.Contains(t, string(answer), "attached")
}
