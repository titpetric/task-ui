// Package note is the notes board: messages users leave for each other,
// the replies to them, and the notes that hand someone an instruction.
//
// A note on the board, or a reply, that names another user with @, such
// as @claude or @consuela, from someone who may run jobs, is an
// instruction to the users it names. For the agent, the board starts the
// agent job as the agent's user, the note's text in its environment, and
// replies with how the run ended. Any other user, or the agent without an
// agent job, picks the note up over MCP: list_notes lists it open, and
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
	"time"

	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// Topic is the event bus topic the board publishes every new or changed
// note on, as "note".
const Topic = "notes"

// ErrInvalid reports a note the board does not take: an empty one, an
// unknown state, or a reply to a reply.
var ErrInvalid = errors.New("invalid note")

// ErrNotAssigned reports a state set by someone the instruction does not
// name, who is neither its author nor an admin.
var ErrNotAssigned = errors.New("only a user the note names, its author or an admin sets its state")

// answerer names Claude's replies to prompts without an agent.
const answerer = "claude"

// tail is how many lines of a run's output the agent's reply quotes.
const tail = 30

// Board posts notes and replies, and starts the agent on its notes.
type Board struct {
	Notes *storage.NoteStorage
	Users *storage.UserStorage
	// Runs reads the run a prompt asks about; nil turns those off.
	Runs       *storage.RunStorage
	Catalogue  folder.Catalogue
	Runner     *runner.Manager
	Recordings *recording.Store
	Events     *event.Bus
	// Agent is the user name a note names to have the board run the
	// agent job on it; empty leaves every instruction to MCP.
	Agent string
	// AgentJob is the job the board starts for an instruction:
	// @claude-note, the Notes agent in the root, or folder/job; empty
	// leaves instructions open for an agent over MCP.
	AgentJob string
	// AgentCommand is the command the Claude jobs of the board run, a
	// prompt's and the Notes agent's, their subcommand after it; nil runs
	// the jobs' own, task-ui agent.
	AgentCommand []string
	// Root is the directory of the projects, the default workspace, where
	// a prompt at the root and the Notes agent run.
	Root string
	// Hosts lists the inventory for a prompt's run, as beszel.Hosts does,
	// in TASK_UI_HOSTS; nil, or an error of beszel.ErrOff, leaves it out.
	// The notes agent's runs never get it.
	Hosts func(ctx context.Context) (string, error)
}

// Post leaves a note on the board, or a reply to one. A note naming
// another user opens an instruction, and so does a reply naming one; when
// it names the agent, the board starts the agent's job on the whole
// thread.
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
		if parent.ParentID != "" {
			return nil, fmt.Errorf("%w: a reply goes to a note, not to a reply", ErrInvalid)
		}
		// A reply in a prompt's thread asks again, Claude's own excepted.
		if parent.IsPrompt() && user.Name != b.Answerer() {
			return b.Ask(ctx, user, parent.ID, "", body)
		}
	}
	note := &model.Note{Channel: model.ChannelNotes, ParentID: parentID, UserID: user.ID, UserName: user.Name, Body: body}
	if parent != nil && parent.IsPrompt() {
		note.WorkspaceID, note.Dir, note.AboutRunID = parent.WorkspaceID, parent.Dir, parent.AboutRunID
		if err := b.Notes.Create(ctx, note); err != nil {
			return nil, err
		}
		b.publish(note)
		return note, nil
	}
	instruction := b.instructs(ctx, user, body)
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
	if b.forAgent(user, body) {
		b.dispatch(thread, nil)
	}
	return note, nil
}

// permissions returns what a user may do; without the users, an admin
// may do everything and anyone else nothing.
func (b *Board) permissions(ctx context.Context, user *model.User) permission.Set {
	if b.Users == nil {
		return permission.Set{User: user}
	}
	return b.Users.Permissions(ctx, user)
}

// Comment replies to a note, and moves its instruction to the state the
// caller names: working, done or failed. A user the instruction names
// reports with it; its author and an admin may set the state too, and
// anyone who may post may reply.
func (b *Board) Comment(ctx context.Context, user *model.User, id, body, state string) (*model.Note, error) {
	note, err := b.Notes.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	switch state {
	case "", model.NoteWorking, model.NoteDone, model.NoteFailed:
	default:
		return nil, fmt.Errorf("%w: no state %q", ErrInvalid, state)
	}
	if state != "" && note.State == "" {
		return nil, fmt.Errorf("%w: the note is not an instruction", ErrInvalid)
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
		if err := b.setState(ctx, note, state); err != nil {
			return nil, err
		}
		b.publish(note)
	}
	if reply == nil {
		return note, nil
	}
	return reply, nil
}

// working returns the reply of a thread a rerun rewrites, the one in
// the working state, or nil.
func (b *Board) working(ctx context.Context, note *model.Note) *model.Note {
	replies, _ := b.Notes.Replies(ctx, []string{note.ID})
	for _, reply := range replies[note.ID] {
		if reply.State == model.NoteWorking {
			return &reply
		}
	}
	return nil
}

// dispatch starts the agent's job for an instruction, in the background,
// and replies when the run ends: with a new reply, or in again, the
// agent's answer to run again, on the thread before it. A job or an agent
// user that is missing, or a run that does not start, fails the note.
func (b *Board) dispatch(note, again *model.Note) {
	ctx := context.Background()
	fail := func(reason string) { b.finish(ctx, note, model.NoteFailed, reason, again) }
	f, job := b.agentJob()
	if f == nil || job == nil {
		fail("The agent job " + b.AgentJob + " is not there.")
		return
	}
	agent, err := b.Users.GetByName(ctx, b.Agent)
	if err != nil {
		fail("There is no user " + b.Agent + " to run the agent job as.")
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
			"TASK_UI_NOTE_ID=" + note.ID, "TASK_UI_NOTE_AUTHOR=" + note.UserName, "TASK_UI_NOTE=" + b.thread(ctx, note, again),
			"TASK_UI_NOTE_REPLY=" + answer,
		},
	}
	go b.run(note, req, answer, "The agent job", again)
}

// run starts a request for a note and replies when the run ends: with
// what the job wrote to the reply file, or else with how the run ended
// and the end of its output, which stays in the run. One run of a job
// in a directory at a time: the note waits for the run in the way, then
// starts its own. An answer run again, again, takes the ID of its run
// when that run is of the same job and place, else a new one, and shows
// working until the new answer takes its place.
func (b *Board) run(note *model.Note, req runner.Request, answer, what string, again *model.Note) {
	ctx := context.Background()
	defer os.Remove(answer)
	if again != nil {
		req.ID = again.RunID
	}
	session, err := b.start(ctx, req)
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
		b.finish(ctx, note, model.NoteFailed, what+" did not start: "+err.Error(), again)
		return
	}
	if err := b.Notes.SetState(ctx, note.ID, model.NoteWorking, session.ID()); err != nil {
		slog.Warn("note state not saved", "note", note.ID, "err", err)
	}
	note.State, note.RunID = model.NoteWorking, session.ID()
	b.publish(note)
	if again != nil {
		if err := b.Notes.Answer(ctx, again.ID, again.Body, model.NoteWorking, session.ID()); err != nil {
			slog.Warn("answer state not saved", "note", again.ID, "err", err)
		}
		again.State, again.RunID = model.NoteWorking, session.ID()
		b.publish(again)
	}
	<-session.Done()
	status := session.Status()
	state := model.NoteDone
	if status.State != model.StateOK {
		state = model.NoteFailed
	}
	run := session.Snapshot()
	if reply := readReply(answer); reply != "" {
		if state == model.NoteFailed {
			reply += fmt.Sprintf("\n\nThe run ended %s (exit %d).", status.State, status.ExitCode)
		}
		b.finish(ctx, note, state, reply, again)
		return
	}
	b.finish(ctx, note, state, report(&run, status, b.output(&run)), again)
}

// start starts a request, after the run in the way of its job ends, or
// a moment later while that run is still starting and has no session;
// a run ID the request may not take turns into a new one.
func (b *Board) start(ctx context.Context, req runner.Request) (*runner.Session, error) {
	session, err := b.Runner.Start(ctx, req)
	for {
		switch {
		case errors.Is(err, runner.ErrBusy) && session != nil:
			<-session.Done()
		case errors.Is(err, runner.ErrBusy):
			time.Sleep(50 * time.Millisecond)
		case errors.Is(err, runner.ErrReuse) && req.ID != "":
			slog.Warn("run not reused", "run", req.ID, "err", err)
			req.ID = ""
		default:
			return session, err
		}
		session, err = b.Runner.Start(ctx, req)
	}
}

// replyFile makes the file a job writes its answer to; the reply on the
// note is that answer, and the run's output stays in the run.
func replyFile() (string, error) {
	answer, err := os.CreateTemp("", "note-reply-*.md")
	if err != nil {
		return "", err
	}
	return answer.Name(), answer.Close()
}

// thread is the text of a note and its replies, oldest first, each under
// its author: the conversation the agent works from. With an answer to
// run again, the thread ends before it.
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
	if len(replies) == 0 {
		return note.Body
	}
	var text strings.Builder
	text.WriteString(note.UserName + ":\n" + note.Body)
	for _, reply := range replies {
		text.WriteString("\n\n" + reply.UserName + ":\n" + reply.Body)
	}
	return text.String()
}

// finish replies to an instruction as the agent and records its state:
// in again, the reply a rerun rewrites, or else in a new reply.
func (b *Board) finish(ctx context.Context, note *model.Note, state, body string, again *model.Note) {
	if again != nil {
		b.answer(ctx, note, again, body)
	} else {
		b.reply(ctx, note, body)
	}
	if err := b.Notes.SetState(ctx, note.ID, state, ""); err != nil {
		slog.Warn("note state not saved", "note", note.ID, "err", err)
	}
	note.State = state
	b.publish(note)
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

// reply replies to a note as the agent, with the note's run.
func (b *Board) reply(ctx context.Context, note *model.Note, body string) {
	reply := &model.Note{Channel: note.Channel, ParentID: note.ID, UserName: b.Answerer(), Body: body, RunID: note.RunID, WorkspaceID: note.WorkspaceID, Dir: note.Dir, AboutRunID: note.AboutRunID}
	if agent, err := b.Users.GetByName(ctx, reply.UserName); err == nil {
		reply.UserID = agent.ID
	}
	if err := b.Notes.Create(ctx, reply); err != nil {
		slog.Warn("agent reply not saved", "note", note.ID, "err", err)
	} else {
		b.publish(reply)
	}
}

// Answerer is the user name the agent's replies carry: Agent, or claude
// for a prompt's answer with no agent set.
func (b *Board) Answerer() string {
	if b.Agent != "" {
		return b.Agent
	}
	return answerer
}

// agentJob resolves AgentJob: @claude-note, the Notes agent in the
// root, or folder/job, a job of the catalogue.
func (b *Board) agentJob() (*folder.Folder, *folder.Job) {
	if b.AgentJob == folder.NoteID && b.Root != "" {
		root := &folder.Folder{ID: model.Default, Title: model.Default, Dir: b.Root}
		return root, b.agentCommand(root.NoteJob(), "note")
	}
	cut := strings.LastIndex(b.AgentJob, "/")
	if cut <= 0 || b.Catalogue == nil {
		return nil, nil
	}
	f := b.Catalogue.Find(b.AgentJob[:cut])
	if f == nil {
		return nil, nil
	}
	return f, f.Job(b.AgentJob[cut+1:])
}

// agentCommand returns a Claude job with AgentCommand and the subcommand
// for its command, when the board has one.
func (b *Board) agentCommand(job *folder.Job, subcommand string) *folder.Job {
	if b.AgentCommand != nil {
		job.Command = append(append([]string{}, b.AgentCommand...), subcommand)
	}
	return job
}

// output returns the last lines of a run's output, from its recording.
func (b *Board) output(run *model.Run) string {
	if b.Recordings == nil {
		return ""
	}
	cast, err := b.Recordings.Read(run)
	if err != nil {
		return ""
	}
	defer cast.Close()
	replay, err := runner.ReplayCast(cast)
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimRight(replay.Output.Text(), " \n"), "\n")
	if len(lines) > tail {
		lines = lines[len(lines)-tail:]
	}
	return strings.Join(lines, "\n")
}

// publish tells the pages a note changed.
func (b *Board) publish(note *model.Note) {
	if b.Events != nil {
		b.Events.Publish(Topic, "note", *note)
	}
}

// readReply returns what the agent's job wrote to the reply file, at
// most 64 KiB.
func readReply(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	if len(data) > 64<<10 {
		data = data[:64<<10]
	}
	return strings.TrimSpace(string(data))
}

// report is the agent's reply when its run ends without an answer: how it ended, and the end
// of the output in a code block.
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
