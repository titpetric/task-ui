package note

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/titpetric/task-ui/model"
)

// TriageQuestion is the question of the thread a schedule's triage asks
// about its failed run; the prompt run sets claude.RunTriage beside it.
const TriageQuestion = "Triage this failed run."

// The verdicts of a triage, the Result line Claude's answer ends with.
const (
	TriageResolved   = "resolved"
	TriageUnresolved = "not resolved"
	TriageNoAction   = "no action"
)

// verdictLine is the Result line of a triage's answer, markdown emphasis
// and a closing stop allowed around it.
var verdictLine = regexp.MustCompile(`(?i)^[\s*_\x60]*result:\s*(resolved|not\s+resolved|no\s+action)[\s*_\x60.]*$`)

// Verdict returns the verdict of Claude's answer, its last Result line
// whatever its case and spacing, or TriageUnresolved without one.
func Verdict(answer string) string {
	lines := strings.Split(answer, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		m := verdictLine.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}
		return strings.Join(strings.Fields(strings.ToLower(m[1])), " ")
	}
	return TriageUnresolved
}

// TriageNote returns the note a triage leaves on the board: the task, the verdict and the run's link under baseURL, a path without one.
func TriageNote(baseURL, folder, job, runID, verdict string) string {
	link := strings.TrimSuffix(baseURL, "/") + "/runs/" + runID
	return fmt.Sprintf("Attempted triage of failing %s/%s task, result: %s. Info: [%s](%s)", folder, job, verdict, link, link)
}

// TriageRun asks Claude to triage a failed run of a schedule, as the
// schedule's user, in a thread on the run's page, and waits for the
// answer; then it leaves a note on the board naming the task, the
// verdict and the run, and returns the note and the verdict. A user
// without prompt/edit gets ErrForbidden, a run the user does not see,
// or Claude's own, ErrNoPrompt, and a thread whose prompt did not run
// ErrPromptFailed, each before any note; a prompt that ran and failed
// ends TriageUnresolved, noted. A stop ends the prompt's run and notes
// nothing.
func (b *Board) TriageRun(stop context.Context, user *model.User, runID string) (*model.Note, string, error) {
	ctx := context.Background()
	if b.Runs == nil {
		return nil, "", fmt.Errorf("%w: prompts about runs aren't available on this instance", ErrNoPrompt)
	}
	if user == nil {
		return nil, "", fmt.Errorf("%w: a prompt needs an author", ErrInvalid)
	}
	run, err := b.Runs.Get(ctx, runID)
	if err != nil {
		return nil, "", err
	}
	if !user.CanSee(run.Folder) || run.IsClaude() {
		return nil, "", fmt.Errorf("%w: no such run", ErrNoPrompt)
	}
	if !b.permissions(ctx, user).Allowed("prompt/edit") {
		return nil, "", ErrForbidden
	}
	req, err := b.promptRequest(ctx, user, run.Where(), false)
	if err != nil {
		return nil, "", err
	}
	thread := &model.Note{
		Channel: model.ChannelNotes, UserID: user.ID, UserName: user.Name, Body: TriageQuestion,
		Dir: run.Where(), AboutRunID: run.ID, State: model.NoteOpen,
	}
	if thread.WorkspaceID = b.Runner.WorkspaceID(ctx, req); thread.WorkspaceID == "" {
		thread.WorkspaceID = model.Default
	}
	if err := b.Notes.Create(ctx, thread); err != nil {
		return nil, "", err
	}
	b.publish(thread)
	req, answer, cleanup, ok := b.promptEnv(thread, req, nil)
	if !ok {
		return nil, "", ErrPromptFailed
	}
	defer cleanup()
	req.Env = append(req.Env, b.hosts(ctx)...)
	answered := b.run(stop, thread, req, answer, "The triage", nil)
	if stop.Err() != nil {
		return nil, "", stop.Err()
	}
	verdict := TriageUnresolved
	if answered && thread.State == model.NoteDone {
		verdict = Verdict(b.lastReply(ctx, thread))
	}
	note, err := b.Post(ctx, user, "", TriageNote(b.BaseURL, run.Folder, run.Job, run.ID, verdict))
	return note, verdict, err
}
