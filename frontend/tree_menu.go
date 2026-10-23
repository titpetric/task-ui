package frontend

import (
	"net/http"
	"path"
	"slices"
	"sync"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/storage"
)

// treeDetect caps the directories of one Files list detected for their
// job sources.
const treeDetect = 200

// treeMenus puts the Run menus of a Files list's rows in TreeMenus: a
// directory that is a folder gets its jobs, and with detect, one that is
// not gets the jobs of the job source Folder.Sub finds at its top, run
// as f's in that directory; a file that is a job source of place, the
// folder the list shows, gets that source's targets. It reports whether
// it found the jobs of such a directory, whose last runs are f's.
func (h *Handlers) treeMenus(p *view.Page, f, place *folder.Folder, detect bool) bool {
	p.TreeMenus = map[string]*view.RunMenu{}
	detected, looked := false, 0
	for _, item := range p.Tree {
		if item.Sealed {
			continue
		}
		if item.Dir {
			if tf := p.TreeFolders[item.Name]; tf != nil {
				if m := view.NewRunMenu(tf, tf.Jobs, nil); m != nil {
					p.TreeMenus[item.Name] = m
				}
				continue
			}
			if !detect || item.Path == "" || looked >= treeDetect {
				continue
			}
			looked++
			if sub := h.subs.get(f, item.Path); sub != nil && len(sub.Jobs) > 0 {
				p.TreeMenus[item.Name] = view.NewRunMenu(sub, sub.Jobs, nil)
				detected = true
			}
			continue
		}
		if place == nil || item.Path == "" {
			continue
		}
		for _, source := range place.Sources {
			if path.Join(place.Rel, source.File) != item.Path {
				continue
			}
			jobs := place.SourceJobs(source)
			var main *folder.Job
			if !slices.ContainsFunc(jobs, (*folder.Job).IsDefault) && item.Exec {
				main = f.Job(folder.FilePrefix + item.Path)
			}
			if m := view.NewRunMenu(place, jobs, main); m != nil {
				p.TreeMenus[item.Name] = m
			}
			break
		}
	}
	return detected
}

// treeSummaries puts the summaries of the rows' last runs in
// TreeSummaries, for a user with prompt/view: of each row's jobs, the
// newest last run that Claude summarized with Run and summarize. Call
// it once Last and TreeMenus hold the rows' runs and jobs.
func (h *Handlers) treeSummaries(r *http.Request, p *view.Page) {
	p.TreeSummaries = map[string]*view.SummaryLink{}
	if h.opts.Board == nil || !p.Can("prompt/view") {
		return
	}
	rows := map[string][]*view.Run{}
	var ids []string
	for _, item := range p.Tree {
		var jobs []*folder.Job
		if m := p.TreeMenus[item.Name]; m != nil {
			jobs = append([]*folder.Job{m.Main}, m.Jobs...)
		} else if !item.Dir && item.Exec && item.Path != "" && p.Folder != nil {
			jobs = []*folder.Job{p.Folder.Job(folder.FilePrefix + item.Path)}
		}
		for _, job := range jobs {
			if job == nil {
				continue
			}
			if run := p.LastRun(job.Key()); run != nil {
				rows[item.Name] = append(rows[item.Name], run)
				ids = append(ids, run.ID)
			}
		}
	}
	if len(ids) == 0 {
		return
	}
	threads, err := h.opts.Board.Notes.List(r.Context(), storage.NoteFilter{Prompts: true, AboutRunIDs: ids})
	if err != nil {
		return
	}
	byRun := map[string]string{}
	var threadIDs []string
	for i := range threads {
		if note.IsSummary(&threads[i]) {
			byRun[threads[i].AboutRunID] = threads[i].ID
			threadIDs = append(threadIDs, threads[i].ID)
		}
	}
	replies, _ := h.opts.Board.Notes.Replies(r.Context(), threadIDs)
	for name, runs := range rows {
		var newest *view.Run
		for _, run := range runs {
			if _, ok := byRun[run.ID]; ok && (newest == nil || newer(run, newest)) {
				newest = run
			}
		}
		if newest == nil {
			continue
		}
		thread := byRun[newest.ID]
		text := ""
		if answers := replies[thread]; len(answers) > 0 {
			text = answers[0].Body
		}
		p.TreeSummaries[name] = view.NewSummaryLink(newest.ID, thread, text)
	}
}

// newer reports whether run a started after run b.
func newer(a, b *view.Run) bool {
	return a.StartedAt != nil && (b.StartedAt == nil || a.StartedAt.After(*b.StartedAt))
}

// subTTL is how long a directory's detected sources hold for the Files
// list: long enough for a page and the reloads after it, short enough
// that a new file shows on the next look.
const subTTL = 10 * time.Second

// subCache keeps what Folder.Sub read of the directories a Files list
// shows, for subTTL, so a page view reads each directory once.
type subCache struct {
	mu      sync.Mutex
	entries map[string]subEntry
}

type subEntry struct {
	sub     *folder.Folder
	expires time.Time
}

// get returns the folder of the directory rel inside f, as Folder.Sub
// reads it, from the cache while it holds.
func (c *subCache) get(f *folder.Folder, rel string) *folder.Folder {
	key := f.ID + "\x00" + f.Dir + "\x00" + rel
	now := time.Now()
	c.mu.Lock()
	entry, ok := c.entries[key]
	c.mu.Unlock()
	if ok && now.Before(entry.expires) {
		return entry.sub
	}
	sub := f.Sub(rel)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.entries == nil {
		c.entries = map[string]subEntry{}
	}
	for k, e := range c.entries {
		if !now.Before(e.expires) {
			delete(c.entries, k)
		}
	}
	c.entries[key] = subEntry{sub: sub, expires: now.Add(subTTL)}
	return sub
}
