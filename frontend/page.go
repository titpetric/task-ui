package frontend

import (
	"context"
	"errors"
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
	"github.com/titpetric/task-ui/permission"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// recentCount is how many runs the rail lists, more than a screen holds:
// the rail scrolls.
const recentCount = 50

// page returns the page model with the chrome filled in: the user, the
// folders they see, the recent runs and what is running.
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
	p.SideFolders = h.sideFolders(r.Context(), p.Folders)
	p.Workspaces = map[string]int{}
	for i, ws := range h.workspaceFolders(r.Context()) {
		p.Workspaces[ws] = i
	}
	p.Roles = h.roles(r.Context(), user)
	sees := h.runSeer(r)
	p.Running = h.running(sees)
	runs := h.visibleRuns(r.Context(), sees, storage.RunFilter{Limit: 2 * recentCount})
	// The headings read against the wall clock, as "ago" does; the page's
	// clock stays fixed for a demo or a test.
	p.Recents = view.GroupRecents(time.Now(), view.Recents(runs, h.visibleFlowRuns(r.Context(), user, recentCount), recentCount))
	p.SideWorkflows = h.visibleWorkflows(r.Context(), p)
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
		roles[u.Name] = u.Role
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
		if !hiddenUnder(hidden, f.ID) {
			side = append(side, f)
		}
	}
	return side
}

// hiddenUnder reports whether the sidebar hides a folder or one above it:
// hidden project hides its folders, in a workspace too.
func hiddenUnder(hidden map[string]bool, id string) bool {
	for {
		if hidden[id] {
			return true
		}
		i := strings.LastIndex(id, "/")
		if i < 0 {
			return false
		}
		id = id[:i]
	}
}

// running returns the live runs the user sees, by job key.
func (h *Handlers) running(sees func(*model.Run) bool) map[string]view.Run {
	result := map[string]view.Run{}
	for _, session := range h.opts.Runner.List() {
		run := session.Snapshot()
		if !sees(&run) || run.Finished() {
			continue
		}
		result[run.Key()] = h.runView(run)
	}
	return result
}

// live returns the runs with a session in memory the user sees, newest
// first, each with its frame: running ones, and the ones that just
// finished.
func (h *Handlers) live(sees func(*model.Run) bool) []view.Run {
	var result []view.Run
	for _, session := range h.opts.Runner.List() {
		run := session.Snapshot()
		if sees(&run) {
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
	if errors.Is(err, recording.ErrExpired) {
		return template.HTML(runner.TextFrame(int(run.TermCols), frameRows, recording.Expired).SVG(runner.DarkPalette)), 0, true
	}
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

// visibleRuns lists the runs of a filter the user sees.
func (h *Handlers) visibleRuns(ctx context.Context, sees func(*model.Run) bool, filter storage.RunFilter) []view.Run {
	runs, err := h.opts.Runs.List(ctx, filter)
	if err != nil {
		slog.Error("listing runs", "err", err)
		return nil
	}
	result := []view.Run{}
	for _, run := range runs {
		if sees(&run) {
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

// crumbsOf walks a nested folder ID into a breadcrumb, one step per
// segment, with a last step that has no link when given. A folder in no
// named workspace starts at the default workspace's _.
func crumbsOf(id, last string, workspaces map[string]int) []view.Crumb {
	var crumbs []view.Crumb
	if strings.HasPrefix(view.FolderHref(id, workspaces), "/f/"+model.Default+"/") {
		crumbs = append(crumbs, view.Crumb{Text: model.Default, Href: view.FolderHref(model.Default, workspaces)})
	}
	path := ""
	for _, segment := range strings.Split(id, "/") {
		if path != "" {
			path += "/"
		}
		path += segment
		crumbs = append(crumbs, view.Crumb{Text: segment, Href: view.FolderHref(path, workspaces)})
	}
	if last != "" {
		crumbs = append(crumbs, view.Crumb{Text: last})
	} else if len(crumbs) > 0 {
		crumbs[len(crumbs)-1].Href = ""
	}
	return crumbs
}
