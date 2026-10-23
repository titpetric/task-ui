package note

import (
	"time"

	"github.com/titpetric/task-ui/model"
)

// Report is the markdown a run stands for, without its terminal: the
// answer a run of Claude's wrote to a prompt, or the summary Claude
// wrote of another run.
type Report struct {
	// RunID is the run the report is of; AboutRunID the run Claude
	// summarized or asked about, empty for a prompt about no run.
	RunID      string
	AboutRunID string
	// ThreadID is the prompt thread that holds the reply, empty for a
	// summary kept without one.
	ThreadID string
	// Question is what Claude answered, and By who asked it.
	Question string
	By       string
	// Body is Claude's reply, in markdown; empty while Pending.
	Body string
	At   *time.Time
	// Pending says a question waits on Claude's reply.
	Pending bool
}

// Href returns the report page of the run.
func (r *Report) Href() string {
	return "/runs/" + r.RunID + "/report"
}

// SummaryOf returns the report of a summary thread with its replies:
// the first answer, the summary, or a pending report while the thread
// waits on it; nil for a thread that is no summary or one that failed
// without an answer.
func SummaryOf(thread *model.Note, replies []model.Note) *Report {
	if !IsSummary(thread) {
		return nil
	}
	report := &Report{RunID: thread.AboutRunID, AboutRunID: thread.AboutRunID, ThreadID: thread.ID, Question: thread.Body, By: thread.UserName}
	if answer := Answer(replies, SummaryRunID(thread, replies)); answer != nil {
		report.Body, report.At = answer.Body, answer.CreatedAt
		return report
	}
	if thread.State == model.NoteFailed {
		return nil
	}
	report.Pending = true
	return report
}

// IsSummary reports whether the report is a run's summary.
func (r *Report) IsSummary() bool {
	asked := r.Question == SummaryQuestion || r.Question == BatchSummaryQuestion
	return r.ThreadID == "" || asked
}
