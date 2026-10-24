package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// ProjectMoved answers a move: the project row under its new ID, the
// pins that followed it, and the rows that still name the old one.
// Those stay behind, as do the runs, under the old ID.
type ProjectMoved struct {
	*model.Project
	// Pins counts the pins, of every user, that name the new ID now:
	// the project's, and those of the folders and jobs inside it.
	Pins int               `json:"pins"`
	Left []model.FolderRef `json:"left"`
}

// moveTarget returns the ID a project takes in a workspace, its own ID
// for the workspace that holds it. A problem says why it may not move, with
// the status to answer.
func (h *Handlers) moveTarget(r *http.Request, id, ws string) (string, int, string) {
	ctx := r.Context()
	if ws == "" {
		ws = model.Default
	}
	if _, err := h.opts.Workspaces.GetByFolder(ctx, id); err == nil {
		return "", http.StatusConflict, id + " is a workspace: a workspace does not move"
	}
	from, last := model.Default, id
	if top, rest, nested := strings.Cut(id, "/"); nested {
		if _, err := h.opts.Workspaces.GetByFolder(ctx, top); err != nil || strings.Contains(rest, "/") {
			return "", http.StatusBadRequest, "only a project at the top of a workspace moves; " + id + " is a folder inside one"
		}
		from, last = top, rest
	}
	if ws == from {
		return id, 0, ""
	}
	if model.IsInternal(ws) {
		return "", http.StatusBadRequest, internalRefused
	}
	to := last
	if ws != model.Default {
		if _, err := h.opts.Workspaces.GetByFolder(ctx, ws); err != nil {
			return "", http.StatusConflict, "there is no workspace named " + ws
		}
		to = ws + "/" + last
	}
	if model.IsReserved(to) {
		return "", http.StatusBadRequest, internalRefused
	}
	if _, err := h.opts.Projects.GetByFolder(ctx, to); err == nil {
		return "", http.StatusConflict, "a project " + to + " exists already"
	}
	if ws == model.Default {
		if problem := h.rootTaken(r, to); problem != "" {
			return "", http.StatusConflict, problem
		}
	} else if _, err := os.Lstat(filepath.Join(h.opts.Root, filepath.FromSlash(to))); err == nil {
		return "", http.StatusConflict, to + " exists in the workspace already"
	}
	return to, 0, ""
}

// moveProject moves a project's checkout and row to a new ID, and the
// user folder grants and the pins with them; a home project leaves its
// workspace's home unset. The runs of the project stay out while
// it moves; one running refuses the move. A step that fails puts back
// what the steps before it did.
func (h *Handlers) moveProject(w http.ResponseWriter, r *http.Request, project *model.Project, to string) {
	ctx, from := r.Context(), project.Folder
	if h.opts.Runner != nil {
		release, err := h.opts.Runner.Freeze(from)
		if errors.Is(err, runner.ErrBusy) {
			fail(w, http.StatusConflict, from+" has a run in progress: move it after the run ends")
			return
		}
		if err != nil {
			failErr(w, err)
			return
		}
		defer release()
	}
	src := filepath.Join(h.opts.Root, filepath.FromSlash(from))
	dst := filepath.Join(h.opts.Root, filepath.FromSlash(to))
	if err := os.Rename(src, dst); err != nil {
		if errors.Is(err, syscall.EXDEV) {
			fail(w, http.StatusConflict, "the workspace is on another file system; task-ui moves a project by renaming it and can't copy it")
			return
		}
		failErr(w, err)
		return
	}
	undo := func() {
		project.Folder = from
		if err := os.Rename(dst, src); err != nil {
			slog.Error("moving a project back", "from", to, "to", from, "err", err)
		}
	}
	project.Folder = to
	if err := h.saveProject(ctx, project); err != nil {
		undo()
		failErr(w, err)
		return
	}
	if h.opts.Users != nil {
		if _, err := h.opts.Users.MoveFolder(ctx, from, to); err != nil {
			_, _ = h.opts.Users.MoveFolder(ctx, to, from)
			undo()
			_ = h.saveProject(ctx, project)
			failErr(w, err)
			return
		}
	}
	if err := h.opts.Workspaces.ClearHome(ctx, project.ID); err != nil {
		slog.Warn("unsetting a moved home project", "folder", to, "err", err)
	}
	pins := 0
	if h.opts.Pins != nil {
		n, err := h.opts.Pins.MoveFolder(ctx, from, to)
		if err != nil {
			slog.Warn("moving the pins of a moved project", "folder", from, "err", err)
		}
		pins = n
	}
	left := []model.FolderRef{}
	if h.opts.FolderRefs != nil {
		refs, err := h.opts.FolderRefs.Under(ctx, from)
		if err != nil {
			slog.Warn("listing what a move leaves", "folder", from, "err", err)
		} else {
			left = refs
		}
	}
	slog.Info("project moved", "from", from, "to", to, "pins", pins, "left", len(left))
	writeJSON(w, http.StatusOK, ProjectMoved{Project: project, Pins: pins, Left: left})
}

// saveProject writes a project row back by its ID, or makes it.
func (h *Handlers) saveProject(ctx context.Context, project *model.Project) error {
	if project.ID == "" {
		return h.opts.Projects.Create(ctx, project)
	}
	return h.opts.Projects.Update(ctx, project)
}
