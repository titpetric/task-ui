package view

import (
	"github.com/titpetric/task-ui/folder"
)

// JobControl returns the Jobs table of a folder's page, JobGroups laid out as the dashboard's control centre.
func (p *Page) JobControl() Control {
	return NewJobControl(p.JobGroups)
}

// RunConfirm returns what a job asks before it runs, as Actions.Confirm returns for the job's skill.
// Most jobs ask nothing.
func (p *Page) RunConfirm(job *folder.Job) string {
	if p.Actions == nil {
		return ""
	}
	for _, skill := range p.Actions.Skills {
		for _, j := range skill.Jobs {
			if j == job {
				return p.Actions.Confirm(skill, job)
			}
		}
	}
	return ""
}

// ControlMenu returns a Jobs row's [...]: per job, Run and summarize, its last run and summary, History, Run force and Kill.
func (p *Page) ControlMenu(row ControlRow) *Menu {
	m := &Menu{Label: "Actions on " + row.Label}
	f := row.Folder
	for _, job := range row.Jobs() {
		key := job.Key()
		run, last, can := p.RunningRun(key), p.LastRun(key), p.CanRun(f, job)
		what := f.ID
		if job.Dir != "" {
			what += "/" + job.Dir
		}
		what += " / " + job.Name
		start := []string{"data-run", "", "data-folder", f.ID, "data-dir", job.Dir, "data-job", job.ID, "data-title", what}
		if confirm := p.RunConfirm(job); confirm != "" {
			start = append(start, "data-confirm", confirm, "data-confirm-ok", "Run "+job.Name)
		}
		if job.TakesArgs {
			start = append(start, "data-args", job.ArgsCommand())
		}
		var items []MenuItem
		add := func(item MenuItem, _ error) {
			items = append(items, item)
		}
		if run == nil && can && p.Summarizes(job) {
			label := "Run and summarize " + job.ID
			if job.TakesArgs {
				label += "..."
			}
			add(NewMenuItem("summary", label, append([]string{"data-summarize", ""}, start...)...))
		}
		if last != nil {
			add(NewMenuItem("eye", "View the last run of "+job.ID, "data-attach", last.ID))
		}
		if summary := p.JobSummary(key); summary != nil {
			add(NewMenuItem("summary", "Summary of "+job.ID, "href", summary.Href, "title", summary.Preview))
		}
		add(NewMenuItem("list", "History of "+job.ID, "href", p.JobHref(f, job)))
		if run == nil && can && f.HoursOf(job) != "" {
			add(NewDangerItem("play", "Run force "+job.ID, append([]string{"data-force", "", "title", "Run out of hours, with a reason"}, start...)...))
		}
		if run != nil && can {
			add(NewDangerItem("stop", "Kill "+job.ID, "data-kill", run.ID))
		}
		m.Items = append(m.Items, append(items, MenuItem{Sep: true})...)
	}
	return m
}
