package frontend

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"sort"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// frameRows is the least rows a frame on a page shows.
const frameRows = 8

// running returns the live runs the user sees, by job key.
func (h *Handlers) running(sees func(*model.Run) bool) map[string]view.Run {
	result := map[string]view.Run{}
	for _, session := range h.opts.Runner.List() {
		run := session.Snapshot()
		if !sees(&run) || run.Finished() {
			continue
		}
		result[run.Key()] = h.runView(run)
	}
	return result
}

// live returns the runs with a session in memory the user sees, newest
// first, each with its frame: running ones, and the ones that just
// finished.
func (h *Handlers) live(sees func(*model.Run) bool) []view.Run {
	var result []view.Run
	for _, session := range h.opts.Runner.List() {
		run := session.Snapshot()
		if sees(&run) {
			result = append(result, h.framed(h.runView(run)))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID > result[j].ID })
	return result
}

// runView decorates a run row with what the runner knows.
func (h *Handlers) runView(run model.Run) view.Run {
	v := view.Run{Run: run, Message: run.Message}
	if session, ok := h.opts.Runner.Get(run.ID); ok {
		status := session.Status()
		v.Run, v.Live, v.Attached = session.Snapshot(), true, session.Attached()
		if status.Message != "" {
			v.Message = status.Message
		}
		if at := h.opts.Runner.Position(run.ID); at > 0 {
			v.Message = fmt.Sprintf("waits for a free slot, number %d in line", at)
		}
	}
	return v
}

// framed adds the rendered frame of a live run.
func (h *Handlers) framed(run view.Run) view.Run {
	if session, ok := h.opts.Runner.Get(run.ID); ok {
		run.FrameSVG = template.HTML(session.Frame().Trim(frameRows).SVG(runner.DarkPalette))
	}
	return run
}

// frameSVG renders the last frame of a run: its session's while that
// is in memory, else the screen its recording ends on. Empty when
// neither is there.
func (h *Handlers) frameSVG(ctx context.Context, runID string) template.HTML {
	svg, _, _ := h.frameOf(ctx, runID)
	return svg
}

// frameOf renders the last frame of a run like frameSVG, with the lines
// of its whole output, and reports whether the frame shows every line:
// a finished run whose output never scrolled off the screen. A run in
// progress may still scroll; its lines are not counted (0).
func (h *Handlers) frameOf(ctx context.Context, runID string) (template.HTML, int, bool) {
	session, live := h.opts.Runner.Get(runID)
	fromSession := func() (template.HTML, int, bool) {
		if !live {
			return "", 0, false
		}
		return template.HTML(session.Frame().Trim(frameRows).SVG(runner.DarkPalette)), 0, false
	}
	if live && !session.Status().Finished() || h.opts.Recordings == nil {
		return fromSession()
	}
	run, err := h.opts.Runs.Get(ctx, runID)
	if err != nil {
		return fromSession()
	}
	cast, err := h.opts.Recordings.Read(run)
	if errors.Is(err, recording.ErrExpired) {
		return template.HTML(runner.TextFrame(int(run.TermCols), frameRows, recording.Expired).SVG(runner.DarkPalette)), 0, true
	}
	if err != nil {
		return fromSession()
	}
	defer cast.Close()
	replay, err := runner.ReplayCast(cast)
	if err != nil {
		return fromSession()
	}
	return template.HTML(replay.Screen.Trim(frameRows).SVG(runner.DarkPalette)), replay.Lines(), replay.Complete()
}

// visibleRuns lists the runs of a filter the user sees.
func (h *Handlers) visibleRuns(ctx context.Context, sees func(*model.Run) bool, filter storage.RunFilter) []view.Run {
	runs, err := h.opts.Runs.List(ctx, filter)
	if err != nil {
		slog.Error("listing runs", "err", err)
		return nil
	}
	result := []view.Run{}
	for _, run := range runs {
		if sees(&run) {
			result = append(result, h.runView(run))
		}
	}
	return result
}
