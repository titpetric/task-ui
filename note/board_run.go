package note

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
)

// The most of a run's output a prompt about the run gets: its last
// lines, as many as fit both.
const (
	aboutLines = 2000
	aboutBytes = 256 << 10
)

// AskRun puts a prompt to Claude about a run, from the run's page: a
// thread in the run's directory that holds the run, which Claude gets
// with the run's details and the end of its output. Asking takes
// prompt/edit and a run the user sees, none of Claude's own; Ask with
// the thread as parent replies in it.
func (b *Board) AskRun(ctx context.Context, user *model.User, runID, body string) (*model.Note, error) {
	if b.Runs == nil {
		return nil, fmt.Errorf("%w: prompts about runs are off", ErrNoPrompt)
	}
	if user == nil {
		return nil, fmt.Errorf("%w: a prompt needs an author and some text", ErrInvalid)
	}
	run, err := b.Runs.Get(ctx, runID)
	if err != nil {
		return nil, err
	}
	if !user.CanSee(run.Folder) || run.IsClaude() {
		return nil, fmt.Errorf("%w: no such run", ErrNoPrompt)
	}
	return b.ask(ctx, user, "", run.Where(), run.ID, body)
}

// SummaryQuestion is the question of the thread Run and summarize
// asks about its run; its prompt run sets claude.RunSummary beside it.
const SummaryQuestion = "Summarize this run."

// SummarizeRun asks Claude for the summary of a run that ended, as AskRun
// does, from the user who started it: a thread about the run that
// reads SummaryQuestion, answered in the shape claude.RunSummary sets.
func (b *Board) SummarizeRun(ctx context.Context, user *model.User, runID string) (*model.Note, error) {
	return b.AskRun(ctx, user, runID, SummaryQuestion)
}

// instructions are what a prompt's run tells Claude besides the
// thread, TASK_UI_PROMPT_INSTRUCTIONS: the summary's shape for the
// question of SummarizeRun alone in its thread; nothing otherwise.
func instructions(note *model.Note, thread string) []string {
	if note.AboutRunID == "" || note.Body != SummaryQuestion || thread != note.Body {
		return nil
	}
	return []string{"TASK_UI_PROMPT_INSTRUCTIONS=" + claude.RunSummary}
}

// aboutRun writes what a prompt about a run tells Claude of it to a
// file, for TASK_UI_PROMPT_RUN: the run, the job's command and the end
// of its output. It returns the file's path.
func (b *Board) aboutRun(ctx context.Context, runID string) (string, error) {
	if b.Runs == nil {
		return "", fmt.Errorf("%w: prompts about runs are off", ErrNoPrompt)
	}
	run, err := b.Runs.Get(ctx, runID)
	if err != nil {
		return "", err
	}
	file, err := os.CreateTemp("", "note-run-*.md")
	if err != nil {
		return "", err
	}
	_, err = file.WriteString(b.runText(run))
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}

// runText describes a run for a prompt: who ran what where, how it
// ended, the job's command when the catalogue still has the job, and
// the end of the output, marked when cut.
func (b *Board) runText(run *model.Run) string {
	var text strings.Builder
	fmt.Fprintf(&text, "Run %s of the job %s in %s, by %s.\n", run.ID, run.Job, run.Where(), run.UserName)
	fmt.Fprintf(&text, "State: %s, exit code %d.\n", run.State, run.ExitCode)
	if run.Reason != "" {
		text.WriteString("Forced out of hours, because: " + run.Reason + "\n")
	}
	if run.StartedAt != nil {
		fmt.Fprintf(&text, "Started %s, took %s.\n", run.StartedAt.Format(time.RFC3339), run.Elapsed().Round(time.Second))
	}
	if args := run.ArgList(); len(args) > 0 {
		text.WriteString("Arguments: " + strings.Join(args, " ") + "\n")
	}
	if command := b.command(run); command != "" {
		text.WriteString("The job's command: " + command + "\n")
	}
	output, cut := b.outputTail(run)
	switch {
	case output == "":
		text.WriteString("\nThe run's output is not there to read.\n")
	case cut:
		fmt.Fprintf(&text, "\nThe end of the run's output, cut to its last %d lines or %d KiB:\n\n```\n%s\n```\n", aboutLines, aboutBytes>>10, output)
	default:
		text.WriteString("\nThe run's output:\n\n```\n" + output + "\n```\n")
	}
	return text.String()
}

// command is the argv of the run's job in the catalogue, empty when the
// job is not there any more.
func (b *Board) command(run *model.Run) string {
	if b.Catalogue == nil {
		return ""
	}
	f := b.Catalogue.Find(run.Folder)
	if f == nil {
		return ""
	}
	sub := f.Sub(run.Dir)
	if sub == nil {
		return ""
	}
	if job := sub.Job(run.Job); job != nil {
		return strings.Join(job.Command, " ")
	}
	return ""
}

// outputTail returns the end of a run's output as text, from its
// recording: the last aboutLines lines, fewer when those pass
// aboutBytes, and whether it left any out.
func (b *Board) outputTail(run *model.Run) (string, bool) {
	if b.Recordings == nil {
		return "", false
	}
	cast, err := b.Recordings.Read(run)
	if err != nil {
		return "", false
	}
	defer cast.Close()
	replay, err := runner.ReplayCast(cast)
	if err != nil {
		return "", false
	}
	return cutTail(strings.TrimRight(replay.Output.Text(), " \n"))
}

// cutTail keeps the last lines of an output, as many as fit aboutLines
// and aboutBytes, and reports whether it cut any.
func cutTail(output string) (string, bool) {
	if output == "" {
		return "", false
	}
	lines := strings.Split(output, "\n")
	cut := false
	if len(lines) > aboutLines {
		lines, cut = lines[len(lines)-aboutLines:], true
	}
	size := 0
	for i := len(lines) - 1; i >= 0; i-- {
		size += len(lines[i]) + 1
		if size > aboutBytes {
			lines, cut = lines[i+1:], true
			break
		}
	}
	return strings.Join(lines, "\n"), cut
}
