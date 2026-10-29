package retention

import (
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// Cleaner measures what each area of data holds and applies the rules,
// one area after another in the order of Areas: the recording files of
// finished runs, the rows of old runs with their recordings, summaries
// and threads, old prompt threads, old workflow runs with their steps,
// the earlier versions of the artifact tree, never the latest of a file,
// and the reads of the secret read log; then the expired tokens and
// sessions, which go every time. A run in progress or queued, a thread
// with an open instruction, a workflow run in progress and the cleanup's
// own run always stay.
type Cleaner struct {
	// DB is the database, which a clean that removed rows vacuums, and
	// DBPath its file, for the size before and after; empty skips both.
	DB     storage.DB
	DBPath string
	// Values keep the rules.
	Values      *storage.SettingStorage
	Runs        *storage.RunStorage
	Summaries   *storage.RunSummaryStorage
	Notes       *storage.NoteStorage
	Flows       *storage.WorkflowRunStorage
	Artifacts   *storage.ArtifactStorage
	Store       *artifactstore.Store
	SecretReads *storage.SecretReadStorage
	Tokens      *storage.UserAuthTokenStorage
	Sessions    *storage.UserSessionStorage
	// RecordingsDir holds the recordings the run rows name.
	RecordingsDir string
	// Tooling lists the atkins.db files of the workspaces' homes and the
	// instance, which the Tool history row measures; nil measures none.
	Tooling func() []string
	// Atkins runs the atkins binary with arguments and returns its output;
	// nil, when atkins is not on PATH, measures the databases and prunes none.
	Atkins func(ctx context.Context, args ...string) ([]byte, error)
	// Self is the ID of the run the cleanup runs in, which stays.
	Self string
	// Now is the clock, time.Now unless set.
	Now func() time.Time
}

// Load returns the rules, the defaults for what no one set.
func (c *Cleaner) Load(ctx context.Context) (Settings, error) {
	if c.Values == nil {
		return Defaults(), nil
	}
	values, err := c.Values.Values(ctx)
	if err != nil {
		return Defaults(), err
	}
	return Load(values), nil
}

// Save stores the rules; the next clean applies them.
func (c *Cleaner) Save(ctx context.Context, s Settings) error {
	return c.Values.Set(ctx, s.Values())
}

// Usage returns what each area holds, by kind, the fixed areas too.
func (c *Cleaner) Usage(ctx context.Context) (map[string]Usage, error) {
	usage := map[string]Usage{}
	for _, area := range Areas {
		items, err := c.items(ctx, area.Kind)
		if err != nil {
			return nil, err
		}
		usage[area.Kind] = usageOf(items)
	}
	return usage, nil
}

// Clean applies the rules area by area, oldest first, and prints what
// each lost and holds after; dry prints what would go and changes
// nothing. The expired sign-ins go every time, and atkins prunes the
// tool history, a line per database. Once it removed rows it
// vacuums the database and prints its size before and after.
func (c *Cleaner) Clean(ctx context.Context, dry bool, out io.Writer) error {
	settings, err := c.Load(ctx)
	if err != nil {
		return err
	}
	now := c.now()
	rows := 0
	for _, area := range Areas {
		items, err := c.items(ctx, area.Kind)
		if err != nil {
			return err
		}
		if area.Kind == Tooling {
			if err := c.prune(ctx, area, settings.Rule(area.Kind), items, dry, out); err != nil {
				return err
			}
			continue
		}
		rule, says, expired := settings.Rule(area.Kind), "", items
		switch {
		case area.Kind == Signins:
			says = "removed every cleanup"
		case area.Sized:
			says = rule.Says()
			expired = Expire(items, rule, now)
		default:
			says = rule.Keeps()
			expired = Expire(items, rule, now)
		}
		gone, all := usageOf(expired), usageOf(items)
		left := Usage{Count: all.Count - gone.Count, Bytes: all.Bytes - gone.Bytes}
		verb := "would remove"
		if !dry {
			verb = "removed"
			if len(expired) > 0 {
				removed, err := c.remove(ctx, area.Kind, expired, now)
				if err != nil {
					return fmt.Errorf("%s: %w", area.Title, err)
				}
				rows += removed
			}
			after, err := c.items(ctx, area.Kind)
			if err != nil {
				return err
			}
			left = usageOf(after)
		}
		fmt.Fprintf(out, "%s (%s): %s %s; %s left\n", area.Title, says, verb, area.Holds(gone), area.Holds(left))
	}
	if rows == 0 || dry || c.DB == nil {
		return nil
	}
	before := c.dbSize()
	if err := storage.Vacuum(ctx, c.DB); err != nil {
		return fmt.Errorf("vacuum: %w", err)
	}
	if c.DBPath != "" {
		fmt.Fprintf(out, "Database: vacuumed, %s before, %s after\n", FormatSize(before), FormatSize(c.dbSize()))
	}
	return nil
}

func (c *Cleaner) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

// dbSize is the database file's size with its write-ahead log.
func (c *Cleaner) dbSize() int64 {
	var size int64
	for _, suffix := range []string{"", "-wal"} {
		if info, err := os.Stat(c.DBPath + suffix); err == nil {
			size += info.Size()
		}
	}
	return size
}

// items returns what an area holds, oldest first.
func (c *Cleaner) items(ctx context.Context, kind string) ([]Item, error) {
	switch kind {
	case Recordings:
		return c.recordings(ctx)
	case Runs:
		return c.runs(ctx)
	case Prompts:
		return c.threads(ctx)
	case WorkflowRuns:
		return c.flows(ctx)
	case Artifacts:
		return c.history(ctx)
	case Signins:
		return c.signins(ctx)
	case Tooling:
		return c.tooling(), nil
	}
	return c.reads(ctx)
}

// threads are the prompt threads by their last reply, one with an open
// or working instruction busy.
func (c *Cleaner) threads(ctx context.Context) ([]Item, error) {
	if c.Notes == nil {
		return nil, nil
	}
	threads, err := c.Notes.Threads(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(threads))
	for _, thread := range threads {
		item := Item{ID: thread.ID, Busy: thread.Open}
		if thread.LastAt != nil {
			item.At = *thread.LastAt
		}
		items = append(items, item)
	}
	return items, nil
}

// flows are the workflow runs, one in progress busy.
func (c *Cleaner) flows(ctx context.Context) ([]Item, error) {
	if c.Flows == nil {
		return nil, nil
	}
	runs, err := c.Flows.History(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(runs))
	for _, run := range runs {
		item := Item{ID: run.ID, Busy: !model.Finished(run.State)}
		if run.StartedAt != nil {
			item.At = *run.StartedAt
		}
		items = append(items, item)
	}
	return items, nil
}

// history are the earlier versions of the artifact tree's files.
func (c *Cleaner) history(ctx context.Context) ([]Item, error) {
	if c.Artifacts == nil || c.Store == nil {
		return nil, nil
	}
	records, err := c.Artifacts.History(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(records))
	for _, record := range records {
		at := time.Time{}
		if record.CreatedAt != nil {
			at = *record.CreatedAt
		}
		items = append(items, Item{ID: record.CommitID, At: at, Size: record.ByteSize})
	}
	return items, nil
}

// reads are the reads of the secret read log.
func (c *Cleaner) reads(ctx context.Context) ([]Item, error) {
	if c.SecretReads == nil {
		return nil, nil
	}
	reads, err := c.SecretReads.All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(reads))
	for _, read := range reads {
		item := Item{ID: read.ID}
		if read.CreatedAt != nil {
			item.At = *read.CreatedAt
		}
		items = append(items, item)
	}
	return items, nil
}

// signins are the tokens and sessions past their expiry.
func (c *Cleaner) signins(ctx context.Context) ([]Item, error) {
	items := []Item{}
	now := c.now()
	if c.Tokens != nil {
		tokens, err := c.Tokens.Expired(ctx, now)
		if err != nil {
			return nil, err
		}
		for _, token := range tokens {
			items = append(items, Item{ID: "token:" + token.ID, At: *token.ExpiresAt})
		}
	}
	if c.Sessions != nil {
		sessions, err := c.Sessions.Expired(ctx, now)
		if err != nil {
			return nil, err
		}
		for _, session := range sessions {
			items = append(items, Item{ID: "session:" + session.ID, At: *session.ExpiresAt})
		}
	}
	return items, nil
}

// remove takes the items out of an area and returns how many rows went.
func (c *Cleaner) remove(ctx context.Context, kind string, items []Item, now time.Time) (int, error) {
	switch kind {
	case Recordings:
		return 0, c.removeRecordings(items)
	case Runs:
		return c.removeRuns(ctx, items)
	case Prompts:
		for _, item := range items {
			if err := c.Notes.Delete(ctx, item.ID); err != nil {
				return 0, err
			}
		}
		return len(items), nil
	case WorkflowRuns:
		for _, item := range items {
			if err := c.Notes.DeleteAboutWorkflowRun(ctx, item.ID); err != nil {
				return 0, err
			}
			if err := c.Flows.Delete(ctx, item.ID); err != nil {
				return 0, err
			}
		}
		return len(items), nil
	case Artifacts:
		return c.expireHistory(ctx, items)
	case Signins:
		return len(items), c.removeSignins(ctx, now)
	}
	if err := c.SecretReads.DeleteThrough(ctx, items[len(items)-1].ID); err != nil {
		return 0, err
	}
	return len(items), nil
}

// removeSignins deletes the tokens and sessions past their expiry.
func (c *Cleaner) removeSignins(ctx context.Context, now time.Time) error {
	if c.Tokens != nil {
		if err := c.Tokens.DeleteExpired(ctx, now); err != nil {
			return err
		}
	}
	if c.Sessions != nil {
		return c.Sessions.DeleteExpired(ctx, now)
	}
	return nil
}

// expireHistory drops the history of the artifact tree up to the newest
// commit that holds a version to go: that commit's successor becomes the
// first. The versions of the dropped commits go with their records,
// but the newest of a path that the kept tree holds unchanged, which
// moves to the new first commit.
func (c *Cleaner) expireHistory(ctx context.Context, items []Item) (int, error) {
	commits, err := c.Store.Commits(ctx)
	if err != nil || len(commits) < 2 {
		return 0, err
	}
	last := -1
	for _, item := range items {
		last = max(last, slices.Index(commits, item.ID))
	}
	keep := commits[min(last+1, len(commits)-1)]
	removed := 0
	err = c.Store.Expire(ctx, keep, func(r artifactstore.Rewrite) error {
		latest, err := c.Artifacts.Latest(ctx)
		if err != nil {
			return err
		}
		newest := map[string]bool{}
		for _, record := range latest {
			newest[record.ID] = true
		}
		removed, err = c.Artifacts.Rewrite(ctx, r.Root, r.Renamed, r.Dropped, func(record model.Artifact) bool {
			return newest[record.ID] && r.Same(record.CommitID, record.Path)
		})
		return err
	})
	return removed, err
}
