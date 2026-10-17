package frontend

import (
	"context"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// recentCount is how many runs the sidebar lists.
const recentCount = 10

// page returns the page model with the chrome filled in: the user, the
// folders they see, the recent runs and what is running.
func (h *Handlers) page(r *http.Request, title string) *view.Page {
	user := access.UserFrom(r.Context())
	q := r.URL.Query()
	p := &view.Page{
		Title:         title,
		Path:          r.URL.Path,
		User:          user,
		LoginRequired: h.opts.Guard.LoginRequired(),
		Version:       h.opts.Version,
		Form:          map[string]string{},
		Notice:        q.Get("notice"),
		Error:         q.Get("error"),
		Now:           time.Now(),
	}
	if h.opts.Now != nil {
		p.Now = h.opts.Now()
	}
	for _, f := range h.opts.Catalogue.List() {
		if user.CanSee(f.ID) {
			p.Folders = append(p.Folders, f)
		}
	}
	p.Running = h.running(user)
	p.Recents = h.visibleRuns(r.Context(), user, storage.RunFilter{Limit: 5 * recentCount})
	if len(p.Recents) > recentCount {
		p.Recents = p.Recents[:recentCount]
	}
	return p
}

// running returns the live runs the user may see, by job key.
func (h *Handlers) running(user *model.User) map[string]view.Run {
	result := map[string]view.Run{}
	for _, session := range h.opts.Runner.List() {
		run := session.Snapshot()
		if !user.CanSee(run.Folder) || run.Finished() {
			continue
		}
		result[run.Key()] = h.runView(run)
	}
	return result
}

// live returns the runs with a session in memory the user may see,
// newest first: running ones, and the ones that just finished.
func (h *Handlers) live(user *model.User) []view.Run {
	var result []view.Run
	for _, session := range h.opts.Runner.List() {
		run := session.Snapshot()
		if user.CanSee(run.Folder) {
			result = append(result, h.runView(run))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID > result[j].ID })
	return result
}

// runView decorates a run row with what the runner knows.
func (h *Handlers) runView(run model.Run) view.Run {
	v := view.Run{Run: run}
	if session, ok := h.opts.Runner.Get(run.ID); ok {
		status := session.Status()
		v.Run, v.Live, v.Attached, v.Message = session.Snapshot(), true, session.Attached(), status.Message
	}
	return v
}

// visibleRuns lists runs the user may see.
func (h *Handlers) visibleRuns(ctx context.Context, user *model.User, filter storage.RunFilter) []view.Run {
	runs, err := h.opts.Runs.List(ctx, filter)
	if err != nil {
		slog.Error("listing runs", "err", err)
		return nil
	}
	result := []view.Run{}
	for _, run := range runs {
		if user.CanSee(run.Folder) {
			result = append(result, h.runView(run))
		}
	}
	return result
}

// render writes a page with a status.
func (h *Handlers) render(w http.ResponseWriter, status int, name string, p *view.Page) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := h.view.Render(w, name, p); err != nil {
		slog.Error("rendering page", "page", name, "err", err)
	}
}

// notFound renders the not found page.
func (h *Handlers) notFound(w http.ResponseWriter, r *http.Request) {
	h.missing(w, r, "There is nothing at "+r.URL.Path+".")
}

// missing renders the not found page with a reason.
func (h *Handlers) missing(w http.ResponseWriter, r *http.Request, reason string) {
	p := h.page(r, "Not found")
	p.Error = reason
	h.render(w, http.StatusNotFound, "notfound", p)
}

// requireAdmin renders a refusal for anyone but an admin and reports it.
func (h *Handlers) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if access.UserFrom(r.Context()).CanAdmin() {
		return true
	}
	p := h.page(r, "Admins only")
	p.Error = "Settings are for admins."
	h.render(w, http.StatusForbidden, "notfound", p)
	return false
}

// crumbsOf walks a nested folder ID into a breadcrumb, one step per
// segment, with a last step that has no link when given.
func crumbsOf(id, last string) []view.Crumb {
	var crumbs []view.Crumb
	path := ""
	for _, segment := range strings.Split(id, "/") {
		if path != "" {
			path += "/"
		}
		path += segment
		crumbs = append(crumbs, view.Crumb{Text: segment, Href: "/f/" + path})
	}
	if last != "" {
		crumbs = append(crumbs, view.Crumb{Text: last})
	} else if len(crumbs) > 0 {
		crumbs[len(crumbs)-1].Href = ""
	}
	return crumbs
}
