package frontend

import (
	"net/http"
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

// jobTable fills JobGroups and JobSummaries from place, the listing's files and directories of no folder, and the page's Actions.
func (h *Handlers) jobTable(r *http.Request, p *view.Page, f, place *folder.Folder) {
	var files []*folder.Job
	var dirs []*folder.Folder
	looked := 0
	for _, item := range p.Tree {
		if item.Sealed || item.Path == "" {
			continue
		}
		if !item.Dir {
			if item.Exec {
				if job := f.Job(folder.FilePrefix + item.Path); job != nil {
					files = append(files, job)
				}
			}
			continue
		}
		if p.TreeFolders[item.Name] != nil || looked >= treeDetect {
			continue
		}
		looked++
		if sub := h.subs.get(f, item.Path); sub != nil && len(sub.Jobs) > 0 {
			dirs = append(dirs, sub)
		}
	}
	p.JobGroups = view.NewJobGroups(f, place, files, p.Actions, dirs)
	h.jobSummaries(r, p)
}

// jobSummaries puts the summaries of the last runs of JobGroups' jobs in JobSummaries, for a user with prompt/view.
// None without a board or permission, nor without a last run or a listing of the threads.
func (h *Handlers) jobSummaries(r *http.Request, p *view.Page) {
	p.JobSummaries = map[string]*view.SummaryLink{}
	if h.opts.Board == nil || !p.Can("prompt/view") {
		return
	}
	keys := map[string]string{}
	var ids []string
	for _, g := range p.JobGroups {
		for _, job := range g.Jobs {
			if run := p.LastRun(job.Key()); run != nil {
				keys[run.ID] = job.Key()
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
	byThread := map[string]string{}
	var threadIDs []string
	for i := range threads {
		if note.IsSummary(&threads[i]) {
			byThread[threads[i].ID] = threads[i].AboutRunID
			threadIDs = append(threadIDs, threads[i].ID)
		}
	}
	replies, _ := h.opts.Board.Notes.Replies(r.Context(), threadIDs)
	for thread, runID := range byThread {
		text := ""
		if answers := replies[thread]; len(answers) > 0 {
			text = answers[0].Body
		}
		p.JobSummaries[keys[runID]] = view.NewSummaryLink(runID, text)
	}
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
