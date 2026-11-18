package meeting

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/integration/claude"
)

// SummaryPrompt is what Claude reads a meeting's transcript with.
const SummaryPrompt = "The input is the transcript of a meeting in task-ui, one line per speaker, in order. " +
	"Summarise it in markdown under exactly these four headings: `## Topics discussed`, `## Action items`, `## Decisions`, `## Open questions`. " +
	"Use short bullets under each heading, and name who does what in an action item where the transcript says so. " +
	"Write `None.` under a heading the transcript does not support. Change nothing."

// summarize asks Claude about an ended room's transcript and writes the reply back to its row.
func (s *Service) summarize(m *model.Meeting) {
	ctx := context.Background()
	lines, err := s.Store.Lines(ctx, m.ID, 0)
	if err != nil {
		s.finishSummary(ctx, m.ID, "", model.SummaryFailed, err.Error())
		return
	}
	env, role := s.envFor(ctx, m.UserID)
	if env == nil {
		s.finishSummary(ctx, m.ID, "", model.SummaryFailed, "the creator's sign-in could not be read")
		return
	}
	text, err := s.Summarize(ctx, env, claude.Context(m.UserName, role, ""), SummaryPrompt, transcriptReport(m, lines))
	if err != nil {
		s.logger().Warn("meetings: summary failed", "meeting", m.ID, "err", err)
		s.finishSummary(ctx, m.ID, "", model.SummaryFailed, err.Error())
		return
	}
	s.finishSummary(ctx, m.ID, text, model.SummaryDone, "")
}

// envFor reads the creator's integration variables and role.
func (s *Service) envFor(ctx context.Context, userID string) (map[string]string, string) {
	user, err := s.Users.Get(ctx, userID)
	if err != nil {
		return nil, ""
	}
	role := user.EffectiveRole()
	env, err := s.Env(ctx, user.Name, role)
	if err != nil {
		return nil, ""
	}
	return env, role
}

// finishSummary stores a summary and tells the room's pipes, if any are still open.
func (s *Service) finishSummary(ctx context.Context, id, text, state, message string) {
	m, err := s.Store.Get(ctx, id)
	if err != nil {
		s.logger().Warn("meetings: summary not stored", "meeting", id, "err", err)
		return
	}
	m.Summary, m.SummaryState, m.SummaryMessage = text, state, message
	if err := s.Store.Update(ctx, m); err != nil {
		s.logger().Warn("meetings: summary not stored", "meeting", id, "err", err)
		return
	}
	s.Hub.Broadcast(id, nil, TextFrame(Message{Type: "summary", State: state, Message: message}))
}

// transcriptReport renders a room's log as the document Claude reads.
func transcriptReport(m *model.Meeting, lines []model.MeetingLine) string {
	var b strings.Builder
	b.WriteString("Meeting: " + m.Title + "\n")
	if m.StartedAt != nil {
		b.WriteString("Started: " + m.StartedAt.UTC().Format(time.RFC3339) + "\n")
	}
	b.WriteString("\nTranscript, in order, each line under the name of the person who said it:\n\n")
	said := 0
	for _, line := range lines {
		if line.Kind != model.LineSaid {
			continue
		}
		said++
		fmt.Fprintf(&b, "%s: %s\n", line.UserName, line.Body)
	}
	if said == 0 {
		b.WriteString("(Nobody's speech was transcribed.)\n")
	}
	return b.String()
}
