//go:build integration

package app_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
)

// A terminal window attaches over a websocket, which hijacks the connection,
// and the middleware is in front of every route of the instance: a tracer that
// hands the handler a writer it cannot hijack panics the attach instead of
// serving it, which is what took the shell off task.cdn.si.
//
// Over a real server, because that is the only writer that is an
// http.Hijacker: httptest.NewRecorder is not one, so a test through the
// recorder proves nothing here. A listening socket is also why this sits in
// the integration suite rather than beside the unit tests.
func TestTelemetryMiddlewareHijack(t *testing.T) {
	t.Parallel()
	telemetry, err := app.NewTelemetry(context.Background(), nil, "task-ui")
	assert.NoError(t, err)

	opts := telemetry.Defaults()
	opts.Enabled = true
	assert.NoError(t, telemetry.Apply(opts), "a tracer that records is the one in front of the attach")

	server := httptest.NewServer(telemetry.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hijacker, ok := w.(http.Hijacker)
		assert.True(t, ok, "the writer the middleware passes on is an http.Hijacker")
		if !ok {
			return
		}
		conn, buffered, err := hijacker.Hijack()
		assert.NoError(t, err)
		defer conn.Close()
		_, _ = buffered.WriteString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: terminal\r\n\r\nattached")
		assert.NoError(t, buffered.Flush())
	})))
	defer server.Close()

	conn, err := net.Dial("tcp", strings.TrimPrefix(server.URL, "http://"))
	assert.NoError(t, err)
	defer conn.Close()

	_, err = conn.Write([]byte("GET /api/runs/1/attach HTTP/1.1\r\nHost: test\r\n\r\n"))
	assert.NoError(t, err)
	answer, err := io.ReadAll(conn)
	assert.NoError(t, err)
	assert.Contains(t, string(answer), "101 Switching Protocols")
	assert.Contains(t, string(answer), "attached")
}
