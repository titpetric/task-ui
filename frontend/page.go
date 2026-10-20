package frontend

import (
	"context"
	"html/template"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
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
	p.SideFolders = h.sideFolders(r.Context(), p.Folders)
	p.Running = h.running(user)
	runs := h.visibleRuns(r.Context(), user, storage.RunFilter{Limit: 5 * recentCount})
	p.Recents = view.Recents(runs, h.visibleFlowRuns(r.Context(), user, recentCount), recentCount)
	p.SideWorkflows = h.visibleWorkflows(r.Context(), p)
	return p
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

// sideFolders returns the folders the sidebar lists: every folder but
// the projects set not to show there, and the folders inside those.
func (h *Handlers) sideFolders(ctx context.Context, folders []*folder.Folder) []*folder.Folder {
	hidden := map[string]bool{}
	if h.opts.Projects != nil {
		projects, _ := h.opts.Projects.List(ctx)
		for _, project := range projects {
			if project.SidebarHidden {
				hidden[project.Folder] = true
			}
		}
	}
	if len(hidden) == 0 {
		return folders
	}
	side := make([]*folder.Folder, 0, len(folders))
	for _, f := range folders {
		top, _, _ := strings.Cut(f.ID, "/")
		if !hidden[f.ID] && !hidden[top] {
			side = append(side, f)
		}
	}
	return side
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
// newest first, each with its frame: running ones, and the ones that
// just finished.
func (h *Handlers) live(user *model.User) []view.Run {
	var result []view.Run
	for _, session := range h.opts.Runner.List() {
		run := session.Snapshot()
		if user.CanSee(run.Folder) {
			result = append(result, h.framed(h.runView(run)))
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

// framed adds the rendered frame of a live run.
func (h *Handlers) framed(run view.Run) view.Run {
	if session, ok := h.opts.Runner.Get(run.ID); ok {
		run.FrameSVG = template.HTML(session.Frame().Trim(frameRows).SVG(runner.DarkPalette))
	}
	return run
}

// frameRows is the least rows a frame on a page shows.
const frameRows = 8

// frameSVG renders the last frame of a run: its session's while that
// is in memory, else the screen its recording ends on. Empty when
// neither is there.
func (h *Handlers) frameSVG(ctx context.Context, runID string) template.HTML {
	svg, _, _ := h.frameOf(ctx, runID)
	return svg
}

// frameOf renders the last frame of a run like frameSVG, with the lines
// of its whole output, and reports whether the frame shows every line:
// a finished run whose output never scrolled off the screen. A run in
// progress may still scroll; its lines are not counted (0).
func (h *Handlers) frameOf(ctx context.Context, runID string) (template.HTML, int, bool) {
	session, live := h.opts.Runner.Get(runID)
	fromSession := func() (template.HTML, int, bool) {
		if !live {
			return "", 0, false
		}
		return template.HTML(session.Frame().Trim(frameRows).SVG(runner.DarkPalette)), 0, false
	}
	if live && !session.Status().Finished() || h.opts.Recordings == nil {
		return fromSession()
	}
	run, err := h.opts.Runs.Get(ctx, runID)
	if err != nil {
		return fromSession()
	}
	cast, err := h.opts.Recordings.Read(run)
	if err != nil {
		return fromSession()
	}
	defer cast.Close()
	replay, err := runner.ReplayCast(cast)
	if err != nil {
		return fromSession()
	}
	return template.HTML(replay.Screen.Trim(frameRows).SVG(runner.DarkPalette)), replay.Lines(), replay.Complete()
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
