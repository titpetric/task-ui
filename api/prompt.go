package api

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// startPrompt runs a prompt in a folder, or in a directory inside one:
// internal/claude-prompt as the user, with TASK_UI_PROMPT, the directory
// in TASK_UI_PROMPT_DIR, which Claude Code works in, and the folder's
// workspace. It answers with the run, for the window to attach to.
func (h *Handlers) startPrompt(w http.ResponseWriter, r *http.Request) {
	if !requireAdmin(w, r) {
		return
	}
	var req PromptRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	f := h.visibleFolder(access.UserFrom(r.Context()), req.Folder)
	if f == nil || model.IsInternal(f.ID) {
		fail(w, http.StatusNotFound, "no such folder")
		return
	}
	dir, ok := f.Path(strings.Trim(req.Dir, "/"))
	if info, err := os.Stat(dir); !ok || err != nil || !info.IsDir() {
		fail(w, http.StatusNotFound, "no such directory")
		return
	}
	h.prompt(w, r, req, dir, runner.Request{WorkspaceOf: f.ID, Env: []string{"TASK_UI_PROMPT_DIR=" + dir}})
}

// prompt starts internal/claude-prompt on the request's prompt in dir, an
// absolute directory under the root, as the internal job of that
// directory: one prompt runs per directory at a time, 409 with the run
// in the way for another. base carries the workspace and the environment
// of the route; TASK_UI_PROMPT comes after it.
func (h *Handlers) prompt(w http.ResponseWriter, r *http.Request, req PromptRequest, dir string, base runner.Request) {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		fail(w, http.StatusBadRequest, "write a prompt")
		return
	}
	user := access.UserFrom(r.Context())
	rel, ok := h.underRoot(dir)
	if !ok {
		fail(w, http.StatusNotFound, "no such directory")
		return
	}
	f, job, at, problem := h.targetIn(r, model.Internal, rel, promptJob)
	if problem != "" {
		fail(w, http.StatusNotFound, problem)
		return
	}
	if !canRun(user, f, job) {
		fail(w, http.StatusForbidden, "your role may view this job, but not run it")
		return
	}
	base.Folder, base.Job, base.User, base.Cols, base.Rows, base.Dir = f, job, user, req.Cols, req.Rows, at
	base.Env = append(base.Env, "TASK_UI_PROMPT="+prompt)
	session, err := h.opts.Runner.Start(r.Context(), base)
	if errors.Is(err, runner.ErrBusy) {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "a prompt is running already", "run": h.view(r, session.Snapshot())})
		return
	}
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, h.view(r, session.Snapshot()))
}

// underRoot returns where a directory is under the root, a slash path,
// "" for the root itself; false for one outside it.
func (h *Handlers) underRoot(dir string) (string, bool) {
	root, err := filepath.EvalSymlinks(h.opts.Root)
	if err != nil {
		return "", false
	}
	at, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(root, at)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	if rel == "." {
		return "", true
	}
	return filepath.ToSlash(rel), true
}
