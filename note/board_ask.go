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
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// ErrForbidden reports a prompt from someone without prompt/edit.
var ErrForbidden = errors.New("you need the prompt/edit permission to prompt Claude")

// ErrNoPrompt reports a place no prompt runs in: a directory that is
// not there.
var ErrNoPrompt = errors.New("no prompt runs here")

// Ask puts a prompt to Claude in dir, or in the thread of parentID, as ask does.
func (b *Board) Ask(ctx context.Context, user *model.User, parentID, dir, body string) (*model.Note, error) {
	return b.ask(ctx, user, parentID, dir, "", "", body)
}

// AskArea puts a prompt of an area's page, such as model.AreaWorkflows, to Claude in dir, as Ask does.
// The thread carries the area, so its runs get the area's instruction; asking takes prompt/edit.
func (b *Board) AskArea(ctx context.Context, user *model.User, area, dir, body string) (*model.Note, error) {
	if area != model.AreaWorkflows {
		return nil, fmt.Errorf("%w: no area %q", ErrInvalid, area)
	}
	return b.ask(ctx, user, "", dir, "", area, body)
}

// ask puts a prompt to Claude in dir, a folder and a directory inside it
// as Run.Where names them, or internal for the root, and answers with the note:
// open, of the folder's workspace, about the run aboutRun names and of
// the area when they name one. With parentID it replies in a prompt's
// thread instead, in the thread's place, keeps its run and area, and
// opens it again. Either way the board starts the prompt job on the
// thread as the user and replies with Claude's answer when the run ends.
// Asking takes prompt/edit.
func (b *Board) ask(ctx context.Context, user *model.User, parentID, dir, aboutRun, area, body string) (*model.Note, error) {
	if user == nil || strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("%w: a prompt needs an author and some text", ErrInvalid)
	}
	if !b.permissions(ctx, user).Allowed("prompt/edit") {
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
		dir, aboutRun, area = parent.Dir, parent.AboutRunID, parent.Area
	}
	req, err := b.promptRequest(ctx, user, dir, parent != nil && parent.AboutWorkflowRunID != "")
	if err != nil {
		return nil, err
	}
	note := &model.Note{Channel: model.ChannelNotes, ParentID: parentID, UserID: user.ID, UserName: user.Name, Body: body, Dir: dir, AboutRunID: aboutRun, Area: area}
	thread := note
	if parent != nil {
		note.WorkspaceID, note.AboutWorkflowRunID = parent.WorkspaceID, parent.AboutWorkflowRunID
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
	// The thread's run changes the note it holds; the caller gets it as
	// asked.
	asked := *note
	b.dispatchPrompt(thread, req, nil)
	return &asked, nil
}

// reask starts the prompt of an open prompt note as its author.
func (b *Board) reask(ctx context.Context, note *model.Note) {
	user, err := b.Users.Get(ctx, note.UserID)
	if err != nil {
		b.finish(ctx, note, model.NoteFailed, "The user "+note.UserName+" who asked is not there any more.", nil)
		return
	}
	req, err := b.promptRequest(ctx, user, note.Dir, note.AboutWorkflowRunID != "")
	if err != nil {
		b.finish(ctx, note, model.NoteFailed, "The prompt did not start: "+err.Error(), nil)
		return
	}
	b.dispatchPrompt(note, req, nil)
}

// dispatchPrompt starts the prompt job on a prompt's thread, in the
// background, and replies with Claude's answer when the run ends: in a
// new reply, or in again, the reply to rewrite, from the thread before
// it.
func (b *Board) dispatchPrompt(note *model.Note, req runner.Request, again *model.Note) {
	req, answer, cleanup, ok := b.promptEnv(note, req, again)
	if !ok {
		return
	}
	go func() {
		defer cleanup()
		req.Env = append(req.Env, b.hosts(context.Background())...)
		b.run(context.Background(), note, req, answer, "The prompt", again)
	}()
}

// promptEnv puts a prompt's thread and the file its answer goes to in
// the request's environment, and returns the request, that file, and a
// cleanup for the files it made. A thread about a run gets the run in a
// file, TASK_UI_PROMPT_RUN, as its output may pass what an environment
// variable holds, a workflow run's prompt the run's steps in
// TASK_UI_PROMPT_WORKFLOW, and the summary of Run and summarize its
// shape in TASK_UI_PROMPT_INSTRUCTIONS; a run with task-ui's MCP tools
// gets the thread it answers in TASK_UI_PROMPT_SOURCE, to reply there.
// When a file is not there to read, the note fails and promptEnv
// returns false.
func (b *Board) promptEnv(note *model.Note, req runner.Request, again *model.Note) (runner.Request, string, func(), bool) {
	ctx := context.Background()
	answer, err := replyFile()
	if err != nil {
		b.finish(ctx, note, model.NoteFailed, "The reply file was not made: "+err.Error(), again)
		return req, "", nil, false
	}
	thread := b.thread(ctx, note, again)
	req.Env = append(append([]string{}, req.Env...), "TASK_UI_PROMPT="+thread, "TASK_UI_PROMPT_REPLY="+answer)
	req.Env = append(req.Env, instructions(note, thread)...)
	var files []string
	mcp, closeMCP := b.mcpEnv(req.User, note.ID)
	req.Env = append(req.Env, mcp...)
	if len(mcp) > 0 {
		req.Env = append(req.Env, "TASK_UI_PROMPT_SOURCE="+b.source(ctx, note))
	}
	cleanup := func() {
		closeMCP()
		for _, file := range files {
			os.Remove(file)
		}
	}
	about := []struct {
		id, env, what string
		write         func(context.Context, string) (string, error)
	}{
		{note.AboutRunID, "TASK_UI_PROMPT_RUN", "The run ", b.aboutRun},
		{note.AboutWorkflowRunID, "TASK_UI_PROMPT_WORKFLOW", "The workflow run ", b.aboutWorkflowRun},
	}
	for _, a := range about {
		if a.id == "" {
			continue
		}
		file, err := a.write(ctx, a.id)
		if err != nil {
			cleanup()
			os.Remove(answer)
			b.finish(ctx, note, model.NoteFailed, a.what+a.id+" is not there to ask about: "+err.Error(), again)
			return req, "", nil, false
		}
		files = append(files, file)
		req.Env = append(req.Env, a.env+"="+file)
	}
	return req, answer, cleanup, true
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

// promptRequest resolves where a prompt runs, as the user: the prompt
// job, @claude-prompt, of the place dir names, a folder and a directory
// in it, a workspace's directory, internal for the root, or, with none,
// no workspace at the workspaces root, with the directory in
// TASK_UI_PROMPT_DIR for a folder's, and the run's workspace. The run is
// the place's, as its @shell run is, with the place's secrets. The user
// sees the place; the prompt job runs by prompt/edit, which the callers
// check, whatever the folder's allow lists.
func (b *Board) promptRequest(ctx context.Context, user *model.User, dir string, none bool) (runner.Request, error) {
	var req runner.Request
	if b.Catalogue == nil || b.Runner == nil {
		return req, fmt.Errorf("%w: prompts aren't available on this instance", ErrNoPrompt)
	}
	var place *folder.Folder
	at, inside := b.Root, ""
	switch f, rest := folder.Resolve(b.Catalogue, dir); {
	case !sees(user, f, dir):
		return req, fmt.Errorf("%w: no such directory", ErrNoPrompt)
	case dir == "" && none:
		// No workspace: Claude works at the workspaces root, over every
		// workspace, with the server's HOME and none of a workspace's
		// secrets or keys. Only a user who sees every folder asks so.
		place = &folder.Folder{ID: model.Default, Title: model.Default, Dir: b.Root}
		req.NoWorkspace = true
		req.Env = []string{"TASK_UI_PROMPT_DIR=" + at}
	case dir == model.Default && b.isWorkspace(ctx, model.Default):
		// The root with a default workspace: Claude works at the root, over
		// every workspace, with the workspace's secrets, keys and HOME.
		place = &folder.Folder{ID: model.Default, Title: model.Default, Dir: layout.SrcDir(b.Root, model.Default)}
		req.Workspace = b.Root
		req.Env = []string{"TASK_UI_PROMPT_DIR=" + at}
	case dir == model.Default:
		// The root without one: Claude works in the root directory, as
		// the Notes agent does, with the server's HOME.
		place = &folder.Folder{ID: model.Default, Title: model.Default, Dir: b.Root}
		req.Env = []string{"TASK_UI_PROMPT_DIR=" + at}
	case f != nil && rest == "" && b.isWorkspace(ctx, f.ID):
		// A workspace: Claude works in its home, TASK_UI_HOME.
		at, place = layout.HomeDir(b.Root, f.ID), f
		req.Workspace = layout.WorkspaceDir(b.Root, f.ID)
	case f != nil:
		path, ok := f.Path(rest)
		if !ok {
			return req, fmt.Errorf("%w: no such directory", ErrNoPrompt)
		}
		at, place, inside = path, f, rest
		req.Env = []string{"TASK_UI_PROMPT_DIR=" + at}
	case dir != "" && !strings.Contains(dir, "/") && !strings.Contains(dir, "..") && b.isWorkspace(ctx, dir):
		// A workspace with no folder of the catalogue, its src empty.
		at = layout.HomeDir(b.Root, dir)
		place = &folder.Folder{ID: dir, Title: dir, Dir: layout.SrcDir(b.Root, dir)}
		req.Workspace = layout.WorkspaceDir(b.Root, dir)
	default:
		return req, fmt.Errorf("%w: no such directory", ErrNoPrompt)
	}
	if info, err := os.Stat(at); err != nil || !info.IsDir() {
		return req, fmt.Errorf("%w: no such directory", ErrNoPrompt)
	}
	if _, ok := b.underRoot(at); !ok {
		return req, fmt.Errorf("%w: no such directory", ErrNoPrompt)
	}
	job := b.agentCommand(place.PromptJob(), "prompt")
	job.Dir = inside
	req.Folder, req.Job, req.Dir, req.User, req.Cols, req.Rows = place, job, at, user, 120, 40
	return req, nil
}

// sees reports whether a user sees the place of a prompt: the folder dir
// resolves to, a workspace's directory by its name, and the root for a
// user who sees every folder.
func sees(user *model.User, f *folder.Folder, dir string) bool {
	switch {
	case f != nil:
		return user.CanSee(f.ID)
	case dir == model.Default:
		return len(user.FolderList()) == 0
	}
	return user.CanSee(dir)
}

// isWorkspace reports whether a folder ID is a named workspace's.
func (b *Board) isWorkspace(ctx context.Context, id string) bool {
	if b.Workspaces == nil {
		return false
	}
	_, err := b.Workspaces.GetByFolder(ctx, id)
	return err == nil
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
