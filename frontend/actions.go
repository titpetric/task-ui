package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// actions builds the actions bar of a place, a folder or the folder of a
// directory inside one: the git line of its checkout, its skills and its
// own atkins jobs, and for an admin Enter Shell.
func (h *Handlers) actions(p *view.Page, place *folder.Folder) *view.Actions {
	a := &view.Actions{Place: place, Self: h.opts.ComposeProject, Skills: place.Skills()}
	if p.User.CanAdmin() {
		a.Shell = place.ShellJob()
	}
	if h.opts.Git != nil && h.opts.Root != "" {
		a.Git = h.opts.Git.Status(place.Dir, h.opts.Root)
	}
	for _, job := range place.Jobs {
		if job.Kind == "atkins" {
			a.Pipeline = append(a.Pipeline, job)
		}
	}
	return a
}

// prompt returns the prompt box of a folder, or of a directory inside
// it, with the threads asked there: for an admin with a Claude token,
// none in the internal folder.
func (h *Handlers) prompt(r *http.Request, p *view.Page, f *folder.Folder, rel string) *view.Prompt {
	if !p.User.CanAdmin() || model.IsInternal(f.ID) || !h.claudeFor(r, p.User) {
		return nil
	}
	return h.withThreads(r, view.PromptIn(f.ID, rel))
}

// withThreads fills a prompt box with the last threads of its place.
func (h *Handlers) withThreads(r *http.Request, prompt *view.Prompt) *view.Prompt {
	if h.opts.Board != nil {
		prompt.Threads = h.threads(r, storage.NoteFilter{Prompts: true, Dir: prompt.Place, Limit: 20})
	}
	return prompt
}
