package retention

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// Cleaner measures what each kind of data holds and applies the rules:
// it removes the recording files of finished runs, and only those, the
// earlier versions of the artifact tree, never the latest of a file in
// the tree, and the reads of the secret read log.
type Cleaner struct {
	// DB is the database, which a clean that removed rows vacuums, and
	// DBPath its file, for the size before and after; empty skips both.
	DB     storage.DB
	DBPath string
	// Values keep the rules.
	Values      *storage.SettingStorage
	Runs        *storage.RunStorage
	Artifacts   *storage.ArtifactStorage
	Store       *artifactstore.Store
	SecretReads *storage.SecretReadStorage
	// KeyReads is the log of the private keys users read.
	KeyReads *storage.KeyReadStorage
	// RecordingsDir holds the recordings the run rows name.
	RecordingsDir string
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

// Usage returns what each kind holds, by kind.
func (c *Cleaner) Usage(ctx context.Context) (map[string]Usage, error) {
	usage := map[string]Usage{}
	for _, kind := range Kinds {
		items, err := c.items(ctx, kind)
		if err != nil {
			return nil, err
		}
		usage[kind] = usageOf(items)
	}
	return usage, nil
}

// titles name the kinds in what Clean prints.
var titles = map[string]string{Recordings: "Run recordings", Artifacts: "Artifact history", SecretReads: "Secret read log", KeyReads: "Key read log"}

// Clean applies the rules, oldest first, and prints what each kind lost
// and holds after; dry prints what would go and changes nothing. Once it
// removed rows it vacuums the database and prints its size before and
// after.
func (c *Cleaner) Clean(ctx context.Context, dry bool, out io.Writer) error {
	settings, err := c.Load(ctx)
	if err != nil {
		return err
	}
	now := c.now()
	rows := 0
	for _, kind := range Kinds {
		rule := settings.Rule(kind)
		items, err := c.items(ctx, kind)
		if err != nil {
			return err
		}
		if kind == Recordings {
			if err := c.busy(ctx, items); err != nil {
				return err
			}
		}
		expired := Expire(items, rule, now)
		gone, all := usageOf(expired), usageOf(items)
		left := Usage{Count: all.Count - gone.Count, Bytes: all.Bytes - gone.Bytes}
		verb := "would remove"
		if !dry {
			verb = "removed"
			if len(expired) > 0 {
				removed, err := c.remove(ctx, kind, expired)
				if err != nil {
					return fmt.Errorf("%s: %w", titles[kind], err)
				}
				if kind != Recordings {
					rows += removed
				}
			}
			after, err := c.items(ctx, kind)
			if err != nil {
				return err
			}
			left = usageOf(after)
		}
		fmt.Fprintf(out, "%s (%s): %s %s; %s left\n", titles[kind], rule.Says(), verb, gone.Says("item", "items"), left.Says("item", "items"))
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

// items returns what a kind holds, oldest first.
func (c *Cleaner) items(ctx context.Context, kind string) ([]Item, error) {
	switch kind {
	case Recordings:
		return c.recordings(ctx)
	case Artifacts:
		return c.history(ctx)
	case KeyReads:
		return c.keyReads(ctx)
	}
	return c.reads(ctx)
}

// recordings are the recording files that are there, of every run that
// names one.
func (c *Cleaner) recordings(ctx context.Context) ([]Item, error) {
	if c.Runs == nil || c.RecordingsDir == "" {
		return nil, nil
	}
	runs, err := c.Runs.Recorded(ctx)
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
		items = append(items, Item{ID: run.ID + "\x00" + run.CastPath, At: at, Size: info.Size()})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].At.Before(items[j].At) })
	return items, nil
}

func (c *Cleaner) recordingPath(rel string) string {
	return filepath.Join(c.RecordingsDir, filepath.FromSlash(rel))
}

// busy marks the recordings that stay whatever the rules say: of a run
// in progress, or a step of a workflow run in progress.
func (c *Cleaner) busy(ctx context.Context, items []Item) error {
	busy, err := c.Runs.Busy(ctx)
	if err != nil {
		return err
	}
	for i := range items {
		run, _, _ := strings.Cut(items[i].ID, "\x00")
		items[i].Busy = busy[run]
	}
	return nil
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

// reads are the reads of the secret read log, each the size Usage of
// the log counts.
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
		at := time.Time{}
		if read.CreatedAt != nil {
			at = *read.CreatedAt
		}
		size := len(read.ID) + len(read.SecretID) + len(read.Name) + len(read.RunID) + len(read.Folder) + len(read.Dir) + len(read.Job) + len(read.UserName) + 24
		items = append(items, Item{ID: read.ID, At: at, Size: int64(size)})
	}
	return items, nil
}

// keyReads are the reads of the key read log, each the size of its row.
func (c *Cleaner) keyReads(ctx context.Context) ([]Item, error) {
	if c.KeyReads == nil {
		return nil, nil
	}
	reads, err := c.KeyReads.All(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(reads))
	for _, read := range reads {
		at := time.Time{}
		if read.CreatedAt != nil {
			at = *read.CreatedAt
		}
		size := len(read.ID) + len(read.KeyName) + len(read.UserName) + len(read.TokenID) + len(read.Via) + 24
		items = append(items, Item{ID: read.ID, At: at, Size: int64(size)})
	}
	return items, nil
}

// remove takes the items out of a kind and returns how many rows went.
func (c *Cleaner) remove(ctx context.Context, kind string, items []Item) (int, error) {
	switch kind {
	case Recordings:
		for _, item := range items {
			_, rel, _ := strings.Cut(item.ID, "\x00")
			if err := os.Remove(c.recordingPath(rel)); err != nil && !os.IsNotExist(err) {
				return 0, err
			}
		}
		return 0, nil
	case Artifacts:
		return c.expireHistory(ctx, items)
	case KeyReads:
		if err := c.KeyReads.DeleteThrough(ctx, items[len(items)-1].ID); err != nil {
			return 0, err
		}
		return len(items), nil
	}
	if err := c.SecretReads.DeleteThrough(ctx, items[len(items)-1].ID); err != nil {
		return 0, err
	}
	return len(items), nil
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
