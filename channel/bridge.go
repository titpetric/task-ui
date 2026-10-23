// Package channel is `task-ui channel`: a stdio MCP server that pushes
// the notes naming the token's user into an open Claude Code session,
// as a Claude Code channel. It declares the claude/channel capability,
// holds the instance's notes stream open with the token, and sends one
// notifications/claude/channel for every open instruction to the user
// from someone else. It has no tools: the session answers with the task
// MCP server's note_comment.
package channel

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/titpetric/task-ui/model"
)

// protocolVersions are the revisions of the protocol the bridge speaks,
// the newest first.
var protocolVersions = []string{"2025-06-18", "2025-03-26", "2024-11-05"}

// instructions tell the session what the channel's events are.
const instructions = "Notes from task-ui arrive as <channel source=\"...\" note_id=\"...\" author=\"...\" url=\"...\">: " +
	"each is an instruction to you, from the author, left on the task-ui notes board. " +
	"Do what it asks, and report on it with the task MCP server's note_comment and the note_id: " +
	"state working when you start, then done or failed with what you did. The channel has no tools."

// Bridge is the channel between an instance's notes and a session.
type Bridge struct {
	// URL is the instance's base link, such as https://task.example.com.
	URL string
	// Token is a token of the user whose notes the bridge pushes.
	Token string
	// In and Out are the session's side of stdio.
	In  io.Reader
	Out io.Writer
	// Log takes what the bridge has to say about itself, never Out.
	Log *log.Logger
	// Client calls the instance; http.DefaultClient unless set.
	Client *http.Client
	// Backoff is the first wait before connecting again after the stream
	// drops, doubling up to MaxBackoff; a second and 30 seconds unless set.
	Backoff    time.Duration
	MaxBackoff time.Duration

	mu   sync.Mutex
	me   string
	sent map[string]bool
}

// request is a JSON-RPC message from the session.
type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// rpcError is a JSON-RPC error.
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Run answers the session on In until it closes, and once the session
// sends notifications/initialized pushes the user's open notes, then each one as it comes,
// until the context ends.
func (b *Bridge) Run(ctx context.Context) error {
	if b.URL == "" || b.Token == "" {
		return errors.New("usage: task-ui channel --url <instance> --token-file <file>")
	}
	b.URL = strings.TrimRight(b.URL, "/")
	if b.Log == nil {
		b.Log = log.New(io.Discard, "", 0)
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	scanner := bufio.NewScanner(b.In)
	scanner.Buffer(make([]byte, 64<<10), 4<<20)
	watching := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var req request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			b.write(map[string]any{"jsonrpc": "2.0", "id": nil, "error": rpcError{Code: -32700, Message: err.Error()}})
			continue
		}
		if len(req.ID) == 0 {
			// A notification: notifications/initialized starts the
			// notes.
			if req.Method == "notifications/initialized" && !watching {
				watching = true
				go b.watch(ctx)
			}
			continue
		}
		result, rerr := b.handle(req)
		reply := map[string]any{"jsonrpc": "2.0", "id": req.ID}
		if rerr != nil {
			reply["error"] = rerr
		} else {
			reply["result"] = result
		}
		b.write(reply)
	}
	return scanner.Err()
}

// handle answers a request by its method.
func (b *Bridge) handle(req request) (any, *rpcError) {
	switch req.Method {
	case "initialize":
		var params struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		_ = json.Unmarshal(req.Params, &params)
		version := protocolVersions[0]
		for _, known := range protocolVersions {
			if known == params.ProtocolVersion {
				version = known
			}
		}
		return map[string]any{
			"protocolVersion": version,
			"capabilities":    map[string]any{"experimental": map[string]any{"claude/channel": map[string]any{}}},
			"serverInfo":      map[string]string{"name": "task-ui-channel", "version": "1"},
			"instructions":    instructions,
		}, nil
	case "ping":
		return map[string]any{}, nil
	case "tools/list":
		return map[string]any{"tools": []any{}}, nil
	}
	return nil, &rpcError{Code: -32601, Message: "no method " + req.Method}
}

// write sends one message to the session, a line of JSON.
func (b *Bridge) write(message any) {
	data, err := json.Marshal(message)
	if err != nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	_, _ = b.Out.Write(append(data, '\n'))
}

// watch finds the token's user, then holds the notes stream open,
// reading the open notes each time it connects, and connects again with
// a growing wait when the stream drops.
func (b *Bridge) watch(ctx context.Context) {
	wait := b.backoff()
	for ctx.Err() == nil {
		err := b.connect(ctx)
		if ctx.Err() != nil {
			return
		}
		if errors.Is(err, errConnected) {
			wait = b.backoff()
		}
		b.Log.Printf("notes stream: %v; connecting again in %s", err, wait)
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		wait = min(2*wait, b.maxBackoff())
	}
}

// errConnected says the stream opened before it dropped, so the wait
// starts over.
var errConnected = errors.New("the stream ended")

// connect reads who the user is once, opens the stream, sends the open
// notes, and pushes the stream's notes until it drops.
func (b *Bridge) connect(ctx context.Context) error {
	if b.me == "" {
		var me struct {
			Name string `json:"name"`
		}
		if err := b.get(ctx, "/api/me", &me); err != nil {
			return err
		}
		if me.Name == "" {
			return errors.New("the token names no user")
		}
		b.me = me.Name
		b.Log.Printf("pushing the notes that name %s", b.me)
	}
	stream, err := b.open(ctx, "/api/sse/notes")
	if err != nil {
		return err
	}
	defer stream.Close()
	if err := b.backlog(ctx); err != nil {
		return err
	}
	if err := readEvents(stream, func(name string, data []byte) {
		if name != "note" {
			return
		}
		var note model.Note
		if json.Unmarshal(data, &note) == nil {
			b.consider(note)
		}
	}); err != nil {
		return fmt.Errorf("%w: %v", errConnected, err)
	}
	return errConnected
}

// backlog sends the open notes that name the user, oldest first.
func (b *Bridge) backlog(ctx context.Context) error {
	q := url.Values{"mention": {b.me}, "state": {model.NoteOpen}, "limit": {"50"}}
	var out struct {
		Notes []model.Note `json:"notes"`
	}
	if err := b.get(ctx, "/api/notes?"+q.Encode(), &out); err != nil {
		return err
	}
	for i := len(out.Notes) - 1; i >= 0; i-- {
		b.consider(out.Notes[i])
	}
	return nil
}

// consider pushes an open instruction to the user from someone else, once while it stays open. Notes to anyone else never
// reach the session: the channel is no way to put text in front of it.
func (b *Bridge) consider(note model.Note) {
	b.mu.Lock()
	if b.sent == nil {
		b.sent = map[string]bool{}
	}
	open := note.ParentID == "" && note.Channel == model.ChannelNotes && !note.IsPrompt() && note.State == model.NoteOpen
	if !open {
		// A reply may open the instruction again, which pushes it again.
		delete(b.sent, note.ID)
		b.mu.Unlock()
		return
	}
	if !note.Mentioned(b.me) || note.UserName == b.me || b.sent[note.ID] {
		b.mu.Unlock()
		return
	}
	b.sent[note.ID] = true
	b.mu.Unlock()
	b.Log.Printf("pushing note %s from %s", note.ID, note.UserName)
	b.write(map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/claude/channel",
		"params": map[string]any{
			"content": fmt.Sprintf("Note %s from %s: %s\n\nReply with note_comment (state working, then done or failed).", note.ID, note.UserName, note.Body),
			"meta":    map[string]string{"note_id": note.ID, "author": note.UserName, "url": b.URL + "/notes#note-" + note.ID},
		},
	})
}

// get decodes the JSON answer to a GET of a path of the instance.
func (b *Bridge) get(ctx context.Context, path string, out any) error {
	body, err := b.open(ctx, path)
	if err != nil {
		return err
	}
	defer body.Close()
	return json.NewDecoder(body).Decode(out)
}

// open sends a GET with the token and returns the body of a success.
func (b *Bridge) open(ctx context.Context, path string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.URL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+b.Token)
	client := b.Client
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close()
		return nil, fmt.Errorf("GET %s: %s", path, res.Status)
	}
	return res.Body, nil
}

func (b *Bridge) backoff() time.Duration {
	if b.Backoff > 0 {
		return b.Backoff
	}
	return time.Second
}

func (b *Bridge) maxBackoff() time.Duration {
	if b.MaxBackoff > 0 {
		return b.MaxBackoff
	}
	return 30 * time.Second
}

// readEvents reads server-sent events until the stream ends, calling
// on with each event's name and data.
func readEvents(r io.Reader, on func(name string, data []byte)) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64<<10), 4<<20)
	name, data := "", []string{}
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "":
			if len(data) > 0 {
				on(name, []byte(strings.Join(data, "\n")))
			}
			name, data = "", data[:0]
		case strings.HasPrefix(line, ":"):
		case strings.HasPrefix(line, "event:"):
			name = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		case strings.HasPrefix(line, "data:"):
			data = append(data, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " "))
		}
	}
	return scanner.Err()
}
