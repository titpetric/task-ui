// Package note is the notes board: messages users leave for each other,
// the replies to them, and the notes that hand someone an instruction.
//
// A note on the board, or a reply, that names another user with @, such
// as @claude or @consuela, from someone who may run jobs, is an
// instruction to the users it names. A reply that names nobody instructs
// the users its note names. A message naming several users instructs each
// of them, and each agent among them gets a run. For the agent, the board
// starts the agent job as the agent's user, the note's text in its
// environment, and replies with the run's outcome. Any other user, and an
// agent with no agent job, picks the note up over MCP: list_notes lists it open, and
// note_comment reports working, done or failed. A note naming no user, or
// from someone who may not run jobs, is a message and nothing more.
//
// A prompt from a folder page is a note too, with the workspace of the
// folder and the place it runs: Claude Code on the thread in that
// directory, as the user who asks, its answer the reply. The board does
// not list prompts; their folder's page does.
package note

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/feature/orb/recording"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/internal/folder"
	"github.com/titpetric/task-ui/internal/permission"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// Topic is the event bus topic the board publishes every new or changed
// note on, as "note".
const Topic = "notes"

// ErrInvalid reports a note the board does not take: an empty one, an
// unknown state, a reply to a reply.
var ErrInvalid = errors.New("invalid note")

// ErrNotAssigned reports a state set by someone the instruction does not
// name, who is neither its author nor an admin.
var ErrNotAssigned = errors.New("only a user the note names, its author or an admin can set its state")

// answerer names Claude's replies to prompts without an agent.
const answerer = "claude"

// tail is how many lines of a run's output the agent's reply quotes.
const tail = 30

// Board posts notes and replies, and starts the agent on its notes.
type Board struct {
	Notes *storage.NoteStorage
	Chats *storage.ChatStorage
	Users *storage.UserStorage
	// Runs reads the run a prompt asks about; nil turns those off.
	Runs *storage.RunStorage
	// WorkflowRuns reads the workflow run whose prompt a thread asks;
	// nil turns workflow prompts off.
	WorkflowRuns *storage.WorkflowRunStorage
	// RunSummaries keep the summaries the Runs page's Summary wrote, a
	// report of a run without a summary thread; nil leaves them out.
	RunSummaries *storage.RunSummaryStorage
	// NoteEvents keeps what Claude did in each run of a thread; nil keeps
	// no transcript and publishes no live work.
	NoteEvents *storage.NoteEventStorage
	Catalogue  folder.Catalogue
	Runner     *runner.Manager
	Recordings *recording.Store
	Events     *event.Bus
	// AgentCommand is the command the Claude jobs of the board run, a
	// prompt's and the Notes agent's, their subcommand after it; nil runs
	// the jobs' own, task-ui agent.
	AgentCommand []string
	// Root is the directory of the workspaces; a prompt at the root and
	// the Notes agent run in the default workspace's home. Workspaces
	// tells a workspace's folder from a project's, for the prompt that
	// works in a workspace's home; nil knows none.
	Root string
	// ClaudeDir holds the live Claude Code sessions, one directory a
	// prompt run; nil resumes none and keeps none.
	ClaudeDir  string
	Workspaces *storage.WorkspaceStorage
	// Hosts lists the inventory for a prompt's run, as beszel.Hosts does,
	// in TASK_UI_HOSTS; nil, or an error of beszel.ErrOff, leaves it out.
	// The notes agent's runs never get it.
	Hosts func(ctx context.Context) (string, error)
	// BaseURL is the public link of the instance, which the links a
	// prompt reads start with; empty leaves them paths.
	BaseURL string
	// RunTokens mints the token of each Claude run the board starts, for
	// the MCP server MCPURL returns; either nil, or an empty URL, leaves
	// the runs without task-ui's tools.
	RunTokens *access.RunTokens
	MCPURL    func() string

	// replied are the prompt threads whose run replied with RunReply
	// while it ran, by the thread's note ID: that run's final message is
	// then no reply.
	repliedMu sync.Mutex
	replied   map[string]bool

	// dispatched counts the runs the board has going per thread, by the
	// thread's note ID. The last to end settles the note.
	dispatchedMu sync.Mutex
	dispatched   map[string]*dispatches

	// closing leaves a run a restart killed working, with its session, for Recover.
	closing atomic.Bool
}

// Post leaves a note on the board, or a reply under one of them. A note naming
// another user opens an instruction, and so does a reply. A reply that
// names nobody addresses the users its note names. The board starts the
// agent's job on the whole thread for each agent a message addresses.
func (b *Board) Post(ctx context.Context, user *model.User, parentID, body string) (*model.Note, error) {
	if user == nil || strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("%w: a note needs an author and some text", ErrInvalid)
	}
	var parent *model.Note
	if parentID != "" {
		var err error
		if parent, err = b.Notes.Get(ctx, parentID); err != nil {
			return nil, err
		}
		if err := chatParent(parent); err != nil {
			return nil, err
		}
		if parent.ParentID != "" {
			return nil, fmt.Errorf("%w: a reply goes to a note, not to a reply", ErrInvalid)
		}
		// A reply in a prompt's thread asks again, Claude's own excepted.
		if parent.IsPrompt() && user.Name != b.Answerer() {
			return b.Ask(ctx, user, parent.ID, "", body, Spoken{})
		}
	}
	note := &model.Note{Channel: model.ChannelNotes, ParentID: parentID, UserID: user.ID, UserName: user.Name, Body: body}
	if parent != nil && parent.IsPrompt() {
		note.WorkspaceID, note.Dir, note.AboutRunID, note.AboutWorkflowRunID, note.Area = parent.WorkspaceID, parent.Dir, parent.AboutRunID, parent.AboutWorkflowRunID, parent.Area
		if err := b.Notes.Create(ctx, note); err != nil {
			return nil, err
		}
		b.publish(note)
		return note, nil
	}
	instruction := b.instructs(ctx, user, parent, body)
	if instruction && parent == nil {
		note.State = model.NoteOpen
	}
	if err := b.Notes.Create(ctx, note); err != nil {
		return nil, err
	}
	b.publish(note)
	if !instruction {
		return note, nil
	}
	thread := note
	if parent != nil {
		// A reply naming someone opens the note it answers again.
		if err := b.Notes.SetState(ctx, parent.ID, model.NoteOpen, ""); err != nil {
			return nil, err
		}
		parent.State = model.NoteOpen
		thread = parent
		b.publish(parent)
	}
	b.dispatchFor(thread, b.forAgents(ctx, user, parent, body), nil)
	return note, nil
}

// permissions returns what a user may do, an admin everything without users.
func (b *Board) permissions(ctx context.Context, user *model.User) permission.Set {
	if b.Users == nil {
		return permission.Set{User: user}
	}
	return b.Users.Permissions(ctx, user)
}

// Comment replies to a note, and moves its instruction to the state the
// caller names: working, done or failed. A user the instruction names
// reports with it; its author and an admin may set the state too, and
// anyone allowed to post may reply.
func (b *Board) Comment(ctx context.Context, user *model.User, id, body, state string) (*model.Note, error) {
	note, err := b.Notes.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := chatParent(note); err != nil {
		return nil, err
	}
	switch state {
	case "", model.NoteWorking, model.NoteDone, model.NoteFailed:
	default:
		return nil, fmt.Errorf("%w: no state %q", ErrInvalid, state)
	}
	if state != "" && note.State == "" {
		return nil, fmt.Errorf("%w: the note isn't an instruction", ErrInvalid)
	}
	if state != "" && !b.assigned(ctx, user, note) {
		return nil, ErrNotAssigned
	}
	var reply *model.Note
	if strings.TrimSpace(body) != "" {
		if reply, err = b.Post(ctx, user, note.ID, body); err != nil {
			return nil, err
		}
	}
	if state != "" {
		if err := b.setState(ctx, user, note, state); err != nil {
			return nil, err
		}
		b.publish(note)
	}
	if reply == nil {
		return note, nil
	}
	return reply, nil
}

// working returns the reply of a thread a rerun rewrites, or nil.
func (b *Board) working(ctx context.Context, note *model.Note) *model.Note {
	replies, _ := b.Notes.Replies(ctx, []string{note.ID})
	for _, reply := range replies[note.ID] {
		if reply.State == model.NoteWorking {
			return &reply
		}
	}
	return nil
}

// dispatch starts the Notes agent for an instruction and replies when the run ends.
func (b *Board) dispatch(note, again *model.Note, agent *model.User) {
	ctx := context.Background()
	own := *note
	fail := func(reason string) {
		own.RunID = ""
		b.finishAs(ctx, &own, model.NoteFailed, reason, agent.Name, again)
	}
	f, job := b.agentJob()
	if f == nil || job == nil {
		fail("The Notes agent " + folder.NoteID + " has no root to run in.")
		return
	}
	answer, err := replyFile()
	if err != nil {
		fail("The reply file was not made: " + err.Error())
		return
	}
	req := runner.Request{
		Folder: f, Job: job, User: agent, Cols: 120, Rows: 40,
		Env: []string{
			"TASK_UI_NOTE_ID=" + own.ID, "TASK_UI_NOTE_AUTHOR=" + own.UserName, "TASK_UI_NOTE=" + b.thread(ctx, &own, again),
			"TASK_UI_NOTE_AGENT=" + agent.Name, "TASK_UI_NOTE_REPLY=" + answer,
		},
	}
	mcp, closeMCP := b.mcpEnv(agent, "")
	req.Env = append(req.Env, mcp...)
	go func() {
		defer closeMCP()
		b.runAs(context.Background(), &own, &own, req, answer, "The agent job", agent.Name, again)
	}()
}

// run starts a request for a note and replies when the run ends.
func (b *Board) run(stop context.Context, note, question *model.Note, req runner.Request, answer, what string, again *model.Note) bool {
	return b.runAs(stop, note, question, req, answer, what, "", again)
}

// runAs is run, with the user a new reply is from.
func (b *Board) runAs(stop context.Context, note, question *model.Note, req runner.Request, answer, what, as string, again *model.Note) bool {
	ctx := context.Background()
	defer os.Remove(answer)
	if question == nil {
		question = note
	}
	b.takeReplied(note.ID)
	req.ParentID, req.ParentKind = parentOf(note)
	if again != nil {
		req.ID = again.RunID
	}
	tail, tailStarted := b.watch(note, question, &req, again)
	if tail != nil {
		defer os.Remove(tail.path)
	}
	session, err := b.start(stop, req)
	if tail != nil && err == nil {
		tail.runID = session.ID()
		go func() {
			ticker := time.NewTicker(transcriptEvery)
			defer ticker.Stop()
			tail.follow(ticker.C, session.Done())
			close(tailStarted)
		}()
	}
	if again != nil {
		// The reply as it stands once the run in the way ends, or a new
		// reply when someone deleted it.
		if fresh, getErr := b.Notes.Get(ctx, again.ID); getErr == nil {
			*again = *fresh
		} else {
			again = nil
		}
	}
	if err != nil {
		b.clearAnswer(ctx, again, model.NoteFailed)
		note.RunID = ""
		b.finishAs(ctx, note, model.NoteFailed, what+" did not start: "+err.Error(), as, again)
		return false
	}
	// The first run of the round owns the row, and the stream says what the
	// row says: a second agent of one note publishes no run of its own.
	held := b.claims(note.ID, session.ID())
	if err := b.Notes.SetState(ctx, note.ID, model.NoteWorking, held); err != nil {
		slog.Warn("note state not saved", "note", note.ID, "err", err)
	}
	note.State, note.RunID = model.NoteWorking, held
	b.publish(note)
	if again != nil {
		var err error
		if strings.TrimSpace(again.Body) == "" {
			err = b.Notes.SetState(ctx, again.ID, model.NoteWorking, session.ID())
		} else {
			err = b.Notes.Answer(ctx, again.ID, again.Body, model.NoteWorking, session.ID())
		}
		if err != nil {
			slog.Warn("answer state not saved", "note", again.ID, "err", err)
		}
		again.State, again.RunID = model.NoteWorking, session.ID()
		b.publish(again)
	}
	select {
	case <-session.Done():
	case <-stop.Done():
		_ = b.Runner.Stop(session.ID())
		<-session.Done()
	}
	if b.stopped(req) {
		return false
	}
	if tail != nil {
		<-tailStarted
		b.finalEvent(ctx, note, question, session.ID())
	}
	status := session.Status()
	state := model.NoteDone
	if status.State != model.StateOK {
		state = model.NoteFailed
	}
	run := session.Snapshot()
	if b.takeReplied(note.ID) {
		// Claude replied in the thread over MCP: its final message is no
		// second reply.
		b.discardAnswer(ctx, again)
		b.settle(ctx, note, state)
		return true
	}
	if reply := readReply(answer); reply != "" {
		if state == model.NoteFailed {
			reply += fmt.Sprintf("\n\nThe run ended %s (exit %d).", status.State, status.ExitCode)
		}
		b.finishAs(ctx, note, state, reply, as, again)
		return true
	}
	b.finishAs(ctx, note, state, report(&run, status, b.output(&run)), as, again)
	return false
}

// start starts a request after the run in the way of its job ends.
func (b *Board) start(stop context.Context, req runner.Request) (*runner.Session, error) {
	ctx := context.Background()
	session, err := b.Runner.Start(ctx, req)
	for {
		switch {
		case errors.Is(err, runner.ErrBusy) && session != nil:
			select {
			case <-session.Done():
			case <-stop.Done():
			}
		case errors.Is(err, runner.ErrBusy):
			time.Sleep(50 * time.Millisecond)
		case errors.Is(err, runner.ErrReuse) && req.ID != "":
			slog.Warn("run not reused", "run", req.ID, "err", err)
			req.ID = ""
		default:
			return session, err
		}
		if stop.Err() != nil {
			return nil, stop.Err()
		}
		session, err = b.Runner.Start(ctx, req)
	}
}

// thread returns the text of a note and its replies, oldest first.
func (b *Board) thread(ctx context.Context, note, again *model.Note) string {
	all, _ := b.Notes.Replies(ctx, []string{note.ID})
	replies := all[note.ID]
	if again != nil {
		for i := range replies {
			if replies[i].ID == again.ID {
				replies = replies[:i]
				break
			}
		}
	}
	var text strings.Builder
	text.WriteString(note.UserName + ":\n" + note.Body)
	empty := true
	for _, reply := range replies {
		if strings.TrimSpace(reply.Body) == "" {
			continue
		}
		empty = false
		text.WriteString("\n\n" + reply.UserName + ":\n" + reply.Body)
	}
	if empty {
		return note.Body
	}
	return text.String()
}

// finish replies to an instruction as the agent and records its state.
func (b *Board) finish(ctx context.Context, note *model.Note, state, body string, again *model.Note) {
	b.finishAs(ctx, note, state, body, "", again)
}

// finishAs is finish, with as, the user a new reply is from.
func (b *Board) finishAs(ctx context.Context, note *model.Note, state, body, as string, again *model.Note) {
	if again != nil {
		b.answer(ctx, note, again, body)
	} else {
		b.replyAs(ctx, note, body, as)
	}
	b.settle(ctx, note, state)
}

// answer puts the body in an answer run again, with the note's run.
func (b *Board) answer(ctx context.Context, note, again *model.Note, body string) {
	if err := b.Notes.Answer(ctx, again.ID, body, "", note.RunID); err != nil {
		slog.Warn("agent reply not saved", "note", note.ID, "err", err)
		return
	}
	if fresh, err := b.Notes.Get(ctx, again.ID); err == nil {
		*again = *fresh
	}
	b.publish(again)
}

// reply replies to a note as the agent and returns the reply.
func (b *Board) reply(ctx context.Context, note *model.Note, body string) *model.Note {
	return b.replyAs(ctx, note, body, "")
}

// replyAs is reply as an agent the thread names, or its answerer.
func (b *Board) replyAs(ctx context.Context, note *model.Note, body, as string) *model.Note {
	name := as
	if name == "" {
		name = b.AnswererOf(ctx, note)
	}
	if name == "" {
		name = answerer
	}
	reply := &model.Note{
		Channel: note.Channel, ParentID: note.ID, UserName: name, Body: body, RunID: note.RunID,
		WorkspaceID: note.WorkspaceID, Dir: note.Dir, AboutRunID: note.AboutRunID, AboutWorkflowRunID: note.AboutWorkflowRunID, Area: note.Area,
	}
	if agent, err := b.Users.GetByName(ctx, reply.UserName); err == nil {
		reply.UserID = agent.ID
	}
	if err := b.Notes.Create(ctx, reply); err != nil {
		slog.Warn("agent reply not saved", "note", note.ID, "err", err)
		return nil
	}
	b.publish(reply)
	return reply
}

// Answerer is the user name Claude's answers to prompts are from.
func (b *Board) Answerer() string {
	return answerer
}

// publish tells the pages a note changed.
func (b *Board) publish(note *model.Note) {
	if b.Events == nil {
		return
	}
	b.Events.Publish(Topic, "note", *note)
	root := note.ID
	if note.ParentID != "" {
		root = note.ParentID
	}
	b.Events.Publish(ThreadTopic(root), "note", *note)
}

// report is the agent's reply when its run produced no answer.
func report(run *model.Run, status runner.Status, output string) string {
	took := run.Elapsed().Round(time.Second)
	text := fmt.Sprintf("Run %s ended %s (exit %d) in %s.", run.ID, status.State, status.ExitCode, took)
	if status.Message != "" {
		text += " " + status.Message
	}
	if output != "" {
		text += "\n\n```\n" + output + "\n```"
	}
	return text
}
