package retention

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/titpetric/task-ui/model"
)

// recordings are the recording files that are there, of every run that
// names one, the busy ones marked.
func (c *Cleaner) recordings(ctx context.Context) ([]Item, error) {
	if c.Runs == nil || c.RecordingsDir == "" {
		return nil, nil
	}
	runs, err := c.Runs.Recorded(ctx)
	if err != nil {
		return nil, err
	}
	busy, err := c.Runs.Busy(ctx)
	if err != nil {
		return nil, err
	}
	items := []Item{}
	for _, run := range runs {
		info, err := os.Stat(c.recordingPath(run.CastPath))
		if err != nil {
			continue
		}
		at := info.ModTime()
		if run.StartedAt != nil {
			at = *run.StartedAt
		}
		// The item names the run and its recording, apart by a NUL.
		items = append(items, Item{ID: run.ID + "\x00" + run.CastPath, At: at, Size: info.Size(), Busy: c.busy(&run, busy)})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].At.Before(items[j].At) })
	return items, nil
}

// busy reports whether a run stays whatever the rules say: in progress
// or queued, a step of a workflow run in progress, or the cleanup's own.
func (c *Cleaner) busy(run *model.Run, busy map[string]bool) bool {
	return busy[run.ID] || !model.Finished(run.State) || run.ID == c.Self
}

func (c *Cleaner) recordingPath(rel string) string {
	return filepath.Join(c.RecordingsDir, filepath.FromSlash(rel))
}

// runs are every run, the size of each its recording's, where the file
// is there.
func (c *Cleaner) runs(ctx context.Context) ([]Item, error) {
	if c.Runs == nil {
		return nil, nil
	}
	runs, err := c.Runs.History(ctx)
	if err != nil {
		return nil, err
	}
	busy, err := c.Runs.Busy(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(runs))
	for _, run := range runs {
		item := Item{ID: run.ID + "\x00" + run.CastPath, Busy: c.busy(&run, busy)}
		if run.StartedAt != nil {
			item.At = *run.StartedAt
		}
		if run.CastPath != "" && c.RecordingsDir != "" {
			if info, err := os.Stat(c.recordingPath(run.CastPath)); err == nil {
				item.Size = info.Size()
			}
		}
		items = append(items, item)
	}
	return items, nil
}

// removeRecordings deletes the recording files the items name.
func (c *Cleaner) removeRecordings(items []Item) error {
	for _, item := range items {
		_, rel, _ := strings.Cut(item.ID, "\x00")
		if rel == "" {
			continue
		}
		if err := os.Remove(c.recordingPath(rel)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// removeRuns deletes the runs the items name with their recordings,
// summaries and the threads about them.
func (c *Cleaner) removeRuns(ctx context.Context, items []Item) (int, error) {
	if err := c.removeRecordings(items); err != nil {
		return 0, err
	}
	for _, item := range items {
		id, _, _ := strings.Cut(item.ID, "\x00")
		if c.Summaries != nil {
			if err := c.Summaries.Delete(ctx, id); err != nil {
				return 0, err
			}
		}
		if c.Notes != nil {
			if err := c.Notes.DeleteAboutRun(ctx, id); err != nil {
				return 0, err
			}
		}
		if err := c.Runs.Delete(ctx, id); err != nil {
			return 0, err
		}
	}
	return len(items), nil
}
