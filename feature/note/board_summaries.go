package note

import (
	"context"

	"github.com/titpetric/task-ui/storage"
)

// Summaries returns the reports of the runs that have a summary, or one under way, by run ID.
// A run's report is the oldest summary thread about it that did not fail, of Run and summarize or Summarize,
// else the summary the Runs page's Summary kept.
func (b *Board) Summaries(ctx context.Context, runIDs []string) map[string]*Report {
	reports := map[string]*Report{}
	if len(runIDs) == 0 {
		return reports
	}
	threads, err := b.Notes.List(ctx, storage.NoteFilter{Prompts: true, AboutRunIDs: runIDs})
	if err == nil {
		ids := make([]string, 0, len(threads))
		for i := range threads {
			if IsSummary(&threads[i]) {
				ids = append(ids, threads[i].ID)
			}
		}
		replies, _ := b.Notes.Replies(ctx, ids)
		// Newest first: the oldest thread comes last and stays.
		for i := range threads {
			if report := SummaryOf(&threads[i], replies[threads[i].ID]); report != nil {
				if kept := reports[report.RunID]; kept == nil || kept.Pending || report.Body != "" {
					reports[report.RunID] = report
				}
			}
		}
	}
	if b.RunSummaries == nil {
		return reports
	}
	for _, id := range runIDs {
		if reports[id] != nil {
			continue
		}
		if kept, err := b.RunSummaries.Get(ctx, id); err == nil && kept.Summary != "" {
			reports[id] = &Report{RunID: id, AboutRunID: id, By: kept.UserName, Body: kept.Summary, At: kept.CreatedAt}
		}
	}
	return reports
}
