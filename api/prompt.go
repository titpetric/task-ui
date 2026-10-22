package api

import (
	"net/http"
	"os"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// startPrompt puts a prompt to Claude in a folder, or in a directory
// inside one: a note of the folder's workspace, which the board answers
// by running internal/claude-prompt there as the user, with the thread in
// TASK_UI_PROMPT and the directory in TASK_UI_PROMPT_DIR. It answers with
// the note; Claude's answer is a reply to it once the run ends.
func (h *Handlers) startPrompt(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "prompt/edit") || !h.board(w) {
		return
	}
	var req PromptRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
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
	if f == nil || model.IsInternal(f.ID) {
		return "", "no such folder"
	}
	at, ok := f.Path(dir)
	if info, err := os.Stat(at); !ok || err != nil || !info.IsDir() {
		return "", "no such directory"
	}
	return model.Run{Folder: f.ID, Dir: dir}.Where(), ""
}

// ask puts the request's prompt to the board in where, or as a reply in
// its parent's thread, and answers 201 with the note. One prompt runs in
// a directory at a time; another there waits for it.
func (h *Handlers) ask(w http.ResponseWriter, r *http.Request, req PromptRequest, where string) {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		fail(w, http.StatusBadRequest, "write a prompt")
		return
	}
	note, err := h.opts.Board.Ask(r.Context(), access.UserFrom(r.Context()), req.Parent, where, prompt)
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, note)
}

// listPrompts returns the prompt threads of a folder, or of ?dir= inside
// it, ?folder=_ for the root, newest first, each with its replies:
// ?limit=. Takes prompt/view.
func (h *Handlers) listPrompts(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "prompt/view") || !h.board(w) {
		return
	}
	q := r.URL.Query()
	where, problem := h.promptPlace(r, q.Get("folder"), q.Get("dir"))
	if problem != "" {
		fail(w, http.StatusNotFound, problem)
		return
	}
	h.writeThreads(w, r, storage.NoteFilter{Prompts: true, Dir: where, Limit: queryInt(r, "limit", 20)})
}
