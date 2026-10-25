package api

import (
	"net/http"
	"os"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// startPrompt puts a prompt to Claude in a folder, or in a directory
// inside one: a note of the folder's workspace, which the board answers
// by running @claude-prompt there as the user, the caller or the one an
// admin names with user, as promptUser picks them, with the thread in
// TASK_UI_PROMPT and the directory in TASK_UI_PROMPT_DIR. It answers with
// the note; Claude's answer is a reply to it once the run ends. With
// run, the prompt asks about that run, in its folder and directory, and
// Claude gets the run and the end of its output.
func (h *Handlers) startPrompt(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "prompt/edit") || !h.board(w) {
		return
	}
	// A run token starts no prompt: a builder would start another
	// builder.
	if access.FromRun(r.Context()) {
		fail(w, http.StatusConflict, "a run's token can't start a prompt: do the work with the task tools")
		return
	}
	var req PromptRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Run != "" && req.Parent == "" {
		h.askRun(w, r, req)
		return
	}
	if req.Area != "" && req.Parent == "" {
		h.askArea(w, r, req)
		return
	}
	where := ""
	if req.Parent == "" {
		var problem string
		if where, problem = h.promptPlace(r, req.Folder, req.Dir); problem != "" {
			fail(w, http.StatusNotFound, problem)
			return
		}
	}
	h.ask(w, r, req, where)
}

// promptPlace returns where a prompt in a folder, or in dir inside it,
// runs, as a note's Dir holds it: _ for the default workspace, the root.
// The second result says what is wrong when the place is none.
func (h *Handlers) promptPlace(r *http.Request, folderID, dir string) (string, string) {
	dir = strings.Trim(dir, "/")
	if folderID == model.Default && dir == "" {
		return model.Default, ""
	}
	f := h.visibleFolder(access.UserFrom(r.Context()), folderID)
	if f == nil {
		return "", "no such folder"
	}
	at, ok := f.Path(dir)
	if info, err := os.Stat(at); !ok || err != nil || !info.IsDir() {
		return "", "no such directory"
	}
	return model.Run{Folder: f.ID, Dir: dir}.Where(), ""
}

// ask puts the request's prompt to the board in where, as the user
// promptUser picks, or as a reply in its parent's thread, and answers
// 201 with the note.
func (h *Handlers) ask(w http.ResponseWriter, r *http.Request, req PromptRequest, where string) {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		fail(w, http.StatusBadRequest, "write a prompt")
		return
	}
	user, status, problem := h.promptUser(r, req, where)
	if problem != "" {
		fail(w, status, problem)
		return
	}
	note, err := h.opts.Board.Ask(r.Context(), user, req.Parent, where, prompt)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, note)
}

// askArea puts the request's prompt of an area's page to the board, in
// its folder and directory, the root unless given, and answers 201 with
// the note; 400 for an area there is none of.
func (h *Handlers) askArea(w http.ResponseWriter, r *http.Request, req PromptRequest) {
	if req.Area != model.AreaWorkflows {
		fail(w, http.StatusBadRequest, "no area "+req.Area)
		return
	}
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		fail(w, http.StatusBadRequest, "write a prompt")
		return
	}
	if req.Folder == "" {
		req.Folder = model.Default
	}
	where, problem := h.promptPlace(r, req.Folder, req.Dir)
	if problem != "" {
		fail(w, http.StatusNotFound, problem)
		return
	}
	user, status, problem := h.promptUser(r, req, where)
	if problem != "" {
		fail(w, status, problem)
		return
	}
	note, err := h.opts.Board.AskArea(r.Context(), user, req.Area, where, prompt)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, note)
}

// askRun puts the request's prompt about its run to the board, as the
// user promptUser picks, and answers 201 with the note; 404 for a run
// the caller does not see.
func (h *Handlers) askRun(w http.ResponseWriter, r *http.Request, req PromptRequest) {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		fail(w, http.StatusBadRequest, "write a prompt")
		return
	}
	run, err := h.opts.Runs.Get(r.Context(), req.Run)
	if err != nil || !access.UserFrom(r.Context()).CanSee(run.Folder) {
		fail(w, http.StatusNotFound, "no such run")
		return
	}
	user, status, problem := h.promptUser(r, req, run.Where())
	if problem != "" {
		fail(w, status, problem)
		return
	}
	note, err := h.opts.Board.AskRun(r.Context(), user, req.Run, prompt)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, note)
}

// listPrompts returns the prompt threads of a folder, or of ?dir= inside
// it, ?folder=_ for the root, or those about a run the user sees, ?run=,
// those of an area's page, ?area=workflows,
// or the prompt's of a workflow run the user sees, ?workflow_run=,
// newest first, each with its replies: ?limit=. Takes prompt/view.
func (h *Handlers) listPrompts(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "prompt/view") || !h.board(w) {
		return
	}
	q := r.URL.Query()
	if area := q.Get("area"); area != "" {
		user := access.UserFrom(r.Context())
		h.writeThreads(w, r, storage.NoteFilter{Prompts: true, Area: area, Limit: queryInt(r, "limit", 20)}, func(note model.Note) bool {
			return seesPlace(h.opts.Catalogue, user, note.Dir)
		})
		return
	}
	if id := q.Get("workflow_run"); id != "" {
		if _, err := h.opts.WorkflowRuns.Get(r.Context(), id); err != nil || !h.visibleRun(r, id) {
			fail(w, http.StatusNotFound, "no such workflow run")
			return
		}
		h.writeThreads(w, r, storage.NoteFilter{Prompts: true, AboutWorkflowRunID: id, Limit: queryInt(r, "limit", 20)})
		return
	}
	if id := q.Get("run"); id != "" {
		run, err := h.opts.Runs.Get(r.Context(), id)
		if err != nil || !access.UserFrom(r.Context()).CanSee(run.Folder) {
			fail(w, http.StatusNotFound, "no such run")
			return
		}
		h.writeThreads(w, r, storage.NoteFilter{Prompts: true, AboutRunID: run.ID, Limit: queryInt(r, "limit", 20)})
		return
	}
	where, problem := h.promptPlace(r, q.Get("folder"), q.Get("dir"))
	if problem != "" {
		fail(w, http.StatusNotFound, problem)
		return
	}
	h.writeThreads(w, r, storage.NoteFilter{Prompts: true, Dir: where, Limit: queryInt(r, "limit", 20)})
}

// seesPlace reports whether the user sees the place a prompt ran in:
// the folder it lies in, or the root for one who sees every folder.
func seesPlace(c folder.Catalogue, user *model.User, place string) bool {
	if place == model.Default {
		return user != nil && len(user.FolderList()) == 0
	}
	f, _ := folder.Resolve(c, place)
	return f != nil && user.CanSee(f.ID)
}
