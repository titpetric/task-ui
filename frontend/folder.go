package frontend

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
)

// folderPath serves everything under /f: the folder a path names, a
// job of it, one of its env files, or a directory inside it with
// nothing to run, which opens as a page to browse. Folders nest, so the
// longest folder that starts the path is the one meant. A path starts
// with its workspace: a named one, or _ for the projects at the root;
// internal names none, the folder of older runs only. /f/_ is the
// default workspace itself, and a path that starts with none moves, for
// good, under /f/_/.
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
	case top == model.Internal:
		h.missing(w, r, "No folder named "+path+".")
		return
	case !named[top]:
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
// for one at the root; a bare job name is one of the root's own jobs,
// which no listed folder holds.
func (h *Handlers) folderAt(w http.ResponseWriter, r *http.Request, path string) {
	f, rest := folder.Resolve(h.opts.Catalogue, path)
	if f == nil && !strings.Contains(path, "/") {
		if root := h.opts.Catalogue.Find(model.Default); root != nil && root.Job(path) != nil {
			f, rest = root, path
		}
	}
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
	case rest == "env":
		h.envFile(w, r, f, "")
	case !strings.Contains(rest, "/"):
		h.job(w, r, f, rest)
	default:
		h.missing(w, r, "There is nothing at "+r.URL.Path+".")
	}
}

// defaultWorkspace shows the default workspace, the root: the projects
// in it the user may see, with prompt/edit its prompt, and with
// project/view what the root holds, each project and workspace linked.
// Without a default workspace the page lists the root's built-in jobs
// and the workspaces' directories.
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
		f.Title, f.Description, f.Dir = workspace.Title, workspace.Description, layout.SrcDir(h.opts.Root, model.Default)
	}
	p := h.page(r, f.Title)
	p.Folder = f
	p.Crumbs = view.Crumbs(crumbName(model.Default))
	projects := map[string]bool{}
	for _, sub := range p.Folders {
		top, _, _ := strings.Cut(sub.ID, "/")
		if _, named := p.Workspaces[top]; named || layout.Reserved(top) {
			continue
		}
		if !strings.Contains(sub.ID, "/") {
			p.SubFolders = append(p.SubFolders, sub)
			projects[sub.ID] = true
		}
	}
	// The root's directories, projects or not, are what project/view
	// lists in the settings.
	if p.Can("project/view") {
		for _, item := range folder.Listing(f.Dir) {
			id := folder.SafeID(item.Name)
			if _, named := p.Workspaces[id]; item.Dir && (projects[id] || named) {
				item.Href = p.Href(id)
			}
			p.Tree = append(p.Tree, item)
		}
	}
	withTreeFolders(p, f.Dir, p.SubFolders)
	if workspace != nil {
		p.FolderMenu = &view.FolderMenu{ID: model.Default, Workspace: true}
		h.homeTree(r, p, model.Default)
	}
	if workspace != nil && p.Can("prompt/edit") && h.claudeFor(r, p.User) {
		p.Prompt = h.withThreads(r, &view.Prompt{API: "/api/workspaces/" + model.Default + "/prompt", Where: "the root", Open: true, Place: model.Default})
	}
	h.render(w, http.StatusOK, "folder", p)
}

// homeTree puts on a workspace's page its home, the $HOME of its runs,
// and what it holds as the Files panel lists it, each path under ~/. A
// home the user does not see shows none.
func (h *Handlers) homeTree(r *http.Request, p *view.Page, workspace string) {
	if h.opts.Workspaces == nil {
		return
	}
	f := h.opts.Catalogue.Find(layout.HomeID(workspace))
	if f == nil || !access.UserFrom(r.Context()).CanSee(f.ID) {
		return
	}
	p.HomeFolder = f.ID
	p.HomeTree = h.checkoutTree(p, f, "", folder.Listing(f.Dir))
	for i := range p.HomeTree {
		if p.HomeTree[i].Path != "" {
			p.HomeTree[i].Path = "~/" + p.HomeTree[i].Path
		}
	}
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
	sub := f.Sub(rel)
	if sub != nil {
		p.Actions = h.actions(r, p, sub)
		p.Prompt = h.prompt(r, p, f, rel)
		p.Summaries = h.summaries(r, p)
		if len(sub.Sources) > 0 {
			p.Place = sub
		}
	}
	p.Tree = h.checkoutTree(p, f, rel, folder.Listing(dir))
	withTreeFolders(p, dir, nil)
	p.Last = h.latestRuns(r, []*folder.Folder{f})
	h.jobTable(r, p, f, sub)
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
	p.Actions = h.actions(r, p, f)
	p.FolderMenu = folderMenu(p, f)
	if p.FolderMenu != nil && p.FolderMenu.Workspace {
		h.homeTree(r, p, f.ID)
	}
	p.Prompt = h.prompt(r, p, f, "")
	p.Summaries = h.summaries(r, p)
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
	_, workspace := p.Workspaces[f.ID]
	for _, sub := range p.Folders {
		if rest, ok := strings.CutPrefix(sub.ID, f.ID+"/"); ok && !strings.Contains(rest, "/") && !(workspace && rest == layout.Home) {
			p.SubFolders = append(p.SubFolders, sub)
		}
	}
	withTreeFolders(p, f.Dir, p.SubFolders)
	p.Last = h.latestRuns(r, []*folder.Folder{f})
	h.jobTable(r, p, f, f)
	// A workspace's page shows its prompt box open.
	if _, ok := p.Workspaces[f.ID]; ok && p.Prompt != nil {
		p.Prompt.Open = true
	}
	h.render(w, http.StatusOK, "folder", p)
}

// withTreeFolders links the listing's directories that are folders, into TreeFolders, and adds the subs it lacks as rows.
func withTreeFolders(p *view.Page, dir string, subs []*folder.Folder) {
	byDir := map[string]*folder.Folder{}
	for _, f := range p.Folders {
		byDir[filepath.Clean(f.Dir)] = f
	}
	p.TreeFolders = map[string]*folder.Folder{}
	held := map[string]bool{}
	dirs := 0
	for i := range p.Tree {
		item := &p.Tree[i]
		if !item.Dir {
			continue
		}
		dirs = i + 1
		f := byDir[filepath.Join(dir, item.Name)]
		if item.Sealed || f == nil {
			continue
		}
		item.Href = p.Href(f.ID)
		p.TreeFolders[item.Name], held[f.ID] = f, true
	}
	var missing []folder.Item
	for _, f := range subs {
		if held[f.ID] {
			continue
		}
		name := f.Name()
		item := folder.Item{Name: name, Dir: true, Href: p.Href(f.ID)}
		if info, err := os.Stat(f.Dir); err == nil {
			item.ModTime = info.ModTime()
		}
		missing = append(missing, item)
		p.TreeFolders[name] = f
	}
	p.Tree = slices.Insert(p.Tree, dirs, missing...)
}

// folderMenu returns the [...] menu of the head of a folder's page: a
// workspace's or a project's, or a folder's inside one, which pins
// alone. The template checks the permissions of its items.
func folderMenu(p *view.Page, f *folder.Folder) *view.FolderMenu {
	m := &view.FolderMenu{ID: f.ID}
	top, _, _ := strings.Cut(f.ID, "/")
	_, named := p.Workspaces[top]
	switch {
	case named && f.Depth() == 0:
		m.Workspace = true
	case (named && f.Depth() == 1) || (!named && f.Depth() == 0):
		m.Project = true
	default:
		return m
	}
	if p.Actions != nil {
		m.Pull = pullJob(p.Actions.Skills)
	}
	return m
}

// pullJob returns the pull job of the git skill, nil without one.
func pullJob(skills []*folder.Skill) *folder.Job {
	for _, skill := range skills {
		if skill.ID != "git" {
			continue
		}
		for _, job := range skill.Jobs {
			if job.Name == "pull" {
				return job
			}
		}
	}
	return nil
}
