package note

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// PromptJob is the internal job a prompt runs, in the directory of the
// prompt.
const PromptJob = "claude-prompt"

// ErrForbidden reports a prompt from someone who is not an admin.
var ErrForbidden = errors.New("only an admin prompts Claude")

// ErrNoPrompt reports a place no prompt runs in: a directory or the
// internal prompt job that is not there.
var ErrNoPrompt = errors.New("no prompt runs here")

// Ask puts a prompt to Claude in dir, a folder and a directory inside it
// as Run.Where names them, or _ for the root, and answers with the note:
// open, of the folder's workspace. With parentID it replies in a
// prompt's thread instead, in the thread's place, and opens it again.
// Either way the board starts the prompt job on the thread as the user
// and replies with Claude's answer when the run ends. Only an admin
// asks.
func (b *Board) Ask(ctx context.Context, user *model.User, parentID, dir, body string) (*model.Note, error) {
	if user == nil || strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("%w: a prompt needs an author and some text", ErrInvalid)
	}
	if !user.CanAdmin() {
		return nil, ErrForbidden
	}
	var parent *model.Note
	if parentID != "" {
		var err error
		if parent, err = b.Notes.Get(ctx, parentID); err != nil {
			return nil, err
		}
		if parent.ParentID != "" || !parent.IsPrompt() {
			return nil, fmt.Errorf("%w: a reply to a prompt goes to the prompt", ErrInvalid)
		}
		dir = parent.Dir
	}
	req, err := b.promptRequest(ctx, user, dir)
	if err != nil {
		return nil, err
	}
	note := &model.Note{Channel: model.ChannelNotes, ParentID: parentID, UserID: user.ID, UserName: user.Name, Body: body, Dir: dir}
	thread := note
	if parent != nil {
		note.WorkspaceID = parent.WorkspaceID
		thread = parent
	} else {
		note.State = model.NoteOpen
		// A workspace the runner does not know still marks a prompt.
		if note.WorkspaceID = b.Runner.WorkspaceID(ctx, req); note.WorkspaceID == "" {
			note.WorkspaceID = model.Default
		}
	}
	if err := b.Notes.Create(ctx, note); err != nil {
		return nil, err
	}
	b.publish(note)
	if parent != nil {
		if err := b.Notes.SetState(ctx, parent.ID, model.NoteOpen, ""); err != nil {
			return nil, err
		}
		parent.State = model.NoteOpen
		b.publish(parent)
	}
	b.dispatchPrompt(thread, req)
	return note, nil
}

// reask starts the prompt of an open prompt note as its author.
func (b *Board) reask(ctx context.Context, note *model.Note) {
	user, err := b.Users.Get(ctx, note.UserID)
	if err != nil {
		b.finish(ctx, note, model.NoteFailed, "The user "+note.UserName+" who asked is not there any more.")
		return
	}
	req, err := b.promptRequest(ctx, user, note.Dir)
	if err != nil {
		b.finish(ctx, note, model.NoteFailed, "The prompt did not start: "+err.Error())
		return
	}
	b.dispatchPrompt(note, req)
}

// dispatchPrompt starts the prompt job on a prompt's thread, in the
// background, and replies with Claude's answer when the run ends.
func (b *Board) dispatchPrompt(note *model.Note, req runner.Request) {
	ctx := context.Background()
	answer, err := replyFile()
	if err != nil {
		b.finish(ctx, note, model.NoteFailed, "The reply file was not made: "+err.Error())
		return
	}
	req.Env = append(append([]string{}, req.Env...), "TASK_UI_PROMPT="+b.thread(ctx, note), "TASK_UI_PROMPT_REPLY="+answer)
	go func() {
		req.Env = append(req.Env, b.hosts(ctx)...)
		b.run(note, req, answer, "The prompt")
	}()
}

// hosts is the inventory a prompt's run gets, TASK_UI_HOSTS: the hosts,
// or a line saying why task-ui got none, so the prompt runs
// either way; nothing when there is no inventory.
func (b *Board) hosts(ctx context.Context) []string {
	if b.Hosts == nil {
		return nil
	}
	hosts, err := b.Hosts(ctx)
	switch {
	case errors.Is(err, beszel.ErrOff):
		return nil
	case err != nil:
		return []string{"TASK_UI_HOSTS=The inventory could not be read: " + err.Error()}
	case hosts == "":
		return nil
	}
	return []string{"TASK_UI_HOSTS=" + hosts}
}

// promptRequest resolves where a prompt runs, as the user: the internal
// prompt job of the directory dir names, a folder and a directory in it
// or a workspace's folder, or _ for the root, with the directory in
// TASK_UI_PROMPT_DIR for a folder's, and the run's workspace.
func (b *Board) promptRequest(ctx context.Context, user *model.User, dir string) (runner.Request, error) {
	var req runner.Request
	if b.Catalogue == nil || b.Runner == nil {
		return req, fmt.Errorf("%w: prompts are off", ErrNoPrompt)
	}
	at := b.Root
	switch f, rest := folder.Resolve(b.Catalogue, dir); {
	case dir == model.Default:
		req.Workspace = at
	case f != nil && !model.IsInternal(f.ID):
		path, ok := f.Path(rest)
		if !ok {
			return req, fmt.Errorf("%w: no such directory", ErrNoPrompt)
		}
		at = path
		req.WorkspaceOf, req.Env = f.ID, []string{"TASK_UI_PROMPT_DIR=" + at}
	case dir != "" && !strings.Contains(dir, "..") && !model.IsInternal(dir):
		// A workspace's checkout that is no folder of the catalogue.
		at = filepath.Join(b.Root, filepath.FromSlash(dir))
		req.Workspace = at
	default:
		return req, fmt.Errorf("%w: no such directory", ErrNoPrompt)
	}
	if info, err := os.Stat(at); err != nil || !info.IsDir() {
		return req, fmt.Errorf("%w: no such directory", ErrNoPrompt)
	}
	rel, ok := b.underRoot(at)
	internal := b.Catalogue.Find(model.Internal)
	if !ok || internal == nil {
		return req, fmt.Errorf("%w: no such directory", ErrNoPrompt)
	}
	sub := internal.Sub(rel)
	var job *folder.Job
	if sub != nil {
		job = sub.Job(PromptJob)
	}
	if job == nil {
		return req, fmt.Errorf("%w: no job %s/%s", ErrNoPrompt, model.Internal, PromptJob)
	}
	if !user.CanRun(internal.ID, internal.Allow, job.Allow) {
		return req, ErrForbidden
	}
	if job.Dir != "" {
		if req.Dir, ok = internal.Path(job.Dir); !ok {
			return req, fmt.Errorf("%w: no such directory", ErrNoPrompt)
		}
	}
	req.Folder, req.Job, req.User, req.Cols, req.Rows = internal, job, user, 120, 40
	return req, nil
}

// underRoot returns where a directory is under the root, a slash path,
// "" for the root itself; false for one outside it.
func (b *Board) underRoot(dir string) (string, bool) {
	root, err := filepath.EvalSymlinks(b.Root)
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
