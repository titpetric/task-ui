package view

import (
	"strings"

	"github.com/titpetric/task-ui/folder"
)

// JobGroup is a group of rows of the Jobs table of a folder's page, or
// of a directory's inside one, under a heading row: the targets of one
// job source, the files that run, the jobs of an atkins skill, the
// built-ins, or the targets of a directory inside that runs them in
// place. Each row is one job, run as one of Folder's.
type JobGroup struct {
	// Kind heads the group: the source's kind, files, the skill's ID
	// after @, built-in.
	Kind string
	// Note is what the heading says after Kind: the file and the
	// command, the skill's title.
	Note string
	// File is the job source a Show button opens: one of the folder's
	// own, empty for none.
	File string
	// Dir is the directory inside the folder a group of its targets
	// runs in, which the heading links to; empty for the place's own.
	Dir string
	// Folder is the folder the jobs run as.
	Folder *folder.Folder
	// Skill is the skill of a group of skill jobs, for what one asks
	// before it runs.
	Skill *folder.Skill
	Jobs  []*folder.Job
	// Empty is what a group without jobs says.
	Empty string
}

// NewJobGroups groups what runs in place, f or a directory of it: place's sources, f's files, a's skills and built-ins, then dirs' sources.
func NewJobGroups(f, place *folder.Folder, files []*folder.Job, a *Actions, dirs []*folder.Folder) []JobGroup {
	var groups []JobGroup
	if place != nil {
		groups = append(groups, sourceGroups(place, "")...)
	}
	if len(files) > 0 {
		groups = append(groups, JobGroup{Kind: "files", Note: "that run, from the top of " + f.ID + "/", Folder: f, Jobs: files})
	}
	if a != nil {
		for _, skill := range a.Skills {
			note := skill.Title
			if skill.Dir != "" {
				note = strings.TrimSpace(note + " · in " + skill.Dir + "/")
			}
			groups = append(groups, JobGroup{Kind: folder.SkillPrefix + skill.ID, Note: note, Folder: a.Place, Skill: skill, Jobs: skill.Jobs})
		}
		var builtins []*folder.Job
		for _, job := range []*folder.Job{a.Shell, a.Claude} {
			if job != nil {
				builtins = append(builtins, job)
			}
		}
		if builtins != nil {
			groups = append(groups, JobGroup{Kind: "built-in", Note: "terminals of your own", Folder: a.Place, Jobs: builtins})
		}
	}
	for _, dir := range dirs {
		groups = append(groups, sourceGroups(dir, strings.TrimPrefix(dir.Rel, placeRel(place)))...)
	}
	return groups
}

// placeRel is the directory place browses inside its folder, with its slash, or "" for the folder itself.
func placeRel(place *folder.Folder) string {
	if place == nil || place.Rel == "" {
		return ""
	}
	return place.Rel + "/"
}

// sourceGroups returns a group per job source of f, its targets in order; dir names a directory inside the place.
func sourceGroups(f *folder.Folder, dir string) []JobGroup {
	var groups []JobGroup
	for _, source := range f.Sources {
		g := JobGroup{Kind: source.Kind, Note: source.File, Dir: dir, Folder: f, Jobs: f.SourceJobs(source)}
		if source.Kind != "script" && f.Rel == "" {
			g.File = source.File
		}
		if len(source.Command) > 0 {
			g.Note += " · " + strings.Join(source.Command, " ")
		}
		switch {
		case f.Errors != nil:
			g.Empty = "atkins lists no jobs of " + source.File + " until a fix clears its errors."
		case f.Error != "":
			g.Empty = f.Error
		default:
			g.Empty = "No targets in " + source.File + "."
		}
		groups = append(groups, g)
	}
	return groups
}
