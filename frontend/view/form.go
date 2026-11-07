package view

import (
	"bytes"
	"html/template"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/titpetric/task-ui/form"
	"github.com/titpetric/task-ui/model"
)

// FormBlock is one form fence of a note's body as the formblock template draws it.
type FormBlock struct {
	// Err is what the fence failed to parse with, which the template draws in the form's place.
	Err string
	// Title is the heading above the fields, empty for a form that names none.
	Title string
	// Label is the one button's text.
	Label string
	// Folder, Dir and Job are what the button starts, as Submit named them.
	Folder string
	Dir    string
	Job    string
	// Prompt is Submit.Prompt as written, the placeholders the script fills from the fields.
	Prompt string
	Fields []FormField
}

// FormField is one input of a form block as the formfield template draws it.
type FormField struct {
	Key         string
	Label       string
	Component   string
	Help        string
	Placeholder string
	// Value is what the input starts on, the field's default.
	Value string
	// When is <key>=<value>, the field and value the script shows this one for.
	When     string
	Required bool
	// Hidden says the field's when is not satisfied by the defaults.
	Hidden bool
	// Checked is a bool field whose default is a yes.
	Checked bool
	Options []FormOption
}

// FormOption is one choice of a field as the formcontrol template draws it.
type FormOption struct {
	Value    string
	Label    string
	Next     string
	Selected bool
}

// formTruths are the values a bool field's default reads as a yes.
var formTruths = []string{"1", "t", "y", "on", "yes", "true"}

// forms is the form templates, parsed once, which a note's body renders a fence with.
var forms = sync.OnceValue(func() *template.Template {
	return template.Must(template.New("_form.tpl").ParseFS(templates, "templates/_form.tpl"))
})

// FormHTML returns the markup of a form fence: the form it declares, or the error it failed to parse with.
func FormHTML(block FormBlock) template.HTML {
	var buf bytes.Buffer
	if err := forms().ExecuteTemplate(&buf, "formblock", block); err != nil {
		return template.HTML("<div class=\"alert bad\" role=\"alert\">" + template.HTMLEscapeString(err.Error()) + "</div>")
	}
	return template.HTML(buf.String())
}

// FormOf returns a form fence as its template draws it, the fields the defaults show and the rest hidden.
func (p *Page) FormOf(block form.Block) FormBlock {
	if block.Err != nil {
		return FormBlock{Err: block.Err.Error()}
	}
	f := block.Form
	out := FormBlock{Title: f.Title, Label: f.Submit.Label, Folder: f.Submit.Folder, Dir: f.Submit.Dir, Job: f.Submit.Job, Prompt: f.Submit.Prompt}
	values := map[string]string{}
	for _, field := range f.Fields {
		values[field.Key] = field.Default
	}
	shown := map[string]bool{}
	for _, field := range f.Visible(values) {
		shown[field.Key] = true
	}
	for _, field := range f.Fields {
		out.Fields = append(out.Fields, p.formField(field, !shown[field.Key]))
	}
	return out
}

// formField returns one declared field as its template draws it, its options from the source it names.
func (p *Page) formField(field form.Field, hidden bool) FormField {
	out := FormField{
		Key:         field.Key,
		Label:       field.Label,
		Component:   field.Component,
		Help:        field.Help,
		Placeholder: field.Placeholder,
		Value:       field.Default,
		When:        field.When,
		Required:    field.Required,
		Hidden:      hidden,
		Checked:     slices.Contains(formTruths, strings.ToLower(strings.TrimSpace(field.Default))),
	}
	if out.Label == "" {
		out.Label = field.Key
	}
	for _, option := range field.Options {
		label := option.Label
		if label == "" {
			label = option.Value
		}
		out.Options = append(out.Options, FormOption{Value: option.Value, Label: label, Next: option.Next, Selected: option.Value == field.Default})
	}
	if len(out.Options) == 0 {
		out.Options = p.FormSource(FormSourceOf(field))
		for i := range out.Options {
			out.Options[i].Selected = out.Options[i].Value == field.Default
		}
	}
	return out
}

// FormSourceOf names the source a field's options come from: the one it declares, else its own picker component.
func FormSourceOf(field form.Field) string {
	if field.Source != "" {
		return field.Source
	}
	switch field.Component {
	case "folder", "job", "role", "user", "key", "workspace":
		return field.Component
	}
	return ""
}

// FormSource returns the options a source names, the one lookup point of a form's source; nil for one the page does not hold.
func (p *Page) FormSource(source string) []FormOption {
	switch source {
	case "folder":
		out := make([]FormOption, 0, len(p.Folders))
		for _, f := range p.Folders {
			out = append(out, FormOption{Value: f.ID, Label: f.ID})
		}
		return out
	case "role":
		out := make([]FormOption, 0, len(model.Roles()))
		for _, role := range model.Roles() {
			out = append(out, FormOption{Value: string(role), Label: string(role)})
		}
		return out
	case "user":
		names := p.UserNames()
		out := make([]FormOption, 0, len(names))
		for _, name := range names {
			out = append(out, FormOption{Value: name, Label: name})
		}
		return out
	case "workspace":
		out := make([]FormOption, 0, len(p.Workspaces))
		for folderID := range p.Workspaces {
			out = append(out, FormOption{Value: folderID, Label: folderID})
		}
		sort.Slice(out, func(i, j int) bool { return p.Workspaces[out[i].Value] < p.Workspaces[out[j].Value] })
		return out
	}
	return nil
}

