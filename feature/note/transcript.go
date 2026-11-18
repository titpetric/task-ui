package note

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// transcriptEvery is how often a running transcript is read.
const transcriptEvery = 100 * time.Millisecond

// transcriptMaxLine is the longest record a tail holds before dropping it.
const transcriptMaxLine = 1 << 20

// ThreadTopic returns the bus topic of one thread.
func ThreadTopic(id string) string { return "thread:" + id }

// transcriptRecord is one line of the JSONL a Claude run writes.
type transcriptRecord struct {
	Seq    int64           `json:"seq"`
	Kind   string          `json:"kind"`
	Body   string          `json:"body"`
	ToolID string          `json:"tool_id"`
	Data   json.RawMessage `json:"data"`
}

// eventSink keeps one transcript record and reports the last it kept.
type eventSink interface {
	Append(ctx context.Context, event *model.NoteEvent) error
	Last(ctx context.Context, runID string) (*model.NoteEvent, error)
}

// transcriptTail reads a run's transcript while it runs and keeps each record.
type transcriptTail struct {
	path       string
	noteID     string
	runID      string
	questionID string
	events     eventSink
	bus        *event.Bus
}

// follow reads the transcript at each tick until done, then drains it.
func (t *transcriptTail) follow(tick <-chan time.Time, done <-chan struct{}) {
	file, err := os.Open(t.path)
	if err != nil {
		return
	}
	defer file.Close()
	reader := bufio.NewReaderSize(file, 64<<10)
	var pending []byte
	var skipped bool
	for {
		select {
		case <-tick:
			pending, skipped = t.drain(reader, pending, skipped)
		case <-done:
			t.drain(reader, pending, skipped)
			return
		}
	}
}

// drain writes every complete line the reader holds, keeping a partial one.
func (t *transcriptTail) drain(reader *bufio.Reader, pending []byte, skipped bool) ([]byte, bool) {
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			switch {
			case skipped:
				skipped = line[len(line)-1] != '\n'
			case line[len(line)-1] != '\n':
				pending = append(pending, line...)
				if len(pending) > transcriptMaxLine {
					t.keep("error", "A transcript line was too long and was dropped.", "", 0, nil)
					pending, skipped = nil, true
				}
			default:
				full := line
				if len(pending) > 0 {
					full = append(pending, line...)
					pending = nil
				}
				t.record(bytes.TrimSpace(full))
			}
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				slog.Warn("transcript read failed", "run", t.runID, "err", err)
			}
			return pending, skipped
		}
	}
}

// record keeps one complete line, if it is the agent's JSON.
func (t *transcriptTail) record(line []byte) {
	if len(line) == 0 {
		return
	}
	var rec transcriptRecord
	if err := json.Unmarshal(line, &rec); err != nil {
		return
	}
	if rec.Seq <= 0 || rec.Kind == "" {
		return
	}
	t.keep(rec.Kind, rec.Body, rec.ToolID, rec.Seq, rec.Data)
}

// keep appends one event under the run and publishes it on the thread.
func (t *transcriptTail) keep(kind, body, toolID string, seq int64, data json.RawMessage) {
	if len(data) == 0 {
		data = json.RawMessage("{}")
	}
	event := &model.NoteEvent{
		NoteID: t.noteID, RunID: t.runID, QuestionID: t.questionID,
		Seq: seq, Kind: kind, ToolID: toolID, Body: body, Data: string(data),
	}
	if event.Seq <= 0 {
		event.Seq = t.next()
	}
	if err := t.events.Append(context.Background(), event); err != nil {
		slog.Warn("transcript event not saved", "run", t.runID, "err", err)
		return
	}
	if t.bus != nil {
		t.bus.Publish(ThreadTopic(t.noteID), "event", ThreadEvent{
			Note: t.noteID, Run: t.runID, Question: t.questionID,
			Seq: event.Seq, Kind: kind, ToolID: toolID, Body: body, Data: data,
		})
	}
}

// next returns the next sequence a synthetic record takes.
func (t *transcriptTail) next() int64 {
	last, err := t.events.Last(context.Background(), t.runID)
	if err != nil {
		return 1
	}
	return last.Seq + 1
}

// watch makes the file a run writes its transcript to and the tail that keeps it.
func (b *Board) watch(note, question *model.Note, req *runner.Request, again *model.Note) (*transcriptTail, chan struct{}) {
	if b.NoteEvents == nil || b.Events == nil || req.Transcript != "" {
		return nil, nil
	}
	file, err := os.CreateTemp("", "note-events-*.jsonl")
	if err != nil {
		slog.Warn("transcript file not made", "note", note.ID, "err", err)
		return nil, nil
	}
	_ = file.Close()
	req.Transcript = file.Name()
	if again != nil && again.RunID != "" {
		_ = b.NoteEvents.DeleteByRun(context.Background(), again.RunID)
	}
	return &transcriptTail{
		path: file.Name(), noteID: note.ID, questionID: question.ID,
		events: b.NoteEvents, bus: b.Events,
	}, make(chan struct{})
}

// finalEvent closes a run whose transcript ended without a result or an error.
func (b *Board) finalEvent(ctx context.Context, note, question *model.Note, runID string) {
	if b.NoteEvents == nil || runID == "" {
		return
	}
	last, err := b.NoteEvents.Last(ctx, runID)
	if err != nil || last == nil || last.Kind == "result" || last.Kind == "error" {
		return
	}
	data := json.RawMessage("{}")
	event := &model.NoteEvent{
		NoteID: note.ID, RunID: runID, QuestionID: question.ID,
		Seq: last.Seq + 1, Kind: "error", Body: "The run ended without a final message.", Data: string(data),
	}
	if err := b.NoteEvents.Append(ctx, event); err != nil {
		slog.Warn("transcript end not saved", "run", runID, "err", err)
		return
	}
	if b.Events != nil {
		b.Events.Publish(ThreadTopic(note.ID), "event", ThreadEvent{
			Note: note.ID, Run: runID, Question: question.ID,
			Seq: event.Seq, Kind: event.Kind, Body: event.Body, Data: data,
		})
	}
}

// transcriptSink narrows a NoteEventStorage to the tail's append.
var _ eventSink = (*storage.NoteEventStorage)(nil)
