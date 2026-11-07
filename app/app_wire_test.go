package app_test

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/titpetric/task-ui/internal/assert"
)

// The compressor wire puts in front of every route, measured over the writer
// it passes a handler. Building a whole instance takes a database, a
// catalogue, a runner and a scheduler, and none of them decides this.
//
// A page is a quarter of a megabyte of HTML, so the transfer is most of what a
// reader waits for. A page comes back encoded, a stream the browser reads as it
// arrives does not, and a response already encoded is left alone.
func TestCompressesAPageAndLeavesAStreamAlone(t *testing.T) {
	t.Parallel()

	body := make([]byte, 64*1024)
	for i := range body {
		body[i] = byte('a' + i%26)
	}

	tests := []struct {
		title    string
		kind     string
		encoding string
		want     string
	}{
		{title: "a page", kind: "text/html; charset=utf-8", want: "gzip"},
		{title: "the API", kind: "application/json", want: "gzip"},
		{title: "an event stream", kind: "text/event-stream", want: ""},
		{title: "an artifact", kind: "application/octet-stream", want: ""},
		{title: "an answer that encoded itself", kind: "text/html", encoding: "br", want: "br"},
	}

	for _, test := range tests {
		handler := middleware.Compress(5)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", test.kind)
			if test.encoding != "" {
				w.Header().Set("Content-Encoding", test.encoding)
			}
			_, _ = w.Write(body)
		}))

		request := httptest.NewRequest(http.MethodGet, "/runs", nil)
		request.Header.Set("Accept-Encoding", "gzip")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		got := recorder.Result().Header.Get("Content-Encoding")
		assert.Equal(t, test.want, got, test.title)
		if test.want != "gzip" {
			assert.Equal(t, len(body), recorder.Body.Len(), test.title+": the body is untouched")
			continue
		}

		assert.Equal(t, "Accept-Encoding", recorder.Result().Header.Get("Vary"), test.title)
		assert.True(t, recorder.Body.Len() < len(body)/8, test.title+": 64KiB of one alphabet compresses")

		reader, err := gzip.NewReader(recorder.Body)
		assert.NoError(t, err, test.title)
		read, err := io.ReadAll(reader)
		assert.NoError(t, err, test.title)
		assert.Equal(t, string(body), string(read), test.title+": what comes back is what was written")
	}
}

// A client sending no Accept-Encoding gets the bytes the handler wrote, so a
// curl without --compressed still reads the page.
func TestCompressNeedsTheClientToAsk(t *testing.T) {
	t.Parallel()

	handler := middleware.Compress(5)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html></html>"))
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, "", recorder.Result().Header.Get("Content-Encoding"))
	assert.Equal(t, "<html></html>", recorder.Body.String())
}
