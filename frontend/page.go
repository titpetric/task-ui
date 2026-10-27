package frontend

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/layout"
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
	runs := h.visibleRuns(r.Context(), sees, storage.RunFilter{Limit: 2 * recentCount})
	// The headings read against the wall clock, as "ago" does; the page's
	// clock stays fixed for a demo or a test.
	p.Recents = view.GroupRecents(time.Now(), view.Recents(runs, h.visibleFlowRuns(r.Context(), user, recentCount), recentCount))
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

// sidebar fills the page's folder tree: the folders seen and the workspaces' directories, less hidden and the homes, each workspace's projects cut to its sidebar limit, then the user's pins.
func (h *Handlers) sidebar(ctx context.Context, p *view.Page) {
	hidden, homes := map[string]bool{}, map[string]bool{}
	if h.opts.Projects != nil {
		projects, _ := h.opts.Projects.List(ctx)
		for _, project := range projects {
			if project.SidebarHidden {
				hidden[project.Folder] = true
			}
		}
	}
	hidden[layout.Home], homes[layout.Home] = true, true
	for ws := range p.Workspaces {
		home := layout.HomeID(ws)
		hidden[home], homes[home] = true, true
	}
	side := view.SideDefaults(p.Folders, h.workspaceDirs(p.Workspaces), hidden, p.User.CanSee)
	side = h.sideLimits(ctx).Apply(side, p.Workspaces, h.lastCommitOf(ctx, side))
	p.Pinning = h.opts.Pins != nil && p.User != nil
	var pins []view.SidePin
	if p.Pinning {
		p.Pins, _ = h.opts.Pins.List(ctx, p.User.Name)
		for _, pin := range p.Pins {
			if side, ok := sidePin(p, pin); ok {
				pins = append(pins, side)
			}
		}
	}
	tree := view.NewSidebar(side, pins, p.Workspaces, homes)
	p.SideFolders, p.SidePins = tree.Folders, tree.Pins
}

// lastCommitOf answers the time of a project's last commit, for the
// sidebar's limits: in the directory the catalogue knows for a folder,
// else where the layout puts the project.
func (h *Handlers) lastCommitOf(ctx context.Context, folders []*folder.Folder) func(id string) time.Time {
	dirs := map[string]string{}
	for _, f := range folders {
		if f.Dir != "" {
			dirs[f.ID] = f.Dir
		}
	}
	var named []string
	return func(id string) time.Time {
		dir, ok := dirs[id]
		if !ok {
			if named == nil {
				named = h.workspaceFolders(ctx)
			}
			dir = layout.Dir(h.opts.Root, named, id)
		}
		return h.commits.Get(dir)
	}
}

// sideLimits returns how many projects the sidebar lists for each
// workspace that limits them, the root among them.
func (h *Handlers) sideLimits(ctx context.Context) view.SideLimits {
	limits := view.SideLimits{}
	if h.opts.Workspaces == nil {
		return limits
	}
	workspaces, _ := h.opts.Workspaces.List(ctx)
	for _, ws := range workspaces {
		if ws.SidebarLimit > 0 {
			limits[ws.Folder] = int(ws.SidebarLimit)
		}
	}
	return limits
}

// workspaceDirs returns the directories in the src of each named workspace by ID, none without a root.
func (h *Handlers) workspaceDirs(workspaces map[string]int) []string {
	var ids []string
	for ws := range workspaces {
		if h.opts.Root == "" {
			return nil
		}
		entries, _ := os.ReadDir(layout.SrcDir(h.opts.Root, ws))
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				ids = append(ids, ws+"/"+entry.Name())
			}
		}
	}
	return ids
}

// sidePin resolves a pin among the folders the page's user sees: a job
// of a folder, a folder, or a directory inside one; false for one that
// is not there or not seen.
func sidePin(p *view.Page, pin model.Pin) (view.SidePin, bool) {
	f, rest := folder.Resolve(seen(p.Folders), pin.Folder)
	switch {
	case f == nil:
		return view.SidePin{}, false
	case pin.Job != "":
		sub := f.Sub(pin.Dir)
		if rest != "" || sub == nil || sub.Job(pin.Job) == nil {
			return view.SidePin{}, false
		}
		return view.SidePin{Pin: pin, Folder: f, Job: sub.Job(pin.Job)}, true
	case rest != "":
		dir, ok := f.Path(rest)
		if info, err := os.Stat(dir); !ok || err != nil || !info.IsDir() {
			return view.SidePin{}, false
		}
		f = &folder.Folder{ID: pin.Folder, Title: path.Base(pin.Folder)}
	}
	return view.SidePin{Pin: pin, Folder: f}, true
}

// seen is the folders a user sees as a catalogue, for folder.Resolve.
type seen []*folder.Folder

func (s seen) List() []*folder.Folder { return s }

func (s seen) Find(id string) *folder.Folder {
	for _, f := range s {
		if f.ID == id {
			return f
		}
	}
	return nil
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
	v := view.Run{Run: run, Message: run.Message}
	if session, ok := h.opts.Runner.Get(run.ID); ok {
		status := session.Status()
		v.Run, v.Live, v.Attached = session.Snapshot(), true, session.Attached()
		if status.Message != "" {
			v.Message = status.Message
		}
		if at := h.opts.Runner.Position(run.ID); at > 0 {
			v.Message = fmt.Sprintf("waits for a free slot, number %d in line", at)
		}
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

// crumbName is a folder's name in a crumb: Default for the default
// workspace, _, as the Keys and Workspaces pages call it.
func crumbName(folder string) string {
	if folder == model.Default {
		return "Default"
	}
	return folder
}

// crumbsOf walks a nested folder ID into a breadcrumb, one step per
// segment, with a last step that has no link when given. A folder in no
// named workspace starts at the default workspace, Default.
func crumbsOf(id, last string, workspaces map[string]int) []view.Crumb {
	var crumbs []view.Crumb
	if strings.HasPrefix(view.FolderHref(id, workspaces), "/f/"+model.Default+"/") {
		crumbs = append(crumbs, view.Crumb{Text: crumbName(model.Default), Href: view.FolderHref(model.Default, workspaces)})
	}
	path := ""
	for _, segment := range strings.Split(id, "/") {
		if path != "" {
			path += "/"
		}
		path += segment
		crumbs = append(crumbs, view.Crumb{Text: crumbName(segment), Href: view.FolderHref(path, workspaces)})
	}
	if last != "" {
		crumbs = append(crumbs, view.Crumb{Text: last})
	} else if len(crumbs) > 0 {
		crumbs[len(crumbs)-1].Href = ""
	}
	return crumbs
}
