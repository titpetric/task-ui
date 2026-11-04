package app

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/phpfront"
)

// frontDetect caps the directories of one Jobs table detected for their job
// sources, the cap frontend/job_table.go draws the same table under.
const frontDetect = 200

// frontSubTTL is how long a directory's detected sources hold: long enough
// for a page and the reloads after it, short enough that a new file shows on
// the next look.
const frontSubTTL = 10 * time.Second

// frontSubs keeps what Folder.Sub read of the directories a Jobs table
// probes, so one page view reads each directory once.
type frontSubs struct {
	mu      sync.Mutex
	entries map[string]frontSub
}

// frontSub is one directory's detection and when it stops holding.
type frontSub struct {
	sub     *folder.Folder
	expires time.Time
}

// get returns the folder of the directory rel inside f, as Folder.Sub reads
// it, from the cache while it holds.
func (c *frontSubs) get(f *folder.Folder, rel string) *folder.Folder {
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
		c.entries = map[string]frontSub{}
	}
	for k, e := range c.entries {
		if !now.Before(e.expires) {
			delete(c.entries, k)
		}
	}
	c.entries[key] = frontSub{sub: sub, expires: now.Add(frontSubTTL)}
	return sub
}

// Place is a folder of the catalogue, or a directory inside one, in the depth
// the page about it draws.
func (h *FrontHost) Place(_ context.Context, id, rel string) phpfront.Place {
	place := h.place(id, rel)
	if place == nil {
		return phpfront.Place{}
	}
	sources := make([]phpfront.Source, 0, len(place.Sources))
	for _, source := range place.Sources {
		if source == nil {
			continue
		}
		out := phpfront.Source{Kind: source.Kind, File: source.File, Command: source.Command}
		// A page draws the button that shows a pipeline file and never
		// one that shows a script, so a script's lines are a read no
		// page pays for.
		if source.Kind != "script" {
			out.Lines = place.Lines(source.File)
		}
		sources = append(sources, out)
	}
	includes := make([]phpfront.File, 0, len(place.Includes))
	for _, name := range place.Includes {
		includes = append(includes, phpfront.File{Name: name, Lines: place.Lines(name)})
	}
	return phpfront.Place{
		Found:       true,
		ID:          place.ID,
		Title:       place.Title,
		Description: place.Description,
		Dir:         place.Dir,
		Rel:         place.Rel,
		Allow:       place.Allow,
		Hours:       place.Hours,
		Error:       place.Error,
		Errors:      frontFindings(place.Errors),
		Sources:     sources,
		Includes:    includes,
	}
}

// place is the folder of the catalogue, or the directory rel inside it as
// Folder.Sub reads it; nil for either not being there.
func (h *FrontHost) place(id, rel string) *folder.Folder {
	if h.Catalogue == nil {
		return nil
	}
	f := h.Catalogue.Find(id)
	if f == nil {
		return nil
	}
	return f.Sub(rel)
}

// Groups is the Jobs table of a place, as frontend/job_table.go builds it off
// the same listing: the place's sources, the files at the top of the folder
// that run, the skills, the terminals, then the directories inside that run
// jobs of their own.
func (h *FrontHost) Groups(_ context.Context, id, rel string, seen []string, shell, claude bool) []phpfront.Group {
	if h.Catalogue == nil {
		return nil
	}
	f := h.Catalogue.Find(id)
	if f == nil {
		return nil
	}
	dir := f.Dir
	if rel != "" {
		found, ok := f.Path(rel)
		if !ok {
			return nil
		}
		dir = found
	}
	place := f.Sub(rel)
	byDir := h.frontByDir(seen)
	base := frontAbs(dir)
	var files []*folder.Job
	var dirs []*folder.Folder
	looked := 0
	for _, item := range folder.Listing(dir) {
		if item.Sealed {
			continue
		}
		inner := path.Join(rel, item.Name)
		if _, ok := f.Path(inner); !ok {
			continue
		}
		if !item.Dir {
			if !item.Exec {
				continue
			}
			if job := f.Job(folder.FilePrefix + inner); job != nil {
				files = append(files, job)
			}
			continue
		}
		if child := byDir[filepath.Join(base, item.Name)]; child != nil {
			if len(child.Jobs) > 0 {
				dirs = append(dirs, child)
			}
			continue
		}
		if looked >= frontDetect {
			continue
		}
		looked++
		sub := h.subs.get(f, inner)
		if sub != nil && len(sub.Jobs) > 0 {
			dirs = append(dirs, sub)
		}
	}
	var actions *view.Actions
	if place != nil {
		actions = &view.Actions{Place: place, Self: h.Compose, Skills: place.Skills()}
		if shell {
			actions.Shell = place.ShellJob()
			if claude {
				actions.Claude = place.ClaudeJob()
			}
		}
	}
	return frontGroups(actions, view.NewJobGroups(f, place, files, actions, dirs))
}

// frontGroups is the Jobs table's groups as the front host answers them: the
// folder a row runs as, and the row's jobs with what each asks before it runs.
func frontGroups(actions *view.Actions, found []view.JobGroup) []phpfront.Group {
	out := make([]phpfront.Group, 0, len(found))
	for _, g := range found {
		group := phpfront.Group{
			Kind:    g.Kind,
			Note:    g.Note,
			Source:  g.Source,
			Dir:     g.Dir,
			Empty:   g.Empty,
			Targets: g.Targets,
			Jobs:    make([]phpfront.Job, 0, len(g.Jobs)),
		}
		if g.Folder != nil {
			group.Folder = phpfront.Where{ID: g.Folder.ID, Rel: g.Folder.Rel, Allow: g.Folder.Allow, Hours: g.Folder.Hours}
		}
		for _, job := range g.Jobs {
			target := frontFullJob(job)
			target.Confirm = actions.Confirm(g.Skill, job)
			group.Jobs = append(group.Jobs, target)
		}
		out = append(out, group)
	}
	return out
}

// Listing is a directory as the Files panel draws it: a folder's own, or a
// directory inside it.
func (h *FrontHost) Listing(_ context.Context, id, rel string, seen []string) []phpfront.Entry {
	if h.Catalogue == nil {
		return nil
	}
	f := h.Catalogue.Find(id)
	if f == nil {
		return nil
	}
	dir := f.Dir
	if rel != "" {
		found, ok := f.Path(rel)
		if !ok {
			return nil
		}
		dir = found
	}
	byDir := h.frontByDir(seen)
	base := frontAbs(dir)
	items := folder.Listing(dir)
	out := make([]phpfront.Entry, 0, len(items))
	for _, item := range items {
		entry := phpfront.Entry{
			Name:    item.Name,
			Dir:     item.Dir,
			Size:    item.Size,
			ModTime: frontTime(item.ModTime),
			Sealed:  item.Sealed,
			Exec:    item.Exec,
		}
		if !item.Sealed {
			inner := path.Join(rel, item.Name)
			if _, ok := f.Path(inner); ok {
				entry.Path = inner
			} else {
				// A symlink out of the folder, or a dangling one.
				entry.Exec = false
			}
			if item.Dir {
				if child := byDir[filepath.Join(base, item.Name)]; child != nil {
					entry.Folder = child.ID
				}
			}
		}
		out = append(out, entry)
	}
	return out
}

// RootTree is the root's own directories, as frontend/folder.go lists the
// directory a workspace of no name keeps its checkouts in.
//
// The root is not a folder of the catalogue, so Listing cannot answer it: it
// is the tree the catalogue was discovered from. A workspace row that names
// the root moves the listing into the src of the tree, which is what named
// says; a tree with no row for it lists the tree itself.
//
// No entry carries a Path, because the root's page links a directory to the
// folder it is and never opens a file of its own.
func (h *FrontHost) RootTree(_ context.Context, named bool, seen []string) []phpfront.Entry {
	dir := h.Root
	if named {
		dir = layout.SrcDir(h.Root, model.Default)
	}
	byDir := h.frontByDir(seen)
	base := frontAbs(dir)
	items := folder.Listing(dir)
	out := make([]phpfront.Entry, 0, len(items))
	for _, item := range items {
		entry := phpfront.Entry{
			Name:    item.Name,
			Dir:     item.Dir,
			Size:    item.Size,
			ModTime: frontTime(item.ModTime),
			Sealed:  item.Sealed,
			Exec:    item.Exec,
		}
		if item.Dir && !item.Sealed {
			if child := byDir[filepath.Join(base, item.Name)]; child != nil {
				entry.Folder = child.ID
			}
		}
		out = append(out, entry)
	}
	return out
}

// frontByDir is the folders the reader sees by the directory each is, the way
// frontend/folder.go matches a listing's directories against the catalogue.
func (h *FrontHost) frontByDir(seen []string) map[string]*folder.Folder {
	if h.Catalogue == nil {
		return nil
	}
	sees := make(map[string]bool, len(seen))
	for _, id := range seen {
		sees[id] = true
	}
	byDir := map[string]*folder.Folder{}
	for _, f := range h.Catalogue.List() {
		if sees[f.ID] {
			byDir[filepath.Clean(f.Dir)] = f
		}
	}
	return byDir
}

// frontAbs absolutises a directory: discovery absolutises the root, so every
// folder of the catalogue holds an absolute directory, and an instance
// configured with a relative root builds the root's own from that.
func frontAbs(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return dir
	}
	return abs
}

// frontTime is a moment as Task\Clock reads a stamp; a moment nothing read is
// empty.
func frontTime(at time.Time) string {
	if at.IsZero() {
		return ""
	}
	return at.Format(frontStamp)
}

// ModTime is when a directory last changed, as Task\Clock reads a stamp.
func (h *FrontHost) ModTime(_ context.Context, dir string) string {
	if dir == "" {
		return ""
	}
	info, err := os.Stat(dir)
	if err != nil {
		return ""
	}
	return frontTime(info.ModTime())
}

// Status is what git reports of the checkout a place is in.
func (h *FrontHost) Status(_ context.Context, dir string) phpfront.Status {
	if h.Git == nil || h.Root == "" || dir == "" {
		return phpfront.Status{}
	}
	found := h.Git.Status(dir, h.Root)
	if found == nil {
		return phpfront.Status{}
	}
	out := phpfront.Status{
		Found:    true,
		Checkout: found.Checkout,
		Branch:   found.Branch,
		Detached: found.Detached,
		Upstream: found.Upstream,
		Ahead:    found.Ahead,
		Behind:   found.Behind,
		Dirty:    found.Dirty,
		Changes:  make([]phpfront.Change, 0, len(found.Changes)),
	}
	for _, change := range found.Changes {
		out.Changes = append(out.Changes, phpfront.Change{
			Code: change.Code(), State: change.State(), Path: change.Path, From: change.From,
		})
	}
	return out
}

// Checkout is the commit a folder's checkout is at and the space that
// checkout takes, measured from the top of the checkout the commit was read
// from.
func (h *FrontHost) Checkout(_ context.Context, dir string) phpfront.Checkout {
	if h.Git == nil || dir == "" || !h.Git.IsRepo(dir) {
		return phpfront.Checkout{}
	}
	out := phpfront.Checkout{}
	if info := h.Git.Info(dir); info != nil {
		out.Found = true
		out.Remote, out.Branch, out.Commit, out.Subject = info.Remote, info.Branch, info.Commit, info.Subject
		out.When = frontTime(info.When)
	}
	if h.Disk == nil {
		return out
	}
	if root := h.Git.Checkout(dir, 8); root != "" {
		usage := h.Disk.Measure(root)
		out.Space = phpfront.Space{
			Measured: true, Files: usage.Files, History: usage.History, Ignored: usage.Ignored,
		}
	}
	return out
}

// Hours is an hours rule read: the one line a page prints it as, and whether
// it is open at the instance's moment, which is the one view.Page.InHours
// reads it at.
func (h *FrontHost) Hours(_ context.Context, text string) phpfront.Hours {
	if text == "" {
		return phpfront.Hours{Open: true}
	}
	now := time.Now
	if h.Now != nil {
		now = h.Now
	}
	return phpfront.Hours{Line: view.HoursLine(text), Open: view.InHours(text, now())}
}
