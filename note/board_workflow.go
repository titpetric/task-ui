package note

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/titpetric/task-ui/model"
)

// workflowBytes is the most of the steps' output a workflow run's
// prompt gets, all its steps together.
const workflowBytes = 256 << 10

// ErrPromptFailed reports a workflow run's prompt that ended without an
// answer: its run failed, or it did not start.
var ErrPromptFailed = errors.New("its run failed, see the thread")

// WorkflowQuestion is the question of a workflow run's prompt: the
// workflow's prompt, and the run's input under "Input:".
func WorkflowQuestion(prompt, input string) string {
	prompt, input = strings.TrimSpace(prompt), strings.TrimSpace(input)
	if input == "" {
		return prompt
	}
	return prompt + "\n\nInput:\n" + input
}

// PromptPlace checks a place a workflow's prompt runs in, as the user:
// a folder or a directory in it the user sees, or _ for the root.
func (b *Board) PromptPlace(ctx context.Context, user *model.User, place string) error {
	_, err := b.promptRequest(ctx, user, place)
	return err
}

// AskWorkflow puts a workflow's prompt to Claude once a run's steps
// ended, as the user who started the run, and returns when Claude
// replied: a thread about the workflow run in the workflow's prompt
// place, which Claude gets with the steps' output in
// TASK_UI_PROMPT_WORKFLOW. A stop ends the prompt's run. It returns nil
// for a reply, the stop's error after a stop, and ErrPromptFailed, or
// why it did not start, otherwise. It takes prompt/edit.
func (b *Board) AskWorkflow(stop context.Context, user *model.User, run *model.WorkflowRun, wf *model.Workflow, input string) error {
	ctx := context.Background()
	if b.WorkflowRuns == nil {
		return fmt.Errorf("%w: workflow prompts aren't available on this instance", ErrNoPrompt)
	}
	body := WorkflowQuestion(wf.Prompt, input)
	if user == nil || body == "" {
		return fmt.Errorf("%w: a prompt needs an author and some text", ErrInvalid)
	}
	if !b.permissions(ctx, user).Allowed("prompt/edit") {
		return ErrForbidden
	}
	place := wf.PromptPlace()
	req, err := b.promptRequest(ctx, user, place)
	if err != nil {
		return err
	}
	note := &model.Note{
		Channel: model.ChannelNotes, UserID: user.ID, UserName: user.Name, Body: body,
		Dir: place, AboutWorkflowRunID: run.ID, State: model.NoteOpen,
	}
	if note.WorkspaceID = b.Runner.WorkspaceID(ctx, req); note.WorkspaceID == "" {
		note.WorkspaceID = model.Default
	}
	if err := b.Notes.Create(ctx, note); err != nil {
		return err
	}
	b.publish(note)
	req, answer, cleanup, ok := b.promptEnv(note, req, nil)
	if !ok {
		return ErrPromptFailed
	}
	defer cleanup()
	req.Env = append(req.Env, b.hosts(ctx)...)
	b.run(stop, note, req, answer, "The prompt", nil)
	switch {
	case stop.Err() != nil:
		return stop.Err()
	case note.State != model.NoteDone:
		return ErrPromptFailed
	}
	return nil
}

// aboutWorkflowRun writes what a workflow run's prompt tells Claude of
// the run to a file, for TASK_UI_PROMPT_WORKFLOW: the workflow, and each
// step, how it ended and the end of its output. It returns the file's
// path.
func (b *Board) aboutWorkflowRun(ctx context.Context, id string) (string, error) {
	if b.WorkflowRuns == nil || b.Runs == nil {
		return "", fmt.Errorf("%w: workflow prompts aren't available on this instance", ErrNoPrompt)
	}
	run, err := b.WorkflowRuns.Get(ctx, id)
	if err != nil {
		return "", err
	}
	steps, err := b.WorkflowRuns.Steps(ctx, id)
	if err != nil {
		return "", err
	}
	file, err := os.CreateTemp("", "note-workflow-*.md")
	if err != nil {
		return "", err
	}
	_, err = file.WriteString(b.workflowText(ctx, run, steps))
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}

// stepText is a step of a workflow run as its prompt describes it.
type stepText struct {
	head   string
	output string
	cut    bool
}

// workflowText describes a workflow run for its prompt: the workflow,
// then each step in order, how it ended, its exit code and time, and
// the end of its output, ANSI played out. The steps share workflowBytes
// of output, the newest first: each takes at most an even share of
// the rest, and a step that needs less hands what it does not use to
// the steps before it.
func (b *Board) workflowText(ctx context.Context, run *model.WorkflowRun, steps []model.WorkflowRunStep) string {
	texts := make([]stepText, len(steps))
	var ran []int
	for i, step := range steps {
		texts[i].head = b.stepHead(ctx, i, step)
		if step.RunID != "" {
			ran = append(ran, i)
		}
	}
	left := workflowBytes
	for n := len(ran) - 1; n >= 0; n-- {
		i := ran[n]
		row, err := b.Runs.Get(ctx, steps[i].RunID)
		if err != nil {
			continue
		}
		texts[i].output, texts[i].cut = cutTail(b.outputText(row), left/(n+1))
		left -= len(texts[i].output)
	}

	var text strings.Builder
	fmt.Fprintf(&text, "Workflow run %s of the workflow %s, by %s, with %d step%s.\n", run.ID, run.Name, run.UserName, len(steps), plural(len(steps)))
	if len(steps) == 0 {
		text.WriteString("The workflow has no steps: the prompt is the whole of it.\n")
	}
	for _, t := range texts {
		text.WriteString("\n## " + t.head + "\n")
		switch {
		case t.output != "" && t.cut:
			fmt.Fprintf(&text, "\nThe end of its output, cut to fit the %d KiB the steps share:\n\n```\n%s\n```\n", workflowBytes>>10, fence(t.output))
		case t.output != "":
			text.WriteString("\nIts output:\n\n```\n" + fence(t.output) + "\n```\n")
		}
	}
	return text.String()
}

// stepHead is the line that names a step of a workflow run and says how
// it ended.
func (b *Board) stepHead(ctx context.Context, i int, step model.WorkflowRunStep) string {
	name := step.Folder + " / " + step.Job
	if step.IsWorkflow() {
		name = "the workflow " + step.WorkflowRef
		if step.WorkflowRunRef != "" {
			if inner, err := b.WorkflowRuns.Get(ctx, step.WorkflowRunRef); err == nil {
				name = "the workflow " + inner.Name + ", run " + inner.ID
			}
		}
	}
	head := fmt.Sprintf("Step %d, %s: %s", i+1, name, step.State)
	if step.RunID == "" {
		return head
	}
	row, err := b.Runs.Get(ctx, step.RunID)
	if err != nil {
		return head + ", run " + step.RunID
	}
	head += fmt.Sprintf(", exit code %d", row.ExitCode)
	if row.StartedAt != nil {
		head += ", took " + row.Elapsed().Round(time.Second).String()
	}
	return head + ", run " + row.ID
}

// fence keeps an output from closing the code block it goes in.
func fence(output string) string {
	return strings.ReplaceAll(output, "```", "'''")
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
