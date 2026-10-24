package note

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
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
		return nil, fmt.Errorf("%w: prompts about runs aren't available on this instance", ErrNoPrompt)
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
	return b.ask(ctx, user, "", run.Where(), run.ID, "", body)
}

// SummaryQuestion is the question of the thread Run and summarize
// asks about its run; its prompt run sets claude.RunSummary beside it.
const SummaryQuestion = "Summarize this run."

// SummarizeRun asks Claude, as the run's starter, for the summary of a run that ended.
func (b *Board) SummarizeRun(ctx context.Context, user *model.User, runID string) (*model.Note, error) {
	return b.AskRun(ctx, user, runID, SummaryQuestion)
}

// BatchSummaryQuestion is the question of the thread Run and
// summarize on the inventory asks about a batch of host runs; its
// prompt run sets claude.HostsSummary beside it.
const BatchSummaryQuestion = "Summarize the runs on these hosts."

// SummarizeBatch asks Claude for the summary of a batch of host runs, a thread every run page of the batch lists.
func (b *Board) SummarizeBatch(ctx context.Context, user *model.User, runID string) (*model.Note, error) {
	return b.AskRun(ctx, user, runID, BatchSummaryQuestion)
}

// IsSummary reports whether a thread is the summary of Run and summarize.
func IsSummary(thread *model.Note) bool {
	if thread == nil || thread.ParentID != "" || thread.AboutRunID == "" {
		return false
	}
	return thread.Body == SummaryQuestion || thread.Body == BatchSummaryQuestion
}

// SummaryRunID returns the run that answers a thread's question: the
// run its first reply names, or while there is none, the thread's run.
// A reply asked later runs again; the first answer stays the summary.
func SummaryRunID(thread *model.Note, replies []model.Note) string {
	for _, reply := range replies {
		if reply.RunID != "" {
			return reply.RunID
		}
	}
	if thread == nil {
		return ""
	}
	return thread.RunID
}

// instructions are what a prompt's run tells Claude besides the
// thread, TASK_UI_PROMPT_INSTRUCTIONS: the summary's shape for the
// question of SummarizeRun, SummarizeBatch or a workflow's summary step
// alone in its thread, the triage's of claude.RunTriage for
// TriageRun's, the builder's of claude.WorkflowBuilder for every run of
// a thread of the Workflows page; nothing otherwise.
func instructions(note *model.Note, thread string) []string {
	if note.Area == model.AreaWorkflows {
		return []string{"TASK_UI_PROMPT_INSTRUCTIONS=" + claude.WorkflowBuilder}
	}
	if note.AboutWorkflowRunID != "" && note.Body == WorkflowSummaryQuestion && thread == note.Body {
		return []string{"TASK_UI_PROMPT_INSTRUCTIONS=" + claude.RunSummary}
	}
	if note.AboutRunID == "" || thread != note.Body {
		return nil
	}
	switch note.Body {
	case SummaryQuestion:
		return []string{"TASK_UI_PROMPT_INSTRUCTIONS=" + claude.RunSummary}
	case BatchSummaryQuestion:
		return []string{"TASK_UI_PROMPT_INSTRUCTIONS=" + claude.HostsSummary}
	case TriageQuestion:
		return []string{"TASK_UI_PROMPT_INSTRUCTIONS=" + claude.RunTriage}
	}
	return nil
}

// aboutRun writes what a prompt about a run tells Claude of it to a
// file, for TASK_UI_PROMPT_RUN: the run, the job's command and the end
// of its output. It returns the file's path.
func (b *Board) aboutRun(ctx context.Context, runID string) (string, error) {
	if b.Runs == nil {
		return "", fmt.Errorf("%w: prompts about runs aren't available on this instance", ErrNoPrompt)
	}
	run, err := b.Runs.Get(ctx, runID)
	if err != nil {
		return "", err
	}
	text := b.runText(run)
	if run.BatchID != "" {
		text = b.batchText(ctx, run)
	}
	file, err := os.CreateTemp("", "note-run-*.md")
	if err != nil {
		return "", err
	}
	_, err = file.WriteString(text)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}

// batchText describes the batch of a run for a prompt: every run of it,
// oldest first, as runText does one, each with the end of its output
// cut to its share of what a prompt gets.
func (b *Board) batchText(ctx context.Context, run *model.Run) string {
	runs, err := b.Runs.List(ctx, storage.RunFilter{BatchID: run.BatchID, Limit: 200})
	if err != nil || len(runs) == 0 {
		return b.runText(run)
	}
	slices.Reverse(runs)
	var text strings.Builder
	fmt.Fprintf(&text, "A batch of %d runs started together, one a host.\n", len(runs))
	for i := range runs {
		text.WriteString("\n## " + runs[i].Target + "\n\n")
		text.WriteString(b.runTextCut(&runs[i], aboutBytes/len(runs)))
	}
	return text.String()
}

// runText describes a run for a prompt: who ran what where, how it
// ended, the job's command when the catalogue still has the job, and
// the end of the output, marked when cut.
func (b *Board) runText(run *model.Run) string {
	return b.runTextCut(run, aboutBytes)
}

// runTextCut is runText with the output cut to limit bytes.
func (b *Board) runTextCut(run *model.Run, limit int) string {
	var text strings.Builder
	fmt.Fprintf(&text, "Run %s of the job %s in %s, by %s.\n", run.ID, run.Job, run.Where(), run.UserName)
	if run.Target != "" {
		text.WriteString("On the host " + run.Target + ".\n")
	}
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
	output, cut := cutTail(b.outputText(run), limit)
	switch {
	case output == "":
		text.WriteString("\nThe run's output is not there to read.\n")
	case cut:
		fmt.Fprintf(&text, "\nThe end of the run's output, cut to its last %d lines or %d KiB:\n\n```\n%s\n```\n", aboutLines, max(limit>>10, 1), output)
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

// outputText returns a run's output as text, its colours and cursor
// moves played out, from its recording; empty when there is none.
func (b *Board) outputText(run *model.Run) string {
	if b.Recordings == nil {
		return ""
	}
	cast, err := b.Recordings.Read(run)
	if errors.Is(err, recording.ErrExpired) {
		// The prompt reads that the run has no output any more.
		return recording.Expired
	}
	if err != nil {
		return ""
	}
	defer cast.Close()
	replay, err := runner.ReplayCast(cast)
	if err != nil {
		return ""
	}
	return strings.TrimRight(replay.Output.Text(), " \n")
}

// cutTail keeps the last lines of an output, as many as fit aboutLines
// and limit bytes, and reports whether it cut any.
func cutTail(output string, limit int) (string, bool) {
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
		if size > limit {
			lines, cut = lines[i+1:], true
			break
		}
	}
	return strings.Join(lines, "\n"), cut
}
