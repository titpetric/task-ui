package note_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// ran runs a job of proj as the user to its end, a step of a workflow
// run.
func ran(t *testing.T, f *fixture, user *model.User, id, script string) *model.Run {
	t.Helper()
	proj := f.board.Catalogue.Find("proj")
	job := &folder.Job{Folder: "proj", ID: id, Kind: "script", Command: []string{"sh", "-c", script}}
	job.Name = id
	session, err := f.board.Runner.Start(context.Background(), runner.Request{Folder: proj, Job: job, User: user, Cols: 80, Rows: 24})
	assert.NoError(t, err)
	<-session.Done()
	run, err := f.board.Runs.Get(context.Background(), session.ID())
	assert.NoError(t, err)
	return run
}

// AskWorkflow puts the workflow's prompt, with the run's input, in the
// prompt's place, as a thread about the workflow run: Claude gets every
// step that ran with the end of its output, and the skipped ones, in
// TASK_UI_PROMPT_WORKFLOW, and AskWorkflow returns with Claude's reply
// in. A reply in the thread runs again with the run's steps.
func TestBoard_AskWorkflow(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	admin := f.withPrompts(t)
	ctx := context.Background()
	f.board.AgentCommand = []string{"sh", "-c", `{ printf 'Q: %s\n' "$TASK_UI_PROMPT"; cat "$TASK_UI_PROMPT_WORKFLOW"; } > "$TASK_UI_PROMPT_REPLY"`}

	build := ran(t, f, admin, "build", "echo built image v1.2")
	test := ran(t, f, admin, "test", "echo 3 tests passed; exit 1")
	wf := &model.Workflow{Name: "release", Prompt: "summarize, change nothing", PromptFolder: "proj/images"}
	run := &model.WorkflowRun{WorkflowID: model.NewID(), Name: wf.Name, UserID: admin.ID, UserName: admin.Name}
	assert.NoError(t, f.board.WorkflowRuns.Create(ctx, run, []model.WorkflowStep{{Folder: "proj", Job: "build"}, {Folder: "proj", Job: "test"}, {Folder: "proj", Job: "push"}}))
	steps, err := f.board.WorkflowRuns.Steps(ctx, run.ID)
	assert.NoError(t, err)
	for i, state := range []string{model.StateOK, model.StateFailed, model.StepSkipped} {
		steps[i].State = state
	}
	steps[0].RunID, steps[1].RunID = build.ID, test.ID
	for i := range steps {
		assert.NoError(t, f.board.WorkflowRuns.UpdateStep(ctx, &steps[i]))
	}

	assert.NoError(t, f.board.AskWorkflow(ctx, admin, run, wf, "for the 1.2 tag"))
	threads, err := f.board.Notes.List(ctx, storage.NoteFilter{Prompts: true, AboutWorkflowRunID: run.ID})
	assert.NoError(t, err)
	assert.Len(t, threads, 1)
	thread := threads[0]
	assert.Equal(t, "summarize, change nothing\n\nInput:\nfor the 1.2 tag", thread.Body)
	assert.Equal(t, "proj/images", thread.Dir)
	assert.Equal(t, model.NoteDone, thread.State)
	answer := replies(t, f, thread.ID)
	assert.Len(t, answer, 1)
	body := answer[0].Body
	assert.Equal(t, run.ID, answer[0].AboutWorkflowRunID)
	assert.True(t, strings.HasPrefix(body, "Q: summarize, change nothing\n\nInput:\nfor the 1.2 tag\nWorkflow run "+run.ID+" of the workflow release, by admin, with 3 steps."), body)
	assert.Contains(t, body, "## Step 1, proj / build: ok, exit code 0")
	assert.Contains(t, body, "built image v1.2")
	assert.Contains(t, body, "## Step 2, proj / test: failed, exit code 1")
	assert.Contains(t, body, "3 tests passed")
	assert.True(t, strings.HasSuffix(body, "## Step 3, proj / push: skipped"), body)

	// The thread's place keeps it apart from the directory's prompts.
	here, _ := f.board.Notes.List(ctx, storage.NoteFilter{Prompts: true, Dir: "proj/images"})
	assert.Len(t, here, 0)

	_, err = f.board.Ask(ctx, admin, thread.ID, "", "and the tests")
	assert.NoError(t, err)
	assert.Eventually(t, func() bool { return len(replies(t, f, thread.ID)) == 3 }, 10*time.Second, 20*time.Millisecond)
	again := replies(t, f, thread.ID)[2]
	assert.Contains(t, again.Body, "built image v1.2")
	assert.Equal(t, run.ID, again.AboutWorkflowRunID)
}

// A prompt-only workflow's run has no steps to tell; a prompt that
// fails returns ErrPromptFailed, and one without prompt/edit or a place
// the user sees does not start. A stop ends the prompt's run.
func TestBoard_AskWorkflow_refused(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	admin := f.withPrompts(t)
	ctx := context.Background()
	f.board.AgentCommand = []string{"sh", "-c", `cat "$TASK_UI_PROMPT_WORKFLOW" > "$TASK_UI_PROMPT_REPLY"; case "$TASK_UI_PROMPT" in *fail*) exit 3;; *slow*) sleep 30;; esac`}
	run := &model.WorkflowRun{WorkflowID: model.NewID(), Name: "alone", UserID: admin.ID, UserName: admin.Name}
	assert.NoError(t, f.board.WorkflowRuns.Create(ctx, run, nil))

	assert.NoError(t, f.board.AskWorkflow(ctx, admin, run, &model.Workflow{Name: "alone", Prompt: "say ok"}, ""))
	threads, _ := f.board.Notes.List(ctx, storage.NoteFilter{Prompts: true, AboutWorkflowRunID: run.ID})
	assert.Len(t, threads, 1)
	assert.Equal(t, model.Default, threads[0].Dir)
	assert.Contains(t, replies(t, f, threads[0].ID)[0].Body, "The workflow has no steps: the prompt is the whole of it.")

	assert.ErrorIs(t, f.board.AskWorkflow(ctx, admin, run, &model.Workflow{Prompt: "fail"}, ""), note.ErrPromptFailed)
	assert.ErrorIs(t, f.board.AskWorkflow(ctx, f.ops, run, &model.Workflow{Prompt: "say ok"}, ""), note.ErrForbidden)
	assert.ErrorIs(t, f.board.AskWorkflow(ctx, admin, run, &model.Workflow{Prompt: "say ok", PromptFolder: "gone"}, ""), note.ErrNoPrompt)
	assert.ErrorIs(t, f.board.AskWorkflow(ctx, admin, run, &model.Workflow{}, ""), note.ErrInvalid)
	assert.ErrorIs(t, (&note.Board{}).AskWorkflow(ctx, admin, run, &model.Workflow{Prompt: "say ok"}, ""), note.ErrNoPrompt)
	assert.NoError(t, f.board.PromptPlace(ctx, admin, "proj/images"))
	assert.ErrorIs(t, f.board.PromptPlace(ctx, admin, "proj/gone"), note.ErrNoPrompt)

	stop, cancel := context.WithCancel(ctx)
	go func() {
		time.Sleep(300 * time.Millisecond)
		cancel()
	}()
	started := time.Now()
	assert.ErrorIs(t, f.board.AskWorkflow(stop, admin, run, &model.Workflow{Prompt: "slow"}, ""), context.Canceled)
	assert.True(t, time.Since(started) < 20*time.Second, "the stop ends the prompt's run")
	slow, _ := f.board.Notes.List(ctx, storage.NoteFilter{Prompts: true, AboutWorkflowRunID: run.ID})
	assert.Equal(t, model.NoteFailed, slow[0].State)
}

// The question of a workflow run's prompt is the prompt, and the input
// under Input: when the run has one.
func TestWorkflowQuestion(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "say ok", note.WorkflowQuestion(" say ok ", " "))
	assert.Equal(t, "say ok\n\nInput:\nloud", note.WorkflowQuestion("say ok", "loud\n"))
}

// The steps share 256 KiB of output, the newest first, each at most the
// last 2000 lines a run's prompt gets; a step's output cut to fit says
// so, and keeps its end.
func TestBoard_AskWorkflow_cut(t *testing.T) {
	t.Parallel()
	f := newFixture(t, "")
	admin := f.withPrompts(t)
	ctx := context.Background()
	// The reply holds 64 KiB; it says what the file holds instead.
	f.board.AgentCommand = []string{"sh", "-c", `F="$TASK_UI_PROMPT_WORKFLOW"; { grep 'cut to fit' "$F"; tail -n 3 "$F"; echo "first: $(grep -c 'line 0 of the' "$F")"; echo "bytes: $(wc -c < "$F")"; } > "$TASK_UI_PROMPT_REPLY"`}
	big := ran(t, f, admin, "big", "i=0; while [ $i -lt 4000 ]; do echo line $i of the big step padded out to a hundred bytes or so, more or less; i=$((i+1)); done")
	run := &model.WorkflowRun{WorkflowID: model.NewID(), Name: "big", UserID: admin.ID, UserName: admin.Name}
	assert.NoError(t, f.board.WorkflowRuns.Create(ctx, run, []model.WorkflowStep{{Folder: "proj", Job: "big"}}))
	steps, _ := f.board.WorkflowRuns.Steps(ctx, run.ID)
	steps[0].State, steps[0].RunID = model.StateOK, big.ID
	assert.NoError(t, f.board.WorkflowRuns.UpdateStep(ctx, &steps[0]))

	assert.NoError(t, f.board.AskWorkflow(ctx, admin, run, &model.Workflow{Prompt: "summarize"}, ""))
	threads, _ := f.board.Notes.List(ctx, storage.NoteFilter{Prompts: true, AboutWorkflowRunID: run.ID})
	body := replies(t, f, threads[0].ID)[0].Body
	assert.Contains(t, body, "The end of its output, cut to fit the 256 KiB the steps share:")
	assert.Contains(t, body, "line 3999 of the big step")
	assert.Contains(t, body, "first: 0")
	var size int
	_, err := fmt.Sscanf(body[strings.Index(body, "bytes: "):], "bytes: %d", &size)
	assert.NoError(t, err)
	assert.True(t, size > 100<<10 && size < 260<<10, size)
}
