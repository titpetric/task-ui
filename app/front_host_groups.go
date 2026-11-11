package app

import (
	"strings"

	"github.com/titpetric/task-ui/internal/folder"
	"github.com/titpetric/task-ui/phpfront"
)

// frontPlace is what the rows of a Jobs table run as: the folder or the
// directory inside it the reader is at, the skills that apply there and the
// terminals the reader's permissions allow.
type frontPlace struct {
	place  *folder.Folder
	skills []*folder.Skill
	shell  *folder.Job
	claude *folder.Job
}

// frontGroups groups what runs at a place: its sources, the files at the top of
// f that run, the skills, the terminals, then the sources of the directories
// inside.
func (h *FrontHost) frontGroups(f *folder.Folder, at *frontPlace, files []*folder.Job, dirs []*folder.Folder) []phpfront.Group {
	out := make([]phpfront.Group, 0, len(dirs)+len(files)+1)
	if at != nil {
		out = append(out, h.frontSources(at.place, "")...)
	}
	if len(files) > 0 {
		out = append(out, h.frontGroup(phpfront.Group{Kind: "files", Note: "that run, from the top of " + f.ID + "/"}, f, nil, files))
	}
	if at != nil {
		for _, skill := range at.skills {
			note := skill.Title
			if skill.Dir != "" {
				note = strings.TrimSpace(note + " · in " + skill.Dir + "/")
			}
			out = append(out, h.frontGroup(phpfront.Group{Kind: folder.SkillPrefix + skill.ID, Note: note}, at.place, skill, skill.Jobs))
		}
		var builtins []*folder.Job
		for _, job := range []*folder.Job{at.shell, at.claude} {
			if job != nil {
				builtins = append(builtins, job)
			}
		}
		if builtins != nil {
			out = append(out, h.frontGroup(phpfront.Group{Kind: "built-in", Note: "terminals of your own"}, at.place, nil, builtins))
		}
	}
	for _, dir := range dirs {
		rel := strings.TrimPrefix(dir.Rel, at.rel())
		if dir.ID != f.ID {
			rel = strings.TrimPrefix(dir.ID, f.ID+"/")
		}
		out = append(out, h.frontSources(dir, rel)...)
	}
	return out
}

// rel is the directory the place browses inside its folder, with its slash,
// and "" for the folder itself.
func (p *frontPlace) rel() string {
	if p == nil || p.place == nil || p.place.Rel == "" {
		return ""
	}
	return p.place.Rel + "/"
}

// frontSources is a group per job source of at, its targets in order; dir names
// a directory inside the place the targets run in.
func (h *FrontHost) frontSources(at *folder.Folder, dir string) []phpfront.Group {
	out := make([]phpfront.Group, 0, len(at.Sources))
	for _, source := range at.Sources {
		group := phpfront.Group{Kind: source.Kind, Note: source.File, Source: source.File, Dir: dir, Targets: true}
		if len(source.Command) > 0 {
			group.Note += " · " + strings.Join(source.Command, " ")
		}
		switch {
		case at.Errors != nil:
			group.Empty = "atkins lists no jobs of " + source.File + " until a fix clears its errors."
		case at.Error != "":
			group.Empty = at.Error
		default:
			group.Empty = "No targets in " + source.File + "."
		}
		out = append(out, h.frontGroup(group, at, nil, at.SourceJobs(source)))
	}
	return out
}

// frontGroup is one group of the Jobs table: the folder its rows run as, and
// each row's job with what it asks before it runs.
func (h *FrontHost) frontGroup(group phpfront.Group, at *folder.Folder, skill *folder.Skill, jobs []*folder.Job) phpfront.Group {
	if at != nil {
		group.Folder = phpfront.Where{ID: at.ID, Rel: at.Rel, Allow: at.Allow, Hours: at.Hours}
	}
	group.Jobs = make([]phpfront.Job, 0, len(jobs))
	for _, job := range jobs {
		target := frontFullJob(job)
		target.Confirm = h.frontConfirm(skill, job)
		group.Jobs = append(group.Jobs, target)
	}
	return group
}

// frontConfirm is what a job of a skill asks before it runs, or "": up and down
// of the compose project task-ui runs in stop the instance itself.
func (h *FrontHost) frontConfirm(skill *folder.Skill, job *folder.Job) string {
	if h.Compose == "" || skill == nil || job == nil || skill.ID != "compose" || skill.Project != h.Compose {
		return ""
	}
	switch job.Name {
	case "up":
		return "Run compose up for " + skill.Project + "? It recreates the containers of task-ui itself: this page and every run in it go away while it restarts."
	case "down":
		return "Run compose down for " + skill.Project + "? It stops task-ui itself, and nothing here brings it back."
	}
	return ""
}
