package agent

import (
	"context"
	"errors"
	"io"
	"os"
)

// Note answers the board thread in TASK_UI_NOTE and writes the reply to the terminal and to TASK_UI_NOTE_REPLY, which the board posts.
// It errors when the thread is empty, no token is set, or the reply file cannot be created.
func (a *Agent) Note(ctx context.Context) error {
	thread := a.getenv("TASK_UI_NOTE")
	if thread == "" {
		return errors.New("TASK_UI_NOTE: this job runs from a note on the Notes board")
	}
	if !a.signedIn() {
		return errNoToken
	}
	author := a.getenv("TASK_UI_NOTE_AUTHOR")
	if author == "" {
		author = "a user"
	}
	prompt := notePrompt(a.getenv("TASK_UI_NOTE_ID"), author, a.getenv("TASK_UI_NOTE_AGENT"), thread)

	out := a.Stdout
	if path := a.getenv("TASK_UI_NOTE_REPLY"); path != "" {
		reply, err := os.Create(path)
		if err != nil {
			return err
		}
		defer reply.Close()
		out = io.MultiWriter(a.Stdout, reply)
	}
	return a.ask(ctx, a.getenv("TASK_UI_WORKSPACE_DIR"), prompt, a.context(), out)
}

// notePrompt returns the prompt for note id: the author, the thread, and the agent this run answers as.
// A message naming several agents splits the work, so the agent has to be named.
func notePrompt(id, author, agent, thread string) string {
	you := "you"
	if agent != "" {
		you = "you, @" + agent + ","
	}
	return "An instruction on the task-ui Notes board, note " + id + " by " + author + ", for " + you + " to work on. " +
		"The thread, oldest first, each message under its author; the last message is the instruction, and your earlier replies are there too. " +
		"A message that names several users splits the work between them: do the part that is yours and leave the rest to the others.\n\n" +
		thread + "\n\n" +
		"The working directory holds the projects task-ui runs, each a git checkout of its own. Commit what you change to main and push it. Your final message is your reply on the note, in markdown, to its author: what you changed and where, or the answer they asked for. If you will not or cannot do it, do nothing and say why instead."
}
