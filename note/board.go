// Package note is the notes board: messages users leave for each other,
// the replies to them, and the notes that send the agent an instruction.
//
// A note on the board, or a reply, that names the agent with @, such as
// @claude, is an instruction. With an agent job configured, the board starts that job as
// the agent's user, the note's text in its environment, and replies with
// how the run ended. Without one, the note stays open for an agent that
// reads the board over MCP. A note to anyone else is a message and nothing
// more.
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

// tail is how many lines of a run's output the agent's reply quotes.
const tail = 30

// Board posts notes and replies, and starts the agent on its notes.
type Board struct {
	Notes      *storage.NoteStorage
	Users      *storage.UserStorage
	Catalogue  folder.Catalogue
	Runner     *runner.Manager
	Recordings *recording.Store
	Events     *event.Bus
	// Agent is the user name a note names to send it an instruction;
	// empty turns instructions off.
	Agent string
	// AgentJob is the job the board starts for an instruction, as
	// folder/job; empty leaves instructions open for an agent over MCP.
	AgentJob string
}

// Post leaves a note on the board, or a reply to one. A note naming the
// agent opens an instruction, and so does a reply naming it: the board
// starts the agent's job on the whole thread.
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
	}
	note := &model.Note{Channel: model.ChannelNotes, ParentID: parentID, UserID: user.ID, UserName: user.Name, Body: body}
	// Only someone who may run jobs sends the agent an instruction: a
	// viewer's note to the agent is a message.
	instruction := b.Agent != "" && user.Name != b.Agent && user.Role != model.RoleViewer && contains(model.Mentions(body), b.Agent)
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
		// A reply naming the agent opens the note it answers again.
		if err := b.Notes.SetState(ctx, parent.ID, model.NoteOpen, ""); err != nil {
			return nil, err
		}
		parent.State = model.NoteOpen
		thread = parent
		b.publish(parent)
	}
	if b.AgentJob != "" {
		b.dispatch(thread)
	}
	return note, nil
}

// Comment replies to a note, and moves its instruction to the state the
// caller names: working, done or failed. An agent reports with it.
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
	var reply *model.Note
	if strings.TrimSpace(body) != "" {
		if reply, err = b.Post(ctx, user, note.ID, body); err != nil {
			return nil, err
		}
	}
	if state != "" {
		if err := b.Notes.SetState(ctx, note.ID, state, ""); err != nil {
			return nil, err
		}
		note.State = state
		b.publish(note)
	}
	if reply == nil {
		return note, nil
	}
	return reply, nil
}

// Recover fails the instructions a restart cut short: a note still
// working has no run to wait on any more. An open instruction with no run
// yet, one that never started or one set open by hand, starts the agent.
func (b *Board) Recover(ctx context.Context) error {
	if b.AgentJob != "" {
		open, err := b.Notes.List(ctx, storage.NoteFilter{State: model.NoteOpen})
		if err != nil {
			return err
		}
		for i := range open {
			if open[i].RunID == "" {
				b.dispatch(&open[i])
			}
		}
	}
	working, err := b.Notes.List(ctx, storage.NoteFilter{State: model.NoteWorking})
	if err != nil {
		return err
	}
	for _, note := range working {
		if note.RunID != "" && b.Runner != nil {
			if _, live := b.Runner.Get(note.RunID); live {
				continue
			}
		}
		b.finish(ctx, &note, model.NoteFailed, "task-ui restarted while the run was going.")
	}
	return nil
}

// dispatch starts the agent's job for an instruction, in the background,
// and replies when the run ends. A job or an agent user that is missing,
// or a run that does not start, fails the note.
func (b *Board) dispatch(note *model.Note) {
	ctx := context.Background()
	fail := func(reason string) { b.finish(ctx, note, model.NoteFailed, reason) }
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
	// The job writes its answer to the reply file; the reply on the note
	// is that answer, and the run's output stays in the run.
	answer, err := os.CreateTemp("", "note-reply-*.md")
	if err != nil {
		fail("The reply file was not made: " + err.Error())
		return
	}
	_ = answer.Close()
	req := runner.Request{
		Folder: f, Job: job, User: agent, Cols: 120, Rows: 40,
		Env: []string{
			"TASK_UI_NOTE_ID=" + note.ID, "TASK_UI_NOTE_AUTHOR=" + note.UserName, "TASK_UI_NOTE=" + b.thread(ctx, note),
			"TASK_UI_NOTE_REPLY=" + answer.Name(),
		},
	}
	go func() {
		defer os.Remove(answer.Name())
		// One run of the job at a time: an instruction waits for the
		// run in the way, then starts its own.
		session, err := b.Runner.Start(ctx, req)
		for errors.Is(err, runner.ErrBusy) && session != nil {
			<-session.Done()
			session, err = b.Runner.Start(ctx, req)
		}
		if err != nil {
			fail("The agent job did not start: " + err.Error())
			return
		}
		if err := b.Notes.SetState(ctx, note.ID, model.NoteWorking, session.ID()); err != nil {
			slog.Warn("note state not saved", "note", note.ID, "err", err)
		}
		note.State, note.RunID = model.NoteWorking, session.ID()
		b.publish(note)
		<-session.Done()
		status := session.Status()
		state := model.NoteDone
		if status.State != model.StateOK {
			state = model.NoteFailed
		}
		run := session.Snapshot()
		if reply := readReply(answer.Name()); reply != "" {
			if state == model.NoteFailed {
				reply += fmt.Sprintf("\n\nThe run ended %s (exit %d).", status.State, status.ExitCode)
			}
			b.finish(ctx, note, state, reply)
			return
		}
		b.finish(ctx, note, state, report(&run, status, b.output(&run)))
	}()
}

// thread is the text of a note and its replies, oldest first, each under
// its author: the conversation the agent works from.
func (b *Board) thread(ctx context.Context, note *model.Note) string {
	replies, _ := b.Notes.Replies(ctx, []string{note.ID})
	if len(replies[note.ID]) == 0 {
		return note.Body
	}
	var text strings.Builder
	text.WriteString(note.UserName + ":\n" + note.Body)
	for _, reply := range replies[note.ID] {
		text.WriteString("\n\n" + reply.UserName + ":\n" + reply.Body)
	}
	return text.String()
}

// finish replies to an instruction as the agent and records its state.
func (b *Board) finish(ctx context.Context, note *model.Note, state, body string) {
	reply := &model.Note{Channel: note.Channel, ParentID: note.ID, UserName: b.Agent, Body: body, RunID: note.RunID}
	if agent, err := b.Users.GetByName(ctx, b.Agent); err == nil {
		reply.UserID = agent.ID
	}
	if err := b.Notes.Create(ctx, reply); err != nil {
		slog.Warn("agent reply not saved", "note", note.ID, "err", err)
	} else {
		b.publish(reply)
	}
	if err := b.Notes.SetState(ctx, note.ID, state, ""); err != nil {
		slog.Warn("note state not saved", "note", note.ID, "err", err)
	}
	note.State = state
	b.publish(note)
}

// agentJob resolves AgentJob, folder/job, in the catalogue.
func (b *Board) agentJob() (*folder.Folder, *folder.Job) {
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

// contains reports whether names holds name.
func contains(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}
