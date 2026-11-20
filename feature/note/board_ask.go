package note

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/integration/beszel"
	"github.com/titpetric/task-ui/internal/folder"
	"github.com/titpetric/task-ui/runner"
)

// ErrForbidden reports a prompt from someone without prompt/edit.
var ErrForbidden = errors.New("you need the prompt/edit permission to prompt Claude")

// ErrNoPrompt reports a place no prompt runs in: a directory that does not exist.
var ErrNoPrompt = errors.New("no prompt runs here")

// Ask puts a prompt to Claude in dir, or in the thread of parentID, as ask does.
func (b *Board) Ask(ctx context.Context, user *model.User, parentID, dir, body string, heard Spoken) (*model.Note, error) {
	return b.ask(ctx, user, parentID, dir, "", "", body, heard, "", "")
}

// AskAs is Ask as a named agent, from a page the origin names.
func (b *Board) AskAs(ctx context.Context, user *model.User, parentID, dir, body string, heard Spoken, agent, origin string) (*model.Note, error) {
	return b.ask(ctx, user, parentID, dir, "", "", body, heard, agent, origin)
}

// AskArea puts an area's prompt to Claude in dir, as Ask does; it takes prompt/edit.
func (b *Board) AskArea(ctx context.Context, user *model.User, area, dir, body string) (*model.Note, error) {
	return b.AskAreaAs(ctx, user, area, dir, body, "")
}

// AskAreaAs is AskArea as a named agent.
func (b *Board) AskAreaAs(ctx context.Context, user *model.User, area, dir, body, agent string) (*model.Note, error) {
	if area != model.AreaWorkflows {
		return nil, fmt.Errorf("%w: no area %q", ErrInvalid, area)
	}
	return b.ask(ctx, user, "", dir, "", area, body, Spoken{}, agent, "area:"+area)
}

// ask sends a prompt to dir, or replies in the prompt parentID names, and answers with its note.
func (b *Board) ask(ctx context.Context, user *model.User, parentID, dir, aboutRun, area, body string, heard Spoken, agent, origin string) (*model.Note, error) {
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
	note := &model.Note{Channel: model.ChannelNotes, ParentID: parentID, UserID: user.ID, UserName: user.Name, Body: body, Dir: dir, AboutRunID: aboutRun, Area: area, AudioPath: heard.Audio, Spoken: heard.Text}
	thread := note
	if parent != nil {
		note.WorkspaceID, note.AboutWorkflowRunID = parent.WorkspaceID, parent.AboutWorkflowRunID
		thread = parent
	} else {
		note.State = model.NoteOpen
		note.Origin = origin
		// A workspace the runner does not know still marks a prompt.
		if note.WorkspaceID = b.Runner.WorkspaceID(ctx, req); note.WorkspaceID == "" {
			note.WorkspaceID = model.Default
		}
		answered, err := b.resolveAgent(ctx, agent)
		if err != nil {
			return nil, err
		}
		note.Agent = answered
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
	b.dispatchPrompt(thread, note, req, nil)
	return &asked, nil
}

// reask starts the prompt of an open prompt note as the note's author.
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
	b.dispatchPrompt(note, note, req, nil)
}

// dispatchPrompt starts a thread's prompt run and fills its answer when the run ends.
func (b *Board) dispatchPrompt(thread, question *model.Note, req runner.Request, again *model.Note) {
	if question == nil {
		question = thread
	}
	if again == nil {
		again = b.answerPlaceholder(context.Background(), thread)
	}
	req, answer, cleanup, ok := b.promptEnv(thread, req, again)
	if !ok {
		return
	}
	go func() {
		defer cleanup()
		req.Env = append(req.Env, b.hosts(context.Background())...)
		b.run(context.Background(), thread, question, req, answer, "The prompt", again)
	}()
}

// answerPlaceholder drops the empty reply a crash left and returns the one a run fills.
func (b *Board) answerPlaceholder(ctx context.Context, thread *model.Note) *model.Note {
	name := b.AnswererOf(ctx, thread)
	if name == "" {
		name = answerer
	}
	if replies, err := b.Notes.Replies(ctx, []string{thread.ID}); err == nil {
		for i := range replies[thread.ID] {
			reply := &replies[thread.ID][i]
			if reply.RunID != "" && reply.State == model.NoteWorking && strings.TrimSpace(reply.Body) == "" && b.ended(ctx, reply.RunID) {
				if err := b.Notes.Remove(ctx, reply.ID); err != nil {
					slog.Warn("stale answer not removed", "note", reply.ID, "err", err)
				}
			}
		}
	}
	reply := &model.Note{
		Channel: thread.Channel, ParentID: thread.ID, UserName: name, Agent: name, RunID: model.NewID(),
		State: model.NoteWorking, WorkspaceID: thread.WorkspaceID, Dir: thread.Dir,
		AboutRunID: thread.AboutRunID, AboutWorkflowRunID: thread.AboutWorkflowRunID, Area: thread.Area,
	}
	if b.Users != nil {
		if agent, err := b.Users.GetByName(ctx, name); err == nil {
			reply.UserID = agent.ID
		}
	}
	if err := b.Notes.CreateAnswer(ctx, reply); err != nil {
		slog.Warn("answer placeholder not made", "note", thread.ID, "err", err)
		return nil
	}
	b.publish(reply)
	return reply
}

// ended reports whether a run the store holds has ended.
func (b *Board) ended(ctx context.Context, runID string) bool {
	if b.Runs == nil {
		return false
	}
	_, err := b.Runs.Get(ctx, runID)
	return err == nil && !b.live(runID)
}

// clearAnswer drops the run an answer never started.
func (b *Board) clearAnswer(ctx context.Context, reply *model.Note, state string) {
	if reply == nil {
		return
	}
	if err := b.Notes.Release(ctx, reply.ID, state); err != nil {
		slog.Warn("answer run not cleared", "note", reply.ID, "err", err)
	}
}

// discardAnswer removes the empty reply of a run that answered in the thread itself.
func (b *Board) discardAnswer(ctx context.Context, reply *model.Note) {
	if reply == nil || strings.TrimSpace(reply.Body) != "" {
		return
	}
	if err := b.Notes.Remove(ctx, reply.ID); err != nil {
		return
	}
}

// promptEnv writes the thread, about and reply files and returns the request and its cleanup.
func (b *Board) promptEnv(note *model.Note, req runner.Request, again *model.Note) (runner.Request, string, func(), bool) {
	ctx := context.Background()
	if req.ClaudeSession == "" && !req.Resume {
		req.ClaudeSession = model.NewID()
	}
	answer, err := replyFile()
	if err != nil {
		b.clearAnswer(ctx, again, model.NoteFailed)
		b.finish(ctx, note, model.NoteFailed, "The reply file was not made: "+err.Error(), again)
		return req, "", nil, false
	}
	thread := b.thread(ctx, note, again)
	req.Env = append(append([]string{}, req.Env...), "TASK_UI_PROMPT="+thread, "TASK_UI_PROMPT_REPLY="+answer)
	if agent := b.AnswererOf(ctx, note); agent != "" {
		req.Env = append(req.Env, "TASK_UI_PROMPT_AGENT="+agent)
	}
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
			b.clearAnswer(ctx, again, model.NoteFailed)
			b.finish(ctx, note, model.NoteFailed, a.what+a.id+" is not there to ask about: "+err.Error(), again)
			return req, "", nil, false
		}
		files = append(files, file)
		req.Env = append(req.Env, a.env+"="+file)
	}
	return req, answer, cleanup, true
}

// resolveAgent names the agent a new thread's runs answer as, the first for none.
func (b *Board) resolveAgent(ctx context.Context, name string) (string, error) {
	if name != "" {
		if b.Users == nil {
			return "", fmt.Errorf("%w: no agent %q", ErrInvalid, name)
		}
		user, err := b.Users.GetByName(ctx, name)
		if err != nil || !user.AgentReply || !user.Mcp {
			return "", fmt.Errorf("%w: no agent %q", ErrInvalid, name)
		}
		return user.Name, nil
	}
	if b.Users != nil {
		if agents, err := b.Users.Agents(ctx); err == nil && len(agents) > 0 {
			return agents[0].Name, nil
		}
	}
	return answerer, nil
}

// hosts returns TASK_UI_HOSTS for a prompt's run, or the line saying why it got none.
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

// promptRequest resolves the place dir names and returns the prompt job's request for it.
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
		place = &folder.Folder{ID: model.Default, Title: model.Default, Dir: folder.SrcDir(b.Root, model.Default)}
		req.Workspace = b.Root
		req.Env = []string{"TASK_UI_PROMPT_DIR=" + at}
	case dir == model.Default:
		// The root without one: Claude works in the root directory, as
		// the Notes agent does, with the server's HOME.
		place = &folder.Folder{ID: model.Default, Title: model.Default, Dir: b.Root}
		req.Env = []string{"TASK_UI_PROMPT_DIR=" + at}
	case f != nil && rest == "" && b.isWorkspace(ctx, f.ID):
		// A workspace: Claude works in its home, TASK_UI_HOME.
		at, place = folder.HomeDir(b.Root, f.ID), f
		req.Workspace = folder.WorkspaceDir(b.Root, f.ID)
	case f != nil:
		path, ok := f.Path(rest)
		if !ok {
			return req, fmt.Errorf("%w: no such directory", ErrNoPrompt)
		}
		at, place, inside = path, f, rest
		req.Env = []string{"TASK_UI_PROMPT_DIR=" + at}
	case dir != "" && !strings.Contains(dir, "/") && !strings.Contains(dir, "..") && b.isWorkspace(ctx, dir):
		// A workspace with no folder of the catalogue, its src empty.
		at = folder.HomeDir(b.Root, dir)
		place = &folder.Folder{ID: dir, Title: dir, Dir: folder.SrcDir(b.Root, dir)}
		req.Workspace = folder.WorkspaceDir(b.Root, dir)
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

// sees reports whether a user sees the place dir names.
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

// underRoot returns dir below the root as a slash path, and false for a directory outside it.
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
