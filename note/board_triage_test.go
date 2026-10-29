package note_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/note"
)

// The verdict is the last Result line, in any case, spacing or
// emphasis; without one the verdict is not resolved.
func TestVerdict(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name, answer, verdict string
	}{
		{"resolved", "The script read a file that was not there.\n\nResult: resolved\nAdded config.yml and committed it.", note.TriageResolved},
		{"not resolved", "Result: not resolved\nThe registry is down.", note.TriageUnresolved},
		{"no action", "Result: no action\nTransient network error.", note.TriageNoAction},
		{"case and spacing", "  result:   NO   ACTION  ", note.TriageNoAction},
		{"bold", "**Result: resolved**", note.TriageResolved},
		{"code and stop", "`Result: not resolved`.", note.TriageUnresolved},
		{"the last line decides", "Result: no action\n\nOn a second look:\nResult: resolved\nFixed the path.", note.TriageResolved},
		{"missing", "I looked at the run and found nothing.", note.TriageUnresolved},
		{"empty", "", note.TriageUnresolved},
		{"in a sentence", "The result: resolved, I think.", note.TriageUnresolved},
		{"another word", "Result: fixed", note.TriageUnresolved},
	}
	for _, c := range cases {
		assert.Equal(t, c.verdict, note.Verdict(c.answer), c.name)
	}
}

// The note names the task, the verdict and the run's link, the link's
// text the link itself, under the base URL or as a path without one.
func TestTriageNote(t *testing.T) {
	t.Parallel()
	assert.Equal(t,
		"Attempted triage of failing infra/plan task, result: resolved. Info: [https://task.example/runs/r1](https://task.example/runs/r1)",
		note.TriageNote("https://task.example/", "infra", "plan", "r1", note.TriageResolved))
	assert.Equal(t,
		"Attempted triage of failing acme/api/test task, result: no action. Info: [/runs/r2](/runs/r2)",
		note.TriageNote("", "acme/api", "test", "r2", note.TriageNoAction))
}
