package channel_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/titpetric/task-ui/channel"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// note is a note's JSON as the instance sends it.
func note(id, author, body, state, mentions string) string {
	data, _ := json.Marshal(model.Note{ID: id, Channel: model.ChannelNotes, UserName: author, Body: body, State: state, Mentions: model.EncodeMentions([]string{mentions})})
	return string(data)
}

// instance is a fake task-ui: the token's user is consuela, one open
// note waits for her, and the notes stream drops after its first
// events, then holds the second connection open with one more.
func instance(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	var streams atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("/api/me", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tui_test" {
			http.Error(w, "no", http.StatusUnauthorized)
			return
		}
		_, _ = io.WriteString(w, `{"name":"consuela"}`)
	})
	mux.HandleFunc("/api/notes", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "consuela", r.URL.Query().Get("mention"))
		_, _ = io.WriteString(w, `{"notes":[`+note("n1", "black", "@consuela tidy the images", "open", "consuela")+`]}`)
	})
	mux.HandleFunc("/api/sse/notes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		send := func(data string) {
			fmt.Fprintf(w, "event: note\ndata: %s\n\n", data)
			w.(http.Flusher).Flush()
		}
		_, _ = io.WriteString(w, ": connected\n\n")
		if streams.Add(1) == 1 {
			send(note("n2", "black", "@ops not yours", "open", "ops"))
			send(note("n3", "consuela", "@consuela a note to herself", "open", "consuela"))
			send(note("n1", "black", "@consuela tidy the images", "open", "consuela"))
			send(note("n4", "black", "@consuela deploy the blog", "open", "consuela"))
			return
		}
		send(note("n5", "black", "@consuela and the docs", "open", "consuela"))
		<-r.Context().Done()
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server, &streams
}

// An open note to the token's user from someone else reaches the
// session once, with its ID, author and link; notes to others and the
// user's own do not; the bridge reads the open notes when it starts and
// connects again after the stream drops.
func TestBridge_Run(t *testing.T) {
	t.Parallel()
	server, streams := instance(t)
	in, toBridge := io.Pipe()
	fromBridge, out := io.Pipe()
	b := &channel.Bridge{URL: server.URL + "/", Token: "tui_test", In: in, Out: out, Backoff: 10 * time.Millisecond}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = b.Run(ctx) }()

	lines := make(chan map[string]any, 16)
	go func() {
		scanner := bufio.NewScanner(fromBridge)
		for scanner.Scan() {
			var message map[string]any
			if json.Unmarshal(scanner.Bytes(), &message) == nil {
				lines <- message
			}
		}
	}()
	next := func() map[string]any {
		t.Helper()
		select {
		case message := <-lines:
			return message
		case <-time.After(5 * time.Second):
			t.Fatal("nothing from the bridge")
			return nil
		}
	}

	_, _ = io.WriteString(toBridge, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}`+"\n")
	init := next()["result"].(map[string]any)
	assert.Equal(t, "2025-03-26", init["protocolVersion"])
	assert.Equal(t, map[string]any{"experimental": map[string]any{"claude/channel": map[string]any{}}}, init["capabilities"])
	assert.Contains(t, init["instructions"].(string), "note_comment")
	_, _ = io.WriteString(toBridge, `{"jsonrpc":"2.0","method":"notifications/initialized"}`+"\n")

	ids := []string{}
	for range 3 {
		message := next()
		assert.Equal(t, "notifications/claude/channel", message["method"])
		params := message["params"].(map[string]any)
		meta := params["meta"].(map[string]any)
		ids = append(ids, meta["note_id"].(string))
		assert.Equal(t, "black", meta["author"])
		assert.Equal(t, server.URL+"/notes#note-"+meta["note_id"].(string), meta["url"])
		if meta["note_id"] == "n1" {
			assert.Equal(t, "Note n1 from black: @consuela tidy the images\n\nReply with note_comment (state working, then done or failed).", params["content"])
		}
	}
	assert.Equal(t, []string{"n1", "n4", "n5"}, ids)
	assert.True(t, streams.Load() >= 2)

	_, _ = io.WriteString(toBridge, `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"reply"}}`+"\n")
	assert.Equal(t, float64(-32601), next()["error"].(map[string]any)["code"])
	select {
	case extra := <-lines:
		t.Fatalf("one push too many: %v", extra)
	case <-time.After(100 * time.Millisecond):
	}
}

// The bridge needs an instance and a token.
func TestBridge_Usage(t *testing.T) {
	t.Parallel()
	err := (&channel.Bridge{}).Run(context.Background())
	assert.Error(t, err)
}
