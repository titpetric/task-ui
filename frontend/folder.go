package frontend

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// folderPath serves everything under /f: the folder a path names, a
// job of it, one of its env files, or a directory inside it with
// nothing to run, which opens as a page to browse. Folders nest, so the
// longest folder that starts the path is the one meant. A path starts
// with its workspace: a named one, or _ for the projects at the root;
// internal is task-ui's own. /f/_ is the default workspace itself, and a
// path that starts with none moves, for good, under /f/_/.
func (h *Handlers) folderPath(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(param(r, "*"), "/")
	named := map[string]bool{}
	for _, ws := range h.workspaceFolders(r.Context()) {
		named[ws] = true
	}
	top, rest, _ := strings.Cut(path, "/")
	switch {
	case top == model.Default && rest == "":
		h.defaultWorkspace(w, r)
		return
	case top == model.Default:
		inner, _, _ := strings.Cut(rest, "/")
		if named[inner] || model.IsReserved(inner) {
			h.missing(w, r, "No folder named "+path+".")
			return
		}
		path = rest
	case !named[top] && top != model.Internal:
		target := "/f/" + model.Default + "/" + escapePath(path)
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusMovedPermanently)
		return
	}
	h.folderAt(w, r, path)
}

// folderAt serves the path of a folder, its workspace's prefix taken off
// for one at the root.
func (h *Handlers) folderAt(w http.ResponseWriter, r *http.Request, path string) {
	f, rest := folder.Resolve(h.opts.Catalogue, path)
	if f == nil || !access.UserFrom(r.Context()).CanSee(f.ID) {
		h.missing(w, r, "No folder named "+path+".")
		return
	}
	switch {
	case rest == "":
		h.folder(w, r, f)
	case strings.HasPrefix(rest, "env/"):
		h.envFile(w, r, f, strings.TrimPrefix(rest, "env/"))
	case !strings.Contains(rest, "/") && f.Job(rest) != nil:
		h.job(w, r, f, rest)
	case isDirIn(f, rest):
		h.browse(w, r, f, rest)
	case fileJob(f, rest) != "":
		h.job(w, r, f, fileJob(f, rest))
	case dirJob(f, rest) != nil:
		h.job(w, r, dirJob(f, rest), rest[strings.LastIndex(rest, "/")+1:])
	case !strings.Contains(rest, "/"):
		h.job(w, r, f, rest)
	default:
		h.missing(w, r, "There is nothing at "+r.URL.Path+".")
	}
}

// defaultWorkspace shows the default workspace, the root: the projects
// in it the user may see, and for an admin its prompt and what the root
// holds, each project and workspace linked.
func (h *Handlers) defaultWorkspace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	f := &folder.Folder{ID: model.Default, Title: "Default", Dir: h.opts.Root}
	var workspace *model.Workspace
	if h.opts.Workspaces != nil {
		if row, err := h.opts.Workspaces.GetByFolder(ctx, model.Default); err == nil {
			workspace = row
		}
	}
	if workspace != nil {
		f.Title, f.Description = workspace.Title, workspace.Description
	}
	p := h.page(r, f.Title)
	p.Folder = f
	p.Crumbs = view.Crumbs(model.Default)
	projects := map[string]bool{}
	for _, sub := range p.Folders {
		top, _, _ := strings.Cut(sub.ID, "/")
		if _, named := p.Workspaces[top]; named || model.IsReserved(top) {
			continue
		}
		if !strings.Contains(sub.ID, "/") {
			p.SubFolders = append(p.SubFolders, sub)
			projects[sub.ID] = true
		}
	}
	if p.User.CanAdmin() {
		for _, item := range folder.Listing(h.opts.Root) {
			id := folder.SafeID(item.Name)
			if _, named := p.Workspaces[id]; item.Dir && (projects[id] || named) {
				item.Href = p.Href(id)
			}
			p.Tree = append(p.Tree, item)
		}
		if workspace != nil && h.claudeFor(r, p.User) {
			p.Prompt = &view.Prompt{API: "/api/workspaces/" + model.Default + "/prompt", Where: "the root", Open: true}
		}
	}
	h.render(w, http.StatusOK, "folder", p)
}

// isDirIn reports whether rel names a directory inside the folder that
// task-ui opens: not sealed, not reached through a symlink out of it.
func isDirIn(f *folder.Folder, rel string) bool {
	dir, ok := f.Path(rel)
	if !ok || rel == "" {
		return false
	}
	info, err := os.Stat(dir)
	return err == nil && info.IsDir()
}

// fileJob returns the ID of the job that runs the file rel names, with
// or without the ./ of its ID, or "" when the file does not run.
func fileJob(f *folder.Folder, rel string) string {
	id := folder.FilePrefix + strings.TrimPrefix(rel, folder.FilePrefix)
	if f.Job(id) == nil {
		return ""
	}
	return id
}

// dirJob returns the folder of the directory a job's path names, dir/job,
// when the job is one of that directory's, as Folder.Sub reads it, or a
// skill atkins applies there; nil otherwise.
func dirJob(f *folder.Folder, rest string) *folder.Folder {
	dir, id := path.Split(rest)
	if dir == "" || id == "" {
		return nil
	}
	sub := f.Sub(strings.TrimSuffix(dir, "/"))
	if sub == nil || sub.Job(id) == nil {
		return nil
	}
	return sub
}

// browse shows a directory inside a folder that is not a folder itself:
// the targets of its own that run there, as Folder.Sub reads them, and
// what it holds, to go on into.
func (h *Handlers) browse(w http.ResponseWriter, r *http.Request, f *folder.Folder, rel string) {
	p := h.page(r, path.Base(rel))
	p.Folder = f
	p.Browse = rel + "/"
	dir, _ := f.Path(rel)
	if sub := f.Sub(rel); sub != nil {
		p.Actions = h.actions(p, sub)
		p.Prompt = h.prompt(r, p, f, rel)
		if len(sub.Sources) > 0 {
			p.Place = sub
			p.Last = h.lastRuns(r, sub)
		}
	}
	p.Tree = h.checkoutTree(p, f, rel, folder.Listing(dir))
	p.Up = p.Href(f.ID)
	if parent := path.Dir(rel); parent != "." {
		p.Up += "/" + escapePath(parent)
	}
	p.Crumbs = crumbsOf(f.ID+"/"+rel, "", p.Workspaces)
	h.render(w, http.StatusOK, "browse", p)
}

// folder shows a folder: its jobs, its kind, what it holds, its env
// files.
func (h *Handlers) folder(w http.ResponseWriter, r *http.Request, f *folder.Folder) {
	p := h.page(r, f.Title)
	p.Folder = f
	p.Tree = h.checkoutTree(p, f, "", folder.Listing(f.Dir))
	p.Up = "/f/" + model.Default
	if parent := f.Parent(); parent != "" {
		p.Up = p.Href(parent)
	}
	p.Crumbs = crumbsOf(f.ID, "", p.Workspaces)
	p.Actions = h.actions(p, f)
	p.Prompt = h.prompt(r, p, f, "")
	if h.opts.Git != nil && h.opts.Git.IsRepo(f.Dir) {
		p.Checkout = h.opts.Git.Info(f.Dir)
		if h.opts.Disk != nil {
			if root := h.opts.Git.Checkout(f.Dir, 8); root != "" {
				usage := h.opts.Disk.Measure(root)
				p.Space = &usage
			}
		}
	}
	if h.opts.Env != nil {
		p.EnvFiles = h.opts.Env.Files(r.Context(), f)
	}
	p.Schedules = h.visibleSchedules(r, f.ID, "")
	p.Last = h.lastRuns(r, f)
	for _, sub := range p.Folders {
		if rest, ok := strings.CutPrefix(sub.ID, f.ID+"/"); ok && !strings.Contains(rest, "/") {
			p.SubFolders = append(p.SubFolders, sub)
		}
	}
	// A workspace's page shows its prompt box open.
	if _, ok := p.Workspaces[f.ID]; ok && p.Prompt != nil {
		p.Prompt.Open = true
	}
	h.render(w, http.StatusOK, "folder", p)
}

// lastRuns returns the latest run of each job of a folder.
func (h *Handlers) lastRuns(r *http.Request, f *folder.Folder) map[string]view.Run {
	runs, err := h.opts.Runs.List(r.Context(), storage.RunFilter{Folder: f.ID, Limit: 20 * len(f.Jobs)})
	if err != nil {
		return nil
	}
	last := map[string]view.Run{}
	for _, run := range runs {
		if _, seen := last[run.Key()]; !seen && run.State != model.StateRunning {
			last[run.Key()] = h.runView(run)
		}
	}
	return last
}
