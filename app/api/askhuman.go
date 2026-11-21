package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/askhuman"
	boardnote "github.com/titpetric/task-ui/feature/note"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/runner"
)

// maxFormBody bounds the markdown one call reads, which is a note's reply.
const maxFormBody = 64 << 10

// askBlock is one ask fence of the body: what it declares, or why it was refused.
type askBlock struct {
	Form  *askhuman.Form `json:"form,omitempty"`
	Error string         `json:"error,omitempty"`
}

// listForms answers with the ask fences of a markdown body in order, each parsed or refused.
func (h *Handlers) listForms(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "prompt/view") {
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxFormBody))
	if err != nil {
		fail(w, http.StatusRequestEntityTooLarge, "the body is longer than a reply")
		return
	}
	blocks := []askBlock{}
	for _, block := range askhuman.Blocks(string(body)) {
		out := askBlock{Form: block.Form}
		if block.Err != nil {
			out.Error = block.Err.Error()
		}
		blocks = append(blocks, out)
	}
	writeJSON(w, http.StatusOK, map[string]any{"forms": blocks})
}

// AskSubmitRequest is a reader's answers to one ask fence of a note.
type AskSubmitRequest struct {
	Note     string            `json:"note"`
	Block    int               `json:"block"`
	Values   map[string]string `json:"values"`
	NoteText string            `json:"note_text"`
}

// submitAsk writes the answers under ask/ and runs the job or the agent the form names.
func (h *Handlers) submitAsk(w http.ResponseWriter, r *http.Request) {
	if !h.board(w) {
		return
	}
	if !h.can(r, "prompt/edit") && !h.can(r, "note/edit") {
		fail(w, http.StatusForbidden, "you need the prompt/edit or note/edit permission")
		return
	}
	var req AskSubmitRequest
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Note) == "" {
		fail(w, http.StatusBadRequest, "the answer names no note")
		return
	}
	threadNote, err := h.opts.Board.Notes.Get(r.Context(), req.Note)
	if err != nil {
		fail(w, http.StatusNotFound, "no such note")
		return
	}
	if !h.threadVisible(r, threadNote) {
		fail(w, http.StatusNotFound, "no such note")
		return
	}
	form, err := askhuman.BlockAt(threadNote.Body, req.Block)
	if err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	sub, err := form.Bind(req.Values)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "fields": sub.Errors})
		return
	}
	user := access.UserFrom(r.Context())
	answers := form.Answer(sub, req.Values, time.Now(), user.Name)
	answers.Note = threadNote.ID
	answers.GeneralNote = strings.TrimSpace(req.NoteText)
	jsonPath, mdPath := form.ArtifactPath(threadNote.ID, req.Block)
	if h.opts.ArtifactStore != nil {
		if err := h.writeAsk(r, jsonPath, answers, user.Name); err != nil {
			failErr(w, err)
			return
		}
		if _, err := h.opts.ArtifactStore.Write(r.Context(), mdPath, answers.Markdown(form), user.Name); err != nil {
			failErr(w, err)
			return
		}
	}
	artifacts := map[string]string{"answers": jsonPath, "decisions": mdPath}
	if form.Submit.Job != "" {
		h.submitAskJob(w, r, form, sub, threadNote, artifacts)
		return
	}
	prompt := askPrompt(form, sub, jsonPath, mdPath)
	thread, err := h.opts.Board.AskAs(r.Context(), user, threadNote.ID, "", prompt, boardnote.Spoken{}, threadNote.Agent, "ask")
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"answers": answers, "artifacts": artifacts, "note": thread})
}

// writeAsk writes answers.json as the artifact tree holds it.
func (h *Handlers) writeAsk(r *http.Request, rel string, answers *askhuman.Answers, author string) error {
	data, err := json.MarshalIndent(answers, "", "  ")
	if err != nil {
		return err
	}
	_, err = h.opts.ArtifactStore.Write(r.Context(), rel, append(data, '\n'), author)
	return err
}

// submitAskJob starts the job a form names, with the answers as its arguments.
func (h *Handlers) submitAskJob(w http.ResponseWriter, r *http.Request, form *askhuman.Form, sub *askhuman.Submission, threadNote *model.Note, artifacts map[string]string) {
	f, job, dir, problem := h.targetIn(r, form.Submit.Folder, form.Submit.Dir, form.Submit.Job)
	if problem != "" {
		fail(w, http.StatusNotFound, problem)
		return
	}
	if !h.canRun(r, f, job) {
		fail(w, http.StatusForbidden, "you can view this job, but you don't have permission to run it")
		return
	}
	user := access.UserFrom(r.Context())
	session, err := h.opts.Runner.Start(r.Context(), runner.Request{Folder: f, Job: job, User: user, Dir: dir, Args: askArgs(form, sub)})
	if err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"run": h.view(r, session.Snapshot()), "artifacts": artifacts, "note": threadNote})
}

// askArgs returns the answers a job takes as its arguments, in field order.
func askArgs(form *askhuman.Form, sub *askhuman.Submission) []string {
	args := []string{}
	for _, field := range form.Visible(sub.Values) {
		if value := sub.Values[field.Key]; value != "" {
			args = append(args, value)
		}
	}
	return args
}

// askPrompt returns what the agent is asked, naming the record the answers leave.
func askPrompt(form *askhuman.Form, sub *askhuman.Submission, jsonPath, mdPath string) string {
	prompt := strings.TrimSpace(form.Render(sub))
	if prompt == "" {
		title := form.Title
		if title == "" {
			title = "the questions"
		}
		prompt = fmt.Sprintf("The reader answered %s.", title)
	}
	return fmt.Sprintf("%s\n\nThe answers are in %s and the decision record in %s. Continue the work they decide; do not ask the same questions again.", prompt, jsonPath, mdPath)
}
