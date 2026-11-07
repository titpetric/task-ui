package frontend

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
	"github.com/titpetric/task-ui/storage"
)

// recentCount is how many runs the rail lists; the rail scrolls.
const recentCount = 50

// page returns the page model with the chrome filled in: the user, the
// folders they see, the recent runs and what is running.
//
// What it fills in is built beside it: the folder tree in page_sidebar.go,
// the runs and their frames in page_runs.go, and the breadcrumbs a handler
// sets itself in page_crumbs.go.
func (h *Handlers) page(r *http.Request, title string) *view.Page {
	user := access.UserFrom(r.Context())
	q := r.URL.Query()
	p := &view.Page{
		Title:         title,
		Path:          r.URL.Path,
		User:          user,
		Permissions:   h.permissions(r),
		LoginRequired: h.opts.Guard.LoginRequired(),
		Version:       h.opts.Version,
		Form:          map[string]string{},
		Notice:        q.Get("notice"),
		Error:         q.Get("error"),
		Now:           time.Now(),
		BaseURL:       h.baseURL(r),
		CanAddUsers:   h.opts.AddUsers,
	}
	if p.Notice == "" {
		p.Notice = rekeyedNotice(q.Get("rekeyed"), q.Get("what"))
	}
	if h.opts.Now != nil {
		p.Now = h.opts.Now()
	}
	for _, f := range h.opts.Catalogue.List() {
		if user.CanSee(f.ID) {
			p.Folders = append(p.Folders, f)
		}
	}
	p.Workspaces = map[string]int{}
	for i, ws := range h.workspaceFolders(r.Context()) {
		p.Workspaces[ws] = i
	}
	h.sidebar(r.Context(), p)
	p.Roles = h.roles(r.Context(), user)
	sees := h.runSeer(r)
	p.Running = h.running(sees)
	// The rail lists the runs nothing dispatched; a run under one of them
	// is a count on its row.
	runs := h.visibleRuns(r.Context(), sees, storage.RunFilter{Roots: true, Limit: 2 * recentCount})
	recents := view.Recents(runs, h.visibleFlowRuns(r.Context(), user, recentCount), recentCount)
	h.countChildren(r.Context(), sees, recents)
	// The headings read against the wall clock, as "ago" does; the page's
	// clock stays fixed for a demo or a test.
	p.Recents = view.GroupRecents(time.Now(), recents)
	p.SideWorkflows = h.visibleWorkflows(r.Context(), p)
	if user.CanAdmin() && p.Can("prompt/edit") && h.opts.Users != nil {
		p.PromptUsers, _ = h.opts.Users.List(r.Context())
		p.UserKeys = h.runAsKeys(r, p.PromptUsers)
	}
	return p
}

// baseURL returns the public base link of the instance: TASK_UI_URL, or
// the scheme and host the request came to, as the proxy in front
// forwards them.
func (h *Handlers) baseURL(r *http.Request) string {
	if h.opts.BaseURL != "" {
		return h.opts.BaseURL
	}
	scheme := "http"
	if access.IsHTTPS(r) {
		scheme = "https"
	}
	host := r.Host
	if forwarded, _, _ := strings.Cut(r.Header.Get("X-Forwarded-Host"), ","); strings.TrimSpace(forwarded) != "" {
		host = strings.TrimSpace(forwarded)
	}
	return scheme + "://" + host
}

// roles returns the role of each user by name, for a signed in user
// only: who someone is goes no further than the page needs.
func (h *Handlers) roles(ctx context.Context, user *model.User) map[string]string {
	roles := map[string]string{}
	if user == nil || h.opts.Users == nil {
		return roles
	}
	users, err := h.opts.Users.List(ctx)
	if err != nil {
		return roles
	}
	for _, u := range users {
		roles[u.Name] = u.EffectiveRole()
	}
	return roles
}

// visibleFlowRuns returns the latest runs of workflows the user may
// see, newest first: those that ran in folders the user sees only.
func (h *Handlers) visibleFlowRuns(ctx context.Context, user *model.User, limit int) []model.WorkflowRun {
	if h.opts.WorkflowRuns == nil {
		return nil
	}
	runs, err := h.opts.WorkflowRuns.List(ctx, 5*limit)
	if err != nil {
		return nil
	}
	visible := []model.WorkflowRun{}
	for _, run := range runs {
		if h.seesFlowRun(ctx, user, run.ID, 0) && len(visible) < limit {
			visible = append(visible, run)
		}
	}
	return visible
}

// seesFlowRun reports whether the user sees every folder a workflow run
// ran in, through the runs of the workflows it ran.
func (h *Handlers) seesFlowRun(ctx context.Context, user *model.User, runID string, depth int) bool {
	rows, err := h.opts.WorkflowRuns.Steps(ctx, runID)
	if err != nil || depth > 8 {
		return false
	}
	for _, row := range rows {
		if row.Folder != "" && !user.CanSee(row.Folder) {
			return false
		}
		if row.WorkflowRunRef != "" && !h.seesFlowRun(ctx, user, row.WorkflowRunRef, depth+1) {
			return false
		}
	}
	return true
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
	p.Crumbs = view.Crumbs("Not found")
	p.Error = reason
	h.render(w, http.StatusNotFound, "notfound", p)
}

// permissions returns what the user of a request may do.
func (h *Handlers) permissions(r *http.Request) permission.Set {
	user := access.UserFrom(r.Context())
	if h.opts.Users == nil {
		return permission.Set{User: user}
	}
	return h.opts.Users.Permissions(r.Context(), user)
}

// runSeer returns whether the user of a request sees a run: one of a
// folder the user sees, a prompt's with prompt/view, the Notes agent's
// with note/view, and such runs the user started.
func (h *Handlers) runSeer(r *http.Request) func(*model.Run) bool {
	user, prompts, notes := access.UserFrom(r.Context()), h.can(r, "prompt/view"), h.can(r, "note/view")
	return func(run *model.Run) bool { return user.SeesRun(run, prompts, notes) }
}

// can reports whether the user of a request may do what a path of
// `config/permissions.yml` names.
func (h *Handlers) can(r *http.Request, path string) bool {
	return h.permissions(r).Allowed(path)
}

// require renders a refusal for a user who may not do what a path
// names, and reports whether the user may.
func (h *Handlers) require(w http.ResponseWriter, r *http.Request, path string) bool {
	if h.can(r, path) {
		return true
	}
	h.refuse(w, r)
	return false
}

// refuse renders the refusal of a page the user may not see.
func (h *Handlers) refuse(w http.ResponseWriter, r *http.Request) {
	p := h.page(r, "Not allowed")
	p.Error = "You need permission to open this page."
	h.render(w, http.StatusForbidden, "forbidden", p)
}
