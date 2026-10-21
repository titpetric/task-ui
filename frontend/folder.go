package frontend

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
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
// longest folder that starts the path is the one meant.
func (h *Handlers) folderPath(w http.ResponseWriter, r *http.Request) {
	path := param(r, "*")
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
	case isDirIn(f.Dir, rest):
		h.browse(w, r, f, rest)
	case !strings.Contains(rest, "/"):
		h.job(w, r, f, rest)
	default:
		h.missing(w, r, "There is nothing at /f/"+path+".")
	}
}

// isDirIn reports whether rel names a directory inside dir, without
// leaving it.
func isDirIn(dir, rel string) bool {
	for _, segment := range strings.Split(rel, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel)))
	return err == nil && info.IsDir()
}

// browse shows a directory inside a folder that is not a folder itself:
// what it holds, to go on into.
func (h *Handlers) browse(w http.ResponseWriter, r *http.Request, f *folder.Folder, rel string) {
	p := h.page(r, path.Base(rel))
	p.Folder = f
	p.Browse = rel + "/"
	p.Tree = linkTree(folder.Listing(filepath.Join(f.Dir, filepath.FromSlash(rel))), "/f/"+f.ID+"/"+escapePath(rel)+"/")
	p.Crumbs = crumbsOf(f.ID+"/"+rel, "")
	h.render(w, http.StatusOK, "browse", p)
}

// folder shows a folder: its jobs, its kind, what it holds, its env
// files.
func (h *Handlers) folder(w http.ResponseWriter, r *http.Request, f *folder.Folder) {
	p := h.page(r, f.Title)
	p.Folder = f
	p.Tree = linkTree(folder.Listing(f.Dir), "/f/"+f.ID+"/")
	p.Crumbs = crumbsOf(f.ID, "")
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
	// A workspace takes a prompt from an admin with a Claude token.
	if _, ok := p.Workspaces[f.ID]; ok && p.User.CanAdmin() && h.claudeFor(r, p.User) {
		if workspace, err := h.opts.Workspaces.GetByFolder(r.Context(), f.ID); err == nil {
			p.Workspace = workspace
		}
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
