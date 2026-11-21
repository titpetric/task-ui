package askhuman

import (
	"fmt"
	"strings"
	"time"
)

// Answer is one question answered, as answers.json holds it.
type Answer struct {
	QuestionID      string          `json:"question_id"`
	Axis            string          `json:"axis,omitempty"`
	SectionID       string          `json:"section_id,omitempty"`
	Kind            string          `json:"kind,omitempty"`
	SelectedOptions []string        `json:"selected_options,omitempty"`
	BooleanChecks   map[string]bool `json:"boolean_checks,omitempty"`
	ExplicitNone    bool            `json:"explicit_none,omitempty"`
	FreeText        string          `json:"free_text,omitempty"`
}

// Answers is the record a submission writes beside the reply, as askhuman's answers.json.
type Answers struct {
	Project     string   `json:"project,omitempty"`
	WorkflowID  string   `json:"workflow_id,omitempty"`
	WorkstepID  string   `json:"workstep_id,omitempty"`
	Form        string   `json:"form,omitempty"`
	Note        string   `json:"note,omitempty"`
	SubmittedAt string   `json:"submitted_at"`
	SubmittedBy string   `json:"submitted_by,omitempty"`
	GeneralNote string   `json:"general_note,omitempty"`
	Answers     []Answer `json:"answers"`
}

// textSuffix names the reason a reader writes beside a qcm.
const textSuffix = ".text"

// Answer returns the record of a bound submission, one answer per question the form shows.
func (f *Form) Answer(s *Submission, values map[string]string, at time.Time, user string) *Answers {
	out := &Answers{
		Project:     f.Project,
		WorkflowID:  f.WorkflowID,
		WorkstepID:  f.WorkstepID,
		Form:        f.ID,
		SubmittedAt: at.UTC().Format(time.RFC3339),
		SubmittedBy: user,
		Answers:     []Answer{},
	}
	for _, field := range f.Visible(values) {
		out.Answers = append(out.Answers, field.answer(s, values))
	}
	return out
}

// answer returns one field's answer from the bound submission.
func (f *Field) answer(s *Submission, values map[string]string) Answer {
	out := Answer{
		QuestionID: f.Key,
		Axis:       f.Axis,
		SectionID:  f.Section,
		Kind:       f.Kind,
		FreeText:   strings.TrimSpace(values[f.Key+textSuffix]),
	}
	switch f.Component {
	case componentChecklist:
		out.BooleanChecks = map[string]bool{}
		for _, option := range f.Options {
			out.BooleanChecks[option.Value] = contains(s.List[f.Key], option.Value)
		}
	case componentMulti:
		out.SelectedOptions = s.List[f.Key]
	default:
		if value := strings.TrimSpace(s.Values[f.Key]); value != "" {
			if f.Kind == kindFreeText && out.FreeText == "" {
				out.FreeText = value
			} else {
				out.SelectedOptions = []string{value}
			}
		}
	}
	if f.Kind == kindQCM && len(out.SelectedOptions) == 0 && out.FreeText == "" && !f.Required {
		out.ExplicitNone = true
	}
	return out
}

// contains reports whether a list holds a value.
func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

// ArtifactPath returns the two files a submission writes under the artifact tree.
func (f *Form) ArtifactPath(note string, index int) (string, string) {
	name := f.ID
	if name == "" {
		name = fmt.Sprintf("ask-%d", index+1)
	}
	base := "ask/" + slug(note) + "/" + slug(name)
	return base + ".answers.json", base + ".decisions.md"
}

// slug returns a name an artifact path takes: lowercase words joined by a dash.
func slug(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('-')
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "ask"
	}
	return out
}

// Markdown returns the decision record a submission writes beside answers.json.
func (a *Answers) Markdown(f *Form) []byte {
	var b strings.Builder
	title := f.Title
	if title == "" {
		title = "Decisions"
	}
	fmt.Fprintf(&b, "# %s\n\n", title)
	if f.Subtitle != "" {
		fmt.Fprintf(&b, "%s\n\n", f.Subtitle)
	}
	if a.Project != "" || a.WorkflowID != "" || a.WorkstepID != "" {
		where := parts("project: "+a.Project, "workflow: "+a.WorkflowID, "step: "+a.WorkstepID)
		fmt.Fprintf(&b, "%s\n\n", strings.Join(where, " · "))
	}
	fmt.Fprintf(&b, "Submitted %s", a.SubmittedAt)
	if a.SubmittedBy != "" {
		fmt.Fprintf(&b, " by %s", a.SubmittedBy)
	}
	b.WriteString(".\n\n")
	if f.Context != "" {
		fmt.Fprintf(&b, "%s\n\n", f.Context)
	}
	for _, line := range f.Prerequisites {
		fmt.Fprintf(&b, "- %s\n", line)
	}
	if len(f.Prerequisites) > 0 {
		b.WriteString("\n")
	}
	byID := map[string]Answer{}
	for _, answer := range a.Answers {
		byID[answer.QuestionID] = answer
	}
	sections := map[string]bool{}
	for _, section := range f.Sections {
		sections[section.ID] = true
		if title := section.Title; title != "" {
			fmt.Fprintf(&b, "## %s\n\n", title)
		}
		if section.Description != "" {
			fmt.Fprintf(&b, "%s\n\n", section.Description)
		}
		for _, field := range f.Fields {
			if field.Section == section.ID {
				b.WriteString(field.decision(byID[field.Key]))
			}
		}
	}
	for _, field := range f.Fields {
		if field.Section == "" {
			b.WriteString(field.decision(byID[field.Key]))
		}
	}
	if a.GeneralNote != "" {
		fmt.Fprintf(&b, "## General note\n\n%s\n", a.GeneralNote)
	}
	return []byte(b.String())
}

// decision returns one question and its answer as the record draws them.
func (f *Field) decision(answer Answer) string {
	var b strings.Builder
	title := f.Label
	if title == "" {
		title = f.Key
	}
	fmt.Fprintf(&b, "### %s\n\n", title)
	if f.Axis != "" {
		fmt.Fprintf(&b, "Axis: %s\n\n", f.Axis)
	}
	if f.Help != "" {
		fmt.Fprintf(&b, "%s\n\n", f.Help)
	}
	if len(answer.BooleanChecks) > 0 {
		for _, option := range f.Options {
			mark := "no"
			if answer.BooleanChecks[option.Value] {
				mark = "yes"
			}
			fmt.Fprintf(&b, "- **%s**: %s\n", option.Label, mark)
		}
		b.WriteString("\n")
	}
	if len(answer.SelectedOptions) > 0 {
		for _, value := range answer.SelectedOptions {
			fmt.Fprintf(&b, "- **%s**", f.labelOf(value))
			if why := f.rationaleOf(value); why != "" {
				fmt.Fprintf(&b, ": %s", why)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	if answer.ExplicitNone {
		b.WriteString("- **Nothing selected.**\n\n")
	}
	if answer.FreeText != "" {
		fmt.Fprintf(&b, "%s\n\n", answer.FreeText)
	}
	return b.String()
}

// labelOf returns an option's label, its value where it names none.
func (f *Field) labelOf(value string) string {
	for _, option := range f.Options {
		if option.Value == value && option.Label != "" {
			return option.Label
		}
	}
	return value
}

// rationaleOf returns an option's rationale, "" where it declares none.
func (f *Field) rationaleOf(value string) string {
	for _, option := range f.Options {
		if option.Value == value {
			return option.Rationale
		}
	}
	return ""
}

// parts returns the non-empty strings of a list.
func parts(list ...string) []string {
	out := []string{}
	for _, item := range list {
		if strings.TrimSpace(item) != "" && !strings.HasSuffix(item, ": ") {
			out = append(out, item)
		}
	}
	return out
}
