package frontend

import (
	"path"
	"slices"
	"sync"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
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
