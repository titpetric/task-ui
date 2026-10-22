package view

import (
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/git"
)

// Actions are the actions bar of a folder page or of a directory browsed
// inside a folder: the git line of the checkout the place is in, a chip
// per skill atkins applies there, and a chip of the place's own atkins
// jobs. Every job is one of Place, run under its folder's rules.
type Actions struct {
	Place *folder.Folder
	Git   *git.Status
	// Skills are the skills with a job that applies here.
	Skills []*folder.Skill
	// Pipeline are the place's own atkins jobs.
	Pipeline []*folder.Job
	// Self is the compose project task-ui itself runs in: compose up and
	// down there ask first.
	Self string
	// Shell is the place's shell job, Enter Shell at the right of the
	// bar: set with the shell permission.
	Shell *folder.Job
	// Claude is the place's Claude job, Enter Claude beside Enter Shell:
	// set with the shell permission and a Claude token.
	Claude *folder.Job
	// WorkspaceSettings and ProjectSettings link the settings page of
	// the workspace or the project a folder page is of, beside Enter
	// Shell: set with workspace/view and project/view, on the folder's
	// own page.
	WorkspaceSettings string
	ProjectSettings   string
}

// Confirm returns what a job of a skill asks before it runs, or "": up
// and down of the compose project task-ui runs in stop the instance
// itself.
func (a *Actions) Confirm(skill *folder.Skill, job *folder.Job) string {
	if a == nil || a.Self == "" || skill == nil || job == nil || skill.ID != "compose" || skill.Project != a.Self {
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

// Empty reports whether the bar has nothing to show.
func (a *Actions) Empty() bool {
	return a == nil || (a.Git == nil && len(a.Skills) == 0 && len(a.Pipeline) == 0 && a.Shell == nil && a.Claude == nil && a.WorkspaceSettings == "" && a.ProjectSettings == "")
}
