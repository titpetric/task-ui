package view

import (
	"slices"
	"strings"

	"github.com/titpetric/task-ui/folder"
)

// RunMenu is what Run and |>summarize start from a row of a Files list:
// the targets of a job source the row names, a file or a directory. The
// button runs Main; with more than that to pick from, a menu beside it
// lists the targets, default first and the rest by name.
type RunMenu struct {
	// Folder is the folder the jobs run as: for a directory that no
	// folder holds, its own from Folder.Sub.
	Folder *folder.Folder
	// Main is the job the button itself runs: the default, the only
	// target, or the file itself when it runs. Nil leaves the button to
	// open the menu.
	Main *folder.Job
	// Jobs are the targets the menu lists, in order.
	Jobs []*folder.Job
}

// NewRunMenu returns the menu of a folder's jobs, default first and the
// rest by ID, case aside. Main is the given job, else the default, else
// the only job; nil for none of those. Nil for no jobs at all.
func NewRunMenu(f *folder.Folder, jobs []*folder.Job, main *folder.Job) *RunMenu {
	if f == nil || (len(jobs) == 0 && main == nil) {
		return nil
	}
	jobs = slices.Clone(jobs)
	slices.SortStableFunc(jobs, func(a, b *folder.Job) int {
		if a.IsDefault() != b.IsDefault() {
			if a.IsDefault() {
				return -1
			}
			return 1
		}
		return strings.Compare(strings.ToLower(a.ID), strings.ToLower(b.ID))
	})
	if main == nil && len(jobs) > 0 && (jobs[0].IsDefault() || len(jobs) == 1) {
		main = jobs[0]
	}
	return &RunMenu{Folder: f, Main: main, Jobs: jobs}
}

// Menu reports whether the button has a menu: a target to pick besides the one it runs.
func (m *RunMenu) Menu() bool {
	return len(m.Jobs) > 1 || (len(m.Jobs) == 1 && m.Jobs[0] != m.Main)
}

// Describe is what the menu says of a job: its description, or that it
// runs the target when it has none or the description repeats its name
// or first step, as a job card leaves it out.
func (m *RunMenu) Describe(job *folder.Job) string {
	desc := job.Description
	if desc != "" && desc != job.Name && (len(job.Steps) == 0 || desc != job.Steps[0]) {
		return desc
	}
	return "Run the " + job.Name + " target"
}
