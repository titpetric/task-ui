package askhuman_test

import (
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/feature/askhuman"
	"github.com/titpetric/task-ui/internal/assert"
)

// questionnaire is the askhuman spelling of a form: sections, kinds and the
// option fields a decision record keeps.
const questionnaire = `id: release
project: malta
workflow_id: release
title: Release
subtitle: Two decisions
context: The tests are green.
prerequisites:
  - the image is pushed
sections:
  - {id: rollout, title: Rollout, description: How it ships}
fields:
  - key: target
    label: Where
    kind: qcm_free_text
    section: rollout
    axis: environment
    required: true
    options:
      - {value: staging, label: Staging, rationale: Rehearsal, recommended: true, default_checked: true}
      - {value: prod, label: Production, rationale: The real thing}
    free_text: true
    free_text_label: Why
  - key: scope
    label: What ships
    kind: boolean_checklist
    section: rollout
    options:
      - {id: api, title: The API}
      - {value: ui, label: The UI}
  - key: notes
    label: Anything else
    kind: free_text
submit:
  prompt: "Ship {{target}}"
`

// TestForm_ArtifactPath names the two files a submission leaves.
func TestForm_ArtifactPath(t *testing.T) {
	t.Parallel()
	f, err := askhuman.Parse([]byte(questionnaire))
	assert.NoError(t, err)
	jsonPath, mdPath := f.ArtifactPath("01hzNote", 0)
	assert.Equal(t, "ask/01hznote/release.answers.json", jsonPath)
	assert.Equal(t, "ask/01hznote/release.decisions.md", mdPath)
	unnamed := &askhuman.Form{}
	jsonPath, mdPath = unnamed.ArtifactPath("01hz", 2)
	assert.Equal(t, "ask/01hz/ask-3.answers.json", jsonPath)
	assert.Equal(t, "ask/01hz/ask-3.decisions.md", mdPath)
}

// TestParse_Questions reads the askhuman spelling into the task-ui one.
func TestParse_Questions(t *testing.T) {
	t.Parallel()
	f, err := askhuman.Parse([]byte(questionnaire))
	assert.NoError(t, err)
	assert.Equal(t, "release", f.ID)
	assert.Equal(t, "malta", f.Project)
	assert.Equal(t, "The tests are green.", f.Context)
	assert.Len(t, f.Sections, 1)
	assert.Len(t, f.Fields, 3)
	assert.Equal(t, "choice", f.Fields[0].Component)
	assert.Equal(t, "staging", f.Fields[0].Default)
	assert.Equal(t, "Rehearsal", f.Fields[0].Options[0].Rationale)
	assert.True(t, f.Fields[0].Options[0].Recommended)
	assert.Equal(t, "checklist", f.Fields[1].Component)
	assert.Equal(t, "api", f.Fields[1].Options[0].Value)
	assert.Equal(t, "The API", f.Fields[1].Options[0].Label)
	assert.Equal(t, "textarea", f.Fields[2].Component)
}

// TestParse_QuestionsErrors refuses the shapes the dialect cannot draw.
func TestParse_QuestionsErrors(t *testing.T) {
	t.Parallel()
	for name, test := range map[string]struct {
		source string
		want   string
	}{
		"fields and questions": {"fields: [{key: a, component: text}]\nquestions: [{id: b, title: B, kind: free_text}]\n", "fields and questions"},
		"a qcm with one option": {"questions:\n  - {id: a, title: A, kind: qcm_free_text, options: [{id: x, label: X}]}\n", "needs two options"},
		"an unknown kind": {"questions:\n  - {id: a, title: A, kind: ranking}\n", `no kind "ranking"`},
		"a section to nowhere": {"fields:\n  - {key: a, component: text, section: nope}\n", `section "nope" is not declared`},
		"a section twice": {"sections: [{id: s}, {id: s}]\nfields:\n  - {key: a, component: text}\n", `section "s"`},
		"free text off a qcm": {"fields:\n  - {key: a, component: text, free_text: true}\n", "free text follows a choice or a multi"},
	} {
		f, err := askhuman.Parse([]byte(test.source))
		assert.Nil(t, f, name)
		assert.ErrorContains(t, err, test.want, name)
	}
}

// TestForm_Answer binds a questionnaire and keeps the record askhuman's sink writes.
func TestForm_Answer(t *testing.T) {
	t.Parallel()
	f, err := askhuman.Parse([]byte(questionnaire))
	assert.NoError(t, err)
	values := map[string]string{
		"target":      "prod",
		"target.text": "the release is rehearsed",
		"scope":       "api",
		"notes":       "ship the API alone",
	}
	sub, err := f.Bind(values)
	assert.NoError(t, err)
	answers := f.Answer(sub, values, time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC), "miguel")

	assert.Equal(t, "malta", answers.Project)
	assert.Equal(t, "2026-11-21T12:00:00Z", answers.SubmittedAt)
	assert.Equal(t, "miguel", answers.SubmittedBy)
	assert.Len(t, answers.Answers, 3)
	assert.Equal(t, []string{"prod"}, answers.Answers[0].SelectedOptions)
	assert.Equal(t, "the release is rehearsed", answers.Answers[0].FreeText)
	assert.Equal(t, "environment", answers.Answers[0].Axis)
	assert.Equal(t, map[string]bool{"api": true, "ui": false}, answers.Answers[1].BooleanChecks)
	assert.Equal(t, "ship the API alone", answers.Answers[2].FreeText)

	record := string(answers.Markdown(f))
	for _, want := range []string{"# Release", "## Rollout", "### Where", "**Production**", "The real thing", "**The API**: yes", "**The UI**: no", "ship the API alone"} {
		assert.True(t, strings.Contains(record, want), want)
	}
}

// A required qcm is answered by the reason beside it, with no option selected.
func TestForm_Answer_Reason(t *testing.T) {
	t.Parallel()
	f, err := askhuman.Parse([]byte("questions:\n" +
		"  - {id: target, title: Where, kind: qcm_free_text, required: true, free_text: true, options: [{id: staging, label: Staging}, {id: prod, label: Production}]}\n"))
	assert.NoError(t, err)
	values := map[string]string{"target.text": "decide for me"}
	sub, err := f.Bind(values)
	assert.NoError(t, err)
	answers := f.Answer(sub, values, time.Unix(0, 0), "")
	assert.Empty(t, answers.Answers[0].SelectedOptions)
	assert.Equal(t, "decide for me", answers.Answers[0].FreeText)
}
