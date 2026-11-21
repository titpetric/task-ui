// Package askhuman is the questionnaire protocol of a note's reply: an
// ```ask fence of yaml, the questions it asks and the record a submission
// leaves. It parses the dialect, names the field a bad declaration is in,
// binds a submission, and writes the answers and decisions an agent resumes on.
// Nothing here reads a database or a template.
package askhuman

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// Form is a form: what it is called, the table it edits, its fields and its submission.
type Form struct {
	// ID names the form in an artifact path, and defaults to the block's place in the body.
	ID string `yaml:"id" json:"id,omitempty"`
	// Project, WorkflowID and WorkstepID name where the questions sit, as askhuman's spec does.
	Project    string `yaml:"project" json:"project,omitempty"`
	WorkflowID string `yaml:"workflow_id" json:"workflow_id,omitempty"`
	WorkstepID string `yaml:"workstep_id" json:"workstep_id,omitempty"`
	Title      string `yaml:"title" json:"title,omitempty"`
	Subtitle   string `yaml:"subtitle" json:"subtitle,omitempty"`
	// Context is the paragraph above the questions, from context or context_summary.
	Context string `yaml:"context" json:"context,omitempty"`
	// ContextSummary is the askhuman spelling of Context.
	ContextSummary string `yaml:"context_summary" json:"context_summary,omitempty"`
	// Prerequisites are the facts the reader should hold before answering.
	Prerequisites []string  `yaml:"prerequisites" json:"prerequisites,omitempty"`
	Sections      []Section `yaml:"sections" json:"sections,omitempty"`
	Table         string    `yaml:"table" json:"table,omitempty"`
	Fields        []Field   `yaml:"fields" json:"fields"`
	// Questions is the askhuman spelling of Fields.
	Questions []Field `yaml:"questions" json:"questions,omitempty"`
	Submit    Submit  `yaml:"submit" json:"submit"`
}

// Section groups the fields of a long form under a title.
type Section struct {
	ID          string `yaml:"id" json:"id"`
	Title       string `yaml:"title" json:"title,omitempty"`
	Description string `yaml:"description" json:"description,omitempty"`
}

// Field is one input of a form.
type Field struct {
	// Key names the field in a submission and in a prompt; ID is the askhuman spelling.
	Key string `yaml:"key" json:"key"`
	ID  string `yaml:"id" json:"id,omitempty"`
	// Label is what the field is called beside its input; Title is the askhuman spelling.
	Label string `yaml:"label" json:"label,omitempty"`
	Title string `yaml:"title" json:"title,omitempty"`
	// Component is the input the renderer draws, one of the components below.
	Component string `yaml:"component" json:"component"`
	// Kind is an askhuman question kind, naming the component when it says none.
	Kind string `yaml:"kind" json:"kind,omitempty"`
	// Axis names what the question settles, which the decision record keeps.
	Axis string `yaml:"axis" json:"axis,omitempty"`
	// Section names the section the field draws under.
	Section string `yaml:"section" json:"section,omitempty"`
	// Help is the line under the input; Context is the askhuman spelling.
	Help    string `yaml:"help" json:"help,omitempty"`
	Context string `yaml:"context" json:"context,omitempty"`
	// Placeholder is the empty input's text.
	Placeholder string `yaml:"placeholder" json:"placeholder,omitempty"`
	// Default is the value an empty submission takes.
	Default string `yaml:"default" json:"default,omitempty"`
	// Source names the DAO getter the renderer asks for the options.
	Source string `yaml:"source" json:"source,omitempty"`
	// When is <key>=<value>, the field and value that show this one.
	When string `yaml:"when" json:"when,omitempty"`
	// Required refuses an empty submission.
	Required bool `yaml:"required" json:"required,omitempty"`
	// MultiSelect is the askhuman spelling of a multi.
	MultiSelect bool `yaml:"multi_select" json:"multi_select,omitempty"`
	// FreeText asks a reason beside a qcm, posted as <key>.text.
	FreeText            bool   `yaml:"free_text" json:"free_text,omitempty"`
	FreeTextLabel       string `yaml:"free_text_label" json:"free_text_label,omitempty"`
	FreeTextPlaceholder string `yaml:"free_text_placeholder" json:"free_text_placeholder,omitempty"`
	MaxFreeTextBytes    int    `yaml:"max_free_text_bytes" json:"max_free_text_bytes,omitempty"`
	// Options are the static choices of a choice, a multi or a checklist.
	Options []Option `yaml:"options" json:"options,omitempty"`
}

// Option is one static choice of a choice, a multi or a checklist.
type Option struct {
	// Value is the option's ID; ID is the askhuman spelling.
	Value string `yaml:"value" json:"value"`
	ID    string `yaml:"id" json:"id,omitempty"`
	// Label is what the option is called; Title is the askhuman spelling.
	Label string `yaml:"label" json:"label,omitempty"`
	Title string `yaml:"title" json:"title,omitempty"`
	// Rationale is why the option is there, drawn under its label.
	Rationale string `yaml:"rationale" json:"rationale,omitempty"`
	// Recommended marks the option the form's author suggests.
	Recommended bool `yaml:"recommended" json:"recommended,omitempty"`
	// DefaultChecked is the option a submission names when the reader changes nothing.
	DefaultChecked bool   `yaml:"default_checked" json:"default_checked,omitempty"`
	Next           string `yaml:"next" json:"next,omitempty"`
}

// Submit is what the form's one button does: the run it starts.
type Submit struct {
	Label  string `yaml:"label" json:"label,omitempty"`
	Job    string `yaml:"job" json:"job,omitempty"`
	Folder string `yaml:"folder" json:"folder,omitempty"`
	Dir    string `yaml:"dir" json:"dir,omitempty"`
	Prompt string `yaml:"prompt" json:"prompt,omitempty"`
}

// components are the inputs the dialect names; a field names one of them.
var components = []string{"text", "textarea", "number", "bool", "hidden", "choice", "multi", "checklist", "folder", "job", "role", "user", "key", "workspace"}

// The components Parse and Bind treat apart: the three that pick from options and the two that coerce.
const (
	componentChoice    = "choice"
	componentMulti     = "multi"
	componentChecklist = "checklist"
	componentNumber    = "number"
	componentBool      = "bool"
)

// The question kinds of askhuman's spec, each naming a component when the field names none.
const (
	kindFreeText  = "free_text"
	kindQCM       = "qcm_free_text"
	kindChecklist = "boolean_checklist"
)

// kinds are the question kinds a field may name.
var kinds = []string{kindFreeText, kindQCM, kindChecklist}

// defaultFreeTextBytes caps the reason beside a qcm, as askhuman's spec does.
const defaultFreeTextBytes = 32768

// defaultLabel is the button's label a form that names none takes.
const defaultLabel = "Run"

// strictDecoder is what Parse asks of the yaml decoder: a document, and an unknown key refused.
type strictDecoder interface {
	KnownFields(bool)
	Decode(any) error
}

var keyPattern = regexp.MustCompile(`^[a-z0-9_-]+$`)

var promptPattern = regexp.MustCompile(`\{\{([^{}]*)\}\}`)

// Parse returns the form a yaml document declares, with an error that names the field a bad declaration is in.
func Parse(data []byte) (*Form, error) {
	var decoder strictDecoder = yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	f := &Form{}
	if err := decoder.Decode(f); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("ask: it declares nothing")
		}
		return nil, fmt.Errorf("ask: %w", err)
	}
	if err := f.normalize(); err != nil {
		return nil, err
	}
	if err := f.validate(); err != nil {
		return nil, err
	}
	if f.Submit.Label == "" {
		f.Submit.Label = defaultLabel
	}
	return f, nil
}

// normalize fills the askhuman spellings from the task-ui ones, refusing a form that declares both.
func (f *Form) normalize() error {
	if len(f.Fields) == 0 {
		f.Fields = f.Questions
	} else if len(f.Questions) > 0 {
		return errors.New("ask: it declares fields and questions")
	}
	if f.Context == "" {
		f.Context = f.ContextSummary
	}
	for i := range f.Fields {
		f.Fields[i].normalize()
	}
	return nil
}

// normalize fills one field's askhuman aliases, the component its kind names and its default option.
func (f *Field) normalize() {
	if f.Key == "" {
		f.Key = f.ID
	}
	if f.Label == "" {
		f.Label = f.Title
	}
	if f.Help == "" {
		f.Help = f.Context
	}
	if f.Component == "" {
		f.Component = componentOf(f.Kind, f.MultiSelect)
	}
	if f.FreeText && f.MaxFreeTextBytes == 0 {
		f.MaxFreeTextBytes = defaultFreeTextBytes
	}
	for j := range f.Options {
		option := &f.Options[j]
		if option.Value == "" {
			option.Value = option.ID
		}
		if option.Label == "" {
			option.Label = option.Title
		}
		if option.DefaultChecked && f.Default == "" && !f.MultiSelect && f.Component == componentChoice {
			f.Default = option.Value
		}
	}
}

// componentOf returns the component a question kind draws, "" for a kind nobody knows.
func componentOf(kind string, multi bool) string {
	switch kind {
	case kindFreeText:
		return "textarea"
	case kindQCM:
		if multi {
			return componentMulti
		}
		return componentChoice
	case kindChecklist:
		return componentChecklist
	}
	return ""
}

// Visible returns the fields the values show in declaration order; a field whose own when-field is hidden stays hidden.
func (f *Form) Visible(values map[string]string) []Field {
	var list []Field
	shown := map[string]bool{}
	by := map[string]*Field{}
	for i := range f.Fields {
		by[f.Fields[i].Key] = &f.Fields[i]
	}
	for _, field := range f.Fields {
		if field.When != "" {
			key, value, _ := strings.Cut(field.When, "=")
			if !shown[key] || !holds(by[key], values[key], value) {
				continue
			}
		}
		shown[field.Key] = true
		list = append(list, field)
	}
	return list
}

// holds reports whether the posted value of a field is the one When names.
// A bool is read as a truth on both sides: a toggle posts 1 or 0, Bind writes
// true or false, and either spelling is accepted.
func holds(on *Field, posted, want string) bool {
	posted = strings.TrimSpace(posted)
	if on != nil && on.Component == componentBool {
		got, ok := truthOf(posted)
		wanted, fine := truthOf(want)
		return ok && fine && got == wanted
	}
	return posted == want
}

// Render returns Submit.Prompt with every placeholder replaced by the submitted value, "" for a field with none.
func (f *Form) Render(s *Submission) string {
	values := map[string]string{}
	if s != nil {
		values = s.Values
	}
	return promptPattern.ReplaceAllStringFunc(f.Submit.Prompt, func(match string) string {
		return values[strings.TrimSpace(strings.Trim(match, "{}"))]
	})
}

// validate reports the first field the dialect refuses, the references and the prompt last.
func (f *Form) validate() error {
	if len(f.Fields) == 0 {
		return errors.New("ask: it declares no field")
	}
	if err := f.validateSections(); err != nil {
		return err
	}
	keys := map[string]*Field{}
	for i := range f.Fields {
		field := &f.Fields[i]
		if err := field.validate(i); err != nil {
			return err
		}
		if keys[field.Key] != nil {
			return fmt.Errorf("ask: field %q: declared twice", field.Key)
		}
		keys[field.Key] = field
	}
	for i := range f.Fields {
		if err := f.Fields[i].validateRefs(keys); err != nil {
			return err
		}
	}
	if err := f.validateSectionRefs(); err != nil {
		return err
	}
	return f.validatePrompt(keys)
}

// validateSections reports a section with no ID or one declared twice.
func (f *Form) validateSections() error {
	seen := map[string]bool{}
	for i, section := range f.Sections {
		switch {
		case section.ID == "":
			return fmt.Errorf("ask: section %d: it has no id", i+1)
		case seen[section.ID]:
			return fmt.Errorf("ask: section %q: declared twice", section.ID)
		}
		seen[section.ID] = true
	}
	return nil
}

// validateSectionRefs reports a field drawn under a section nobody declared.
func (f *Form) validateSectionRefs() error {
	seen := map[string]bool{}
	for _, section := range f.Sections {
		seen[section.ID] = true
	}
	for _, field := range f.Fields {
		if field.Section != "" && !seen[field.Section] {
			return fmt.Errorf("ask: field %q: section %q is not declared", field.Key, field.Section)
		}
	}
	return nil
}

// validate reports what the field itself gets wrong, its index naming one with no usable key.
func (f *Field) validate(i int) error {
	switch {
	case f.Key == "":
		return fmt.Errorf("ask: field %d: it has no key", i+1)
	case !keyPattern.MatchString(f.Key):
		return fmt.Errorf("ask: field %d: key %q is no [a-z0-9_-] word", i+1, f.Key)
	case f.Kind != "" && !slices.Contains(kinds, f.Kind):
		return fmt.Errorf("ask: field %q: no kind %q; one of %s", f.Key, f.Kind, strings.Join(kinds, ", "))
	case f.Component == "":
		return fmt.Errorf("ask: field %q: it names no component", f.Key)
	case !slices.Contains(components, f.Component):
		return fmt.Errorf("ask: field %q: no component %q; one of %s", f.Key, f.Component, strings.Join(components, ", "))
	case f.FreeText && f.Component != componentChoice && f.Component != componentMulti:
		return fmt.Errorf("ask: field %q: free text follows a choice or a multi, not %s", f.Key, f.Component)
	case f.MaxFreeTextBytes < 0:
		return fmt.Errorf("ask: field %q: max_free_text_bytes is no count", f.Key)
	}
	return f.validateOptions()
}

// validateOptions reports the options a component may not hold, a source that already supplies them, and a value twice.
func (f *Field) validateOptions() error {
	picks := f.Component == componentChoice || f.Component == componentMulti || f.Component == componentChecklist
	switch {
	case len(f.Options) > 0 && !picks:
		return fmt.Errorf("ask: field %q: options are for choice, multi and checklist, not %s", f.Key, f.Component)
	case len(f.Options) > 0 && f.Source != "":
		return fmt.Errorf("ask: field %q: source %q supplies the options, so it declares none", f.Key, f.Source)
	case picks && len(f.Options) == 0 && f.Source == "":
		return fmt.Errorf("ask: field %q: a %s needs options or a source", f.Key, f.Component)
	case f.Kind == kindQCM && len(f.Options) < 2 && f.Source == "":
		return fmt.Errorf("ask: field %q: a qcm_free_text needs two options", f.Key)
	}
	seen := map[string]bool{}
	for i, option := range f.Options {
		switch {
		case option.Value == "":
			return fmt.Errorf("ask: field %q: option %d: it has no value", f.Key, i+1)
		case seen[option.Value]:
			return fmt.Errorf("ask: field %q: option %q twice", f.Key, option.Value)
		}
		seen[option.Value] = true
	}
	return nil
}

// validateRefs reports a next or a when of the field that names no field of the form.
func (f *Field) validateRefs(keys map[string]*Field) error {
	for _, option := range f.Options {
		if option.Next != "" && keys[option.Next] == nil {
			return fmt.Errorf("ask: field %q: option %q: next names no field %q", f.Key, option.Value, option.Next)
		}
	}
	if f.When == "" {
		return nil
	}
	key, value, ok := strings.Cut(f.When, "=")
	switch {
	case !ok || key == "" || value == "":
		return fmt.Errorf("ask: field %q: when %q is no <key>=<value>", f.Key, f.When)
	case keys[key] == nil:
		return fmt.Errorf("ask: field %q: when names no field %q", f.Key, key)
	case len(keys[key].Options) > 0 && !keys[key].hasOption(value):
		return fmt.Errorf("ask: field %q: when %q is no option of field %q", f.Key, value, key)
	}
	if _, ok := truthOf(value); keys[key].Component == componentBool && !ok {
		return fmt.Errorf("ask: field %q: when %q is no yes or no of field %q", f.Key, value, key)
	}
	return nil
}

// validatePrompt reports the first placeholder of the prompt that names no field.
func (f *Form) validatePrompt(keys map[string]*Field) error {
	for _, match := range promptPattern.FindAllStringSubmatch(f.Submit.Prompt, -1) {
		if name := strings.TrimSpace(match[1]); keys[name] == nil {
			return fmt.Errorf("ask: submit: prompt %s names no field %q", match[0], name)
		}
	}
	return nil
}

// hasOption reports whether a value is one of the field's static options.
func (f *Field) hasOption(value string) bool {
	for _, option := range f.Options {
		if option.Value == value {
			return true
		}
	}
	return false
}
