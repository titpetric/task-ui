package meeting

import (
	"testing"
	"time"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
)

// A transcript report names the meeting and each said line under its speaker, and leaves the system lines out.
func TestTranscriptReport(t *testing.T) {
	t.Parallel()
	started := time.Date(2026, 11, 18, 9, 0, 0, 0, time.UTC)
	room := &model.Meeting{Title: "Standup", StartedAt: &started}
	lines := []model.MeetingLine{
		{Kind: model.LineJoin, UserName: "ana", Body: "ana joined"},
		{Kind: model.LineSaid, UserName: "ana", Body: "we ship today"},
		{Kind: model.LineSaid, UserName: "bob", Body: "I will deploy"},
	}
	report := transcriptReport(room, lines)
	assert.Contains(t, report, "Meeting: Standup")
	assert.Contains(t, report, "ana: we ship today")
	assert.Contains(t, report, "bob: I will deploy")
	assert.NotContains(t, report, "joined")

	empty := transcriptReport(room, nil)
	assert.Contains(t, empty, "Nobody's speech was transcribed")
	assert.Contains(t, SummaryPrompt, "## Topics discussed")
	assert.Contains(t, SummaryPrompt, "## Open questions")
}
