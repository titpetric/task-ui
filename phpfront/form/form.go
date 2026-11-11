// Package form is the schema of the forms task-ui declares in yaml: the
// form block of a note's reply, which renders as a form whose Run button
// starts a run, and the form of an admin page. It parses the dialect,
// names the field a bad declaration is in, and binds a submission.
// Nothing here reads a database, a template or a file.
package form

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
	Title  string  `yaml:"title" json:"title,omitempty"`
	Table  string  `yaml:"table" json:"table,omitempty"`
	Fields []Field `yaml:"fields" json:"fields"`
	Submit Submit  `yaml:"submit" json:"submit"`
}

// Field is one input of a form.
type Field struct {
	// Key names the field in a submission and in a prompt.
	Key string `yaml:"key" json:"key"`
	// Label is what the field is called beside its input.
	Label string `yaml:"label" json:"label,omitempty"`
	// Component is the input the renderer draws, one of the components below.
	Component string `yaml:"component" json:"component"`
	// Help is the line under the input.
	Help string `yaml:"help" json:"help,omitempty"`
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
	// Options are the static choices of a choice or a multi.
	Options []Option `yaml:"options" json:"options,omitempty"`
}

// Option is one static choice of a choice or a multi.
type Option struct {
	Value string `yaml:"value" json:"value"`
	Label string `yaml:"label" json:"label,omitempty"`
	Next  string `yaml:"next" json:"next,omitempty"`
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
var components = []string{"text", "textarea", "number", "bool", "hidden", "choice", "multi", "folder", "job", "role", "user", "key", "workspace"}

// The components Parse and Bind treat apart: the two that pick from options and the two that coerce.
const (
	componentChoice = "choice"
	componentMulti  = "multi"
	componentNumber = "number"
	componentBool   = "bool"
)

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
			return nil, errors.New("form: it declares nothing")
		}
		return nil, fmt.Errorf("form: %w", err)
	}
	if err := f.validate(); err != nil {
		return nil, err
	}
	if f.Submit.Label == "" {
		f.Submit.Label = defaultLabel
	}
	return f, nil
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
		return errors.New("form: it declares no field")
	}
	keys := map[string]*Field{}
	for i := range f.Fields {
		field := &f.Fields[i]
		if err := field.validate(i); err != nil {
			return err
		}
		if keys[field.Key] != nil {
			return fmt.Errorf("form: field %q: declared twice", field.Key)
		}
		keys[field.Key] = field
	}
	for i := range f.Fields {
		if err := f.Fields[i].validateRefs(keys); err != nil {
			return err
		}
	}
	return f.validatePrompt(keys)
}

// validate reports what the field itself gets wrong, its index naming one with no usable key.
func (f *Field) validate(i int) error {
	switch {
	case f.Key == "":
		return fmt.Errorf("form: field %d: it has no key", i+1)
	case !keyPattern.MatchString(f.Key):
		return fmt.Errorf("form: field %d: key %q is no [a-z0-9_-] word", i+1, f.Key)
	case f.Component == "":
		return fmt.Errorf("form: field %q: it names no component", f.Key)
	case !slices.Contains(components, f.Component):
		return fmt.Errorf("form: field %q: no component %q; one of %s", f.Key, f.Component, strings.Join(components, ", "))
	}
	return f.validateOptions()
}

// validateOptions reports the options a component may not hold, a source that already supplies them, and a value twice.
func (f *Field) validateOptions() error {
	picks := f.Component == componentChoice || f.Component == componentMulti
	switch {
	case len(f.Options) > 0 && !picks:
		return fmt.Errorf("form: field %q: options are for choice and multi, not %s", f.Key, f.Component)
	case len(f.Options) > 0 && f.Source != "":
		return fmt.Errorf("form: field %q: source %q supplies the options, so it declares none", f.Key, f.Source)
	case picks && len(f.Options) == 0 && f.Source == "":
		return fmt.Errorf("form: field %q: a %s needs options or a source", f.Key, f.Component)
	}
	seen := map[string]bool{}
	for i, option := range f.Options {
		switch {
		case option.Value == "":
			return fmt.Errorf("form: field %q: option %d: it has no value", f.Key, i+1)
		case seen[option.Value]:
			return fmt.Errorf("form: field %q: option %q twice", f.Key, option.Value)
		}
		seen[option.Value] = true
	}
	return nil
}

// validateRefs reports a next or a when of the field that names no field of the form.
func (f *Field) validateRefs(keys map[string]*Field) error {
	for _, option := range f.Options {
		if option.Next != "" && keys[option.Next] == nil {
			return fmt.Errorf("form: field %q: option %q: next names no field %q", f.Key, option.Value, option.Next)
		}
	}
	if f.When == "" {
		return nil
	}
	key, value, ok := strings.Cut(f.When, "=")
	switch {
	case !ok || key == "" || value == "":
		return fmt.Errorf("form: field %q: when %q is no <key>=<value>", f.Key, f.When)
	case keys[key] == nil:
		return fmt.Errorf("form: field %q: when names no field %q", f.Key, key)
	case len(keys[key].Options) > 0 && !keys[key].hasOption(value):
		return fmt.Errorf("form: field %q: when %q is no option of field %q", f.Key, value, key)
	}
	if _, ok := truthOf(value); keys[key].Component == componentBool && !ok {
		return fmt.Errorf("form: field %q: when %q is no yes or no of field %q", f.Key, value, key)
	}
	return nil
}

// validatePrompt reports the first placeholder of the prompt that names no field.
func (f *Form) validatePrompt(keys map[string]*Field) error {
	for _, match := range promptPattern.FindAllStringSubmatch(f.Submit.Prompt, -1) {
		if name := strings.TrimSpace(match[1]); keys[name] == nil {
			return fmt.Errorf("form: submit: prompt %s names no field %q", match[0], name)
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
