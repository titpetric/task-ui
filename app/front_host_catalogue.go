package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/skills"
	"github.com/titpetric/task-ui/frontend"
	"github.com/titpetric/task-ui/internal/detector"
	"github.com/titpetric/task-ui/internal/folder"
	"github.com/titpetric/task-ui/layout"
)

// Folders is the catalogue: each folder and the jobs it declares.
func (h *FrontHost) Folders(context.Context) []phpfront.Folder {
	if h.Catalogue == nil {
		return nil
	}
	found := h.Catalogue.List()
	out := make([]phpfront.Folder, 0, len(found))
	for _, f := range found {
		jobs := make([]phpfront.Job, 0, len(f.Jobs))
		for _, job := range f.Jobs {
			jobs = append(jobs, frontJob(job))
		}
		out = append(out, phpfront.Folder{
			ID:      f.ID,
			Title:   f.Title,
			Dir:     f.Dir,
			Allow:   f.Allow,
			Sources: frontSources(f.Sources),
			Error:   f.Error,
			Errors:  frontFindings(f.Errors),
			Jobs:    jobs,
		})
	}
	return out
}

// frontSources is the files a folder's conventions come from, the names its
// sidebar row lists.
func frontSources(sources []*detector.Source) []string {
	out := make([]string, 0, len(sources))
	for _, source := range sources {
		if source != nil {
			out = append(out, source.File)
		}
	}
	return out
}

// frontFindings is what atkins' linter said about a folder's pipelines.
func frontFindings(found []skills.Finding) []phpfront.Finding {
	out := make([]phpfront.Finding, 0, len(found))
	for _, finding := range found {
		out = append(out, phpfront.Finding{
			Path: finding.Path, Job: finding.Job, Issue: finding.Issue, Detail: finding.Detail,
		})
	}
	return out
}

// Projects is the folder IDs of the directories right inside a named
// workspace's src: a checkout with no pipeline file is in no catalogue and
// the sidebar lists it anyway.
func (h *FrontHost) Projects(_ context.Context, workspace string) []string {
	if h.Root == "" || workspace == "" {
		return nil
	}
	entries, err := os.ReadDir(layout.SrcDir(h.Root, workspace))
	if err != nil {
		return nil
	}
	var ids []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			ids = append(ids, workspace+"/"+entry.Name())
		}
	}
	return ids
}

// FolderDir is the directory a folder ID names, the way frontend/side_dirty.go
// answers it: the one the catalogue holds, else where the layout puts the
// project among the named workspaces.
func (h *FrontHost) FolderDir(_ context.Context, id string, workspaces []string) string {
	if id == "" {
		return ""
	}
	if h.Catalogue != nil {
		if f := h.Catalogue.Find(id); f != nil && f.Dir != "" {
			return f.Dir
		}
	}
	if h.Root == "" {
		return ""
	}
	return layout.Dir(h.Root, workspaces, id)
}

// FolderDirs is FolderDir of every ID a tree draws, in one crossing: the
// catalogue's list is indexed once and the layout answers the IDs no folder of
// it holds, so a tree of fifty rows costs one scan.
//
// An ID the layout places nowhere is left out of the map. No folder has an
// empty directory, so a missing key reads as FolderDir's "".
func (h *FrontHost) FolderDirs(ctx context.Context, ids []string, workspaces []string) map[string]string {
	var known map[string]string
	if h.Catalogue != nil {
		found := h.Catalogue.List()
		known = make(map[string]string, len(found))
		for _, each := range found {
			if each.Dir != "" {
				known[each.ID] = each.Dir
			}
		}
	}
	out := make(map[string]string, len(ids))
	for _, id := range ids {
		if id == "" {
			continue
		}
		canonical := model.Canonical(id)
		if dir, ok := known[canonical]; ok {
			out[id] = dir
			continue
		}
		// internal is the root under its own name, which the catalogue
		// lists under none: one ID of a tree, so it keeps the one branch
		// of FolderDir that is not a lookup.
		if canonical == model.Default {
			if dir := h.FolderDir(ctx, id, workspaces); dir != "" {
				out[id] = dir
			}
			continue
		}
		if h.Root == "" {
			continue
		}
		if dir := layout.Dir(h.Root, workspaces, id); dir != "" {
			out[id] = dir
		}
	}
	return out
}

// IsDir reports whether rel exists under the folder id names, the check a
// pinned directory resolves against.
func (h *FrontHost) IsDir(_ context.Context, id, rel string) bool {
	if h.Catalogue == nil {
		return false
	}
	f := h.Catalogue.Find(id)
	if f == nil {
		return false
	}
	dir, ok := f.Path(rel)
	if !ok {
		return false
	}
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// Commit is the time of the last commit in a directory, through the cache both
// front ends share, formatted as frontStamp in UTC so the sidebar can order a
// workspace's projects as text. It is "" for a directory with no commit.
func (h *FrontHost) Commit(_ context.Context, dir string) string {
	if h.Git == nil || dir == "" {
		return ""
	}
	at := h.Git.Commit(dir)
	if at.IsZero() {
		return ""
	}
	return at.UTC().Format(frontStamp)
}

// frontStamp is Task\Clock::STORED, the shape a timestamp column holds.
const frontStamp = "2006-01-02 15:04:05.999 -0700 MST"

// Job is the job a run names, resolved the way frontend/run.go resolves it:
// the folder id names in the catalogue, the subdirectory dir names under it,
// and the job that subdirectory holds. A missing step answers the zero Job,
// which nobody may run.
func (h *FrontHost) Job(_ context.Context, id, dir, job string) phpfront.Job {
	if h.Catalogue == nil {
		return phpfront.Job{}
	}
	f := h.Catalogue.Find(id)
	if f == nil {
		return phpfront.Job{}
	}
	sub := f.Sub(dir)
	if sub == nil {
		return phpfront.Job{}
	}
	found := sub.Job(job)
	if found == nil {
		return phpfront.Job{}
	}
	out := frontFullJob(found)
	out.Found = true
	out.FolderAllow = sub.Allow
	// The button that shows the source is drawn for a target of a pipeline
	// file of the place itself, so a script, a skill and a job of a
	// directory inside are lines no page reads.
	if found.Kind != "script" && found.Kind != "skill" && found.Dir == "" {
		out.SourceLines = sub.Lines(found.Source)
	}
	for _, text := range found.DependsOn {
		dep := phpfront.Depends{Text: text}
		if first := sub.Job(text); first != nil {
			dep.Found, dep.ID, dep.Name, dep.Dir = true, first.ID, first.Name, first.Dir
		}
		out.Depends = append(out.Depends, dep)
	}
	return out
}

// frontFullJob is one job of the catalogue in the depth its own page and the
// Jobs table draw: what the button says, what a run starts, and what gates it.
// Folders answers the narrow frontJob instead, as the sidebar draws every job
// of every folder and names none of them.
func frontFullJob(job *folder.Job) phpfront.Job {
	if job == nil {
		return phpfront.Job{}
	}
	out := frontJob(job)
	out.Kind = job.Kind
	out.Source = job.Source
	out.Description = job.Description
	out.Summary = job.Summary
	out.Command = job.Command
	out.Steps = job.Steps
	out.Aliases = job.Aliases
	out.Interactive = job.Interactive
	out.TakesArgs = job.TakesArgs
	out.ArgsCommand = job.ArgsCommand()
	out.Key = job.Key()
	return out
}

// frontJob is one job of the catalogue as the front end's gate reads it. The
// two kinds come from the folder package, so the rule that names a terminal
// or a host job is stated once, where the jobs are made.
func frontJob(job *folder.Job) phpfront.Job {
	if job == nil {
		return phpfront.Job{}
	}
	return phpfront.Job{
		ID:    job.ID,
		Name:  job.Name,
		Dir:   job.Dir,
		Shell: job.IsShell(),
		Host:  job.IsHost(),
		Allow: job.Allow,
	}
}

// Dirty counts the changed and untracked files under a checkout directory.
func (h *FrontHost) Dirty(_ context.Context, dir string) int {
	if h.Git == nil || h.Root == "" || dir == "" {
		return 0
	}
	status := h.Git.Status(dir, h.Root)
	if status == nil {
		return 0
	}
	rel, err := filepath.Rel(filepath.Join(h.Root, filepath.FromSlash(status.Checkout)), dir)
	if err != nil {
		return 0
	}
	return status.Under(strings.TrimPrefix(filepath.ToSlash(rel), "."))
}
