package app

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/orb/recording"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/phpfront"
	"github.com/titpetric/task-ui/runner"
)

// frontFrameRows is the least rows a frame on a page shows, the number
// frontend/page.go trims to.
const frontFrameRows = 8

// Running is the runs the runner holds, by the key their job is listed under.
func (h *FrontHost) Running(context.Context) []phpfront.Running {
	if h.Runner == nil {
		return nil
	}
	var out []phpfront.Running
	for _, session := range h.Runner.List() {
		run := session.Snapshot()
		if run.Finished() {
			continue
		}
		out = append(out, phpfront.Running{Key: run.Key(), ID: run.ID, State: string(run.State)})
	}
	return out
}

// Flows is the workflow run in progress of every workflow that has one, as
// the workflow runner holds them.
func (h *FrontHost) Flows(context.Context) []phpfront.Flow {
	if h.Workflows == nil {
		return nil
	}
	held := h.Workflows.Running()
	out := make([]phpfront.Flow, 0, len(held))
	for id, run := range held {
		out = append(out, phpfront.Flow{WorkflowID: id, RunID: run})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].WorkflowID < out[j].WorkflowID })
	return out
}

// Publish sends an event on the stream every live label reads.
func (h *FrontHost) Publish(_ context.Context, topic string, payload map[string]any) error {
	if h.Events == nil {
		return nil
	}
	name, _ := payload["event"].(string)
	h.Events.Publish(topic, name, payload)
	return nil
}

// Dispatch starts a run over a comment through the dispatcher the plugins use, with the action file resolved against the front end's own tree.
// An instance with no board returns an error for the page's form to draw.
func (h *FrontHost) Dispatch(ctx context.Context, comment, action, folder string) (string, error) {
	if h.Dispatcher == nil {
		return "", errors.New("this instance starts no runs")
	}
	return h.Dispatcher.Dispatch(ctx, h.FrontRoot, comment, action, folder)
}

// RunOf is the run a note holds, "" before the run starts or on an instance
// that keeps no notes.
func (h *FrontHost) RunOf(ctx context.Context, note string) string {
	if h.Runs == nil {
		return ""
	}
	return h.Runs.RunOf(ctx, note)
}

// Session is a run as the runner holds it. A waiting run reports its queue
// position in place of a message.
func (h *FrontHost) Session(_ context.Context, id string) phpfront.Session {
	found, ok := h.held(id)
	if !ok {
		return phpfront.Session{}
	}
	run := found.Snapshot()
	out := phpfront.Session{
		Live:     true,
		Attached: found.Attached(),
		Message:  found.Status().Message,
		State:    run.State,
		ExitCode: run.ExitCode,
		Title:    run.Title,
	}
	if run.FinishedAt != nil {
		out.FinishedAt = run.FinishedAt.String()
	}
	if at := h.Runner.Position(id); at > 0 {
		out.Message = fmt.Sprintf("waits for a free slot, number %d in line", at)
	}
	return out
}

// Frame is the last screen of a live run, as the SVG a page embeds.
func (h *FrontHost) Frame(_ context.Context, id string) string {
	found, ok := h.held(id)
	if !ok {
		return ""
	}
	return found.Frame().Trim(frontFrameRows).SVG(runner.DarkPalette)
}

// Screen is the last frame of a run that has ended, replayed off the recording
// the run wrote, the frame a pipeline draws under every step. The caller holds
// the row, so castPath and cols arrive with the ID.
//
// A run still going answers from the live session, as Frame does, and so does
// a recording that will not replay: a frame is decoration and a page that
// cannot draw one still draws the step.
func (h *FrontHost) Screen(_ context.Context, id, castPath string, cols int) phpfront.Screen {
	found, live := h.held(id)
	fromSession := func() phpfront.Screen {
		if !live {
			return phpfront.Screen{}
		}
		drawn := found.Frame().Trim(frontFrameRows).SVG(runner.DarkPalette)
		return phpfront.Screen{SVG: drawn}
	}
	if live && !found.Status().Finished() || h.Recordings == nil || castPath == "" {
		return fromSession()
	}
	cast, err := h.Recordings.Read(&model.Run{CastPath: castPath})
	if errors.Is(err, recording.ErrExpired) {
		gone := runner.TextFrame(cols, frontFrameRows, recording.Expired).SVG(runner.DarkPalette)
		return phpfront.Screen{SVG: gone, Complete: true}
	}
	if err != nil {
		return fromSession()
	}
	defer cast.Close()
	replay, err := runner.ReplayCast(cast)
	if err != nil {
		return fromSession()
	}
	drawn := replay.Screen.Trim(frontFrameRows).SVG(runner.DarkPalette)
	return phpfront.Screen{SVG: drawn, Lines: replay.Lines(), Complete: replay.Complete()}
}

// Expired reports whether a run names a recording that is not on disk, which
// is what the retention sweep leaves behind and no column records.
func (h *FrontHost) Expired(_ context.Context, castPath, state string) bool {
	if h.Recordings == nil {
		return false
	}
	return h.Recordings.IsExpired(&model.Run{CastPath: castPath, State: state})
}

// Claude reports whether the reader has a token for Claude. The tokens are
// the integration store's secrets, which the front end never reads itself.
func (h *FrontHost) Claude(ctx context.Context) bool {
	user := access.UserFrom(ctx)
	if h.Integrations == nil || user == nil {
		return false
	}
	env, err := h.Integrations.Env(ctx, user.Name, user.EffectiveRole())
	if err != nil {
		return false
	}
	return env["CLAUDE_CODE_OAUTH_TOKEN"] != "" || env["ANTHROPIC_AUTH_TOKEN"] != ""
}

// held is the session of a run, where the runner holds one.
func (h *FrontHost) held(id string) (*runner.Session, bool) {
	if h.Runner == nil || id == "" {
		return nil, false
	}
	return h.Runner.Get(id)
}
