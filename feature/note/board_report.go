package note

import (
	"context"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// Report returns the report of a run, nil for none: a run of Claude's,
// its answer once it ended; another run, its summary, the batch's for a
// run of a batch. A pending summary is no report.
func (b *Board) Report(ctx context.Context, run *model.Run) *Report {
	if run == nil {
		return nil
	}
	if run.IsPrompt() {
		return b.answerOf(ctx, run)
	}
	about := run.ID
	if run.BatchID != "" && b.Runs != nil {
		if batch, err := b.Runs.List(ctx, storage.RunFilter{BatchID: run.BatchID, Limit: 200}); err == nil && len(batch) > 0 {
			about = batch[len(batch)-1].ID
		}
	}
	report := b.Summaries(ctx, []string{about})[about]
	if report == nil || report.Pending {
		return nil
	}
	report.RunID = run.ID
	return report
}

// answerOf returns the report of a run of Claude's on a prompt: the
// answer it wrote, nil before it ended.
func (b *Board) answerOf(ctx context.Context, run *model.Run) *Report {
	if !model.Finished(run.State) {
		return nil
	}
	thread, err := b.Notes.ByRun(ctx, run.ID)
	if err != nil {
		return nil
	}
	replies, _ := b.Notes.Replies(ctx, []string{thread.ID})
	answer := Answer(replies[thread.ID], run.ID)
	if answer == nil {
		return nil
	}
	question := Question(thread, replies[thread.ID], answer)
	return &Report{
		RunID: run.ID, AboutRunID: thread.AboutRunID, ThreadID: thread.ID,
		Question: question.Body, By: question.UserName, Body: answer.Body, At: answer.CreatedAt,
	}
}
