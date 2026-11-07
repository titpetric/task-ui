package view_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/form"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// deployForm is a form fence a reply holds: a choice that reveals a field,
// and a prompt over both.
const deployForm = "```form\n" +
	"title: Deploy\n" +
	"fields:\n" +
	"  - key: env\n" +
	"    label: Environment\n" +
	"    component: choice\n" +
	"    help: Where it goes\n" +
	"    required: true\n" +
	"    options:\n" +
	"      - value: prod\n" +
	"        label: Production\n" +
	"        next: ticket\n" +
	"      - value: stage\n" +
	"        label: Staging\n" +
	"  - key: ticket\n" +
	"    label: Change ticket\n" +
	"    component: text\n" +
	"    when: env=prod\n" +
	"    placeholder: OPS-1\n" +
	"submit:\n" +
	"  label: Deploy it\n" +
	"  folder: infra\n" +
	"  job: deploy\n" +
	"  prompt: Deploy {{env}}, ticket {{ticket}}.\n" +
	"```"

// blockOf parses one fence of a body and fails the test when the body holds no form.
func blockOf(t *testing.T, body string) form.Block {
	t.Helper()
	blocks := form.Blocks(body)
	assert.Len(t, blocks, 1)
	return blocks[0]
}

// A body with one form renders as markdown, the form, then markdown: the
// prose around the fence is a paragraph each and the yaml is drawn nowhere.
func TestPage_MessageOneForm(t *testing.T) {
	t.Parallel()
	out := string(admin(t).Message("Pick one:\n\n" + deployForm + "\n\nIt runs as you."))
	before, rest, ok := strings.Cut(out, "<form")
	assert.True(t, ok, "a form")
	assert.Equal(t, "<p>Pick one:</p>\n", before)
	assert.Contains(t, rest, "<p>It runs as you.</p>\n")
	assert.NotContains(t, out, "component: choice")
	assert.NotContains(t, out, "<pre>")
}

// A body with two forms renders both, in the order it declares them, with
// the markdown between them between them.
func TestPage_MessageTwoForms(t *testing.T) {
	t.Parallel()
	second := strings.ReplaceAll(deployForm, "Deploy", "Roll back")
	out := string(admin(t).Message(deployForm + "\n\nor else\n\n" + second))
	assert.Equal(t, 2, strings.Count(out, "<form class=\"note-form run-form\""))
	first := strings.Index(out, "Deploy it</button>")
	middle := strings.Index(out, "<p>or else</p>")
	last := strings.Index(out, "Roll back it</button>")
	assert.True(t, first > 0 && first < middle && middle < last, "the first form, the prose, the second")
}

// A body with no form renders as it did before: one markdown document,
// and a fence that is no form stays a code block.
func TestPage_MessageNoForm(t *testing.T) {
	t.Parallel()
	p := admin(t)
	assert.Equal(t, "<p>plain <span class=\"mention\">@ops</span></p>\n", string(p.Message("plain @ops")))
	out := string(p.Message("Done:\n```yaml\nkey: value\n```"))
	assert.Contains(t, out, "<pre><code class=\"language-yaml\">key: value\n</code></pre>")
	assert.NotContains(t, out, "<form")
}

// A fence that fails to parse draws the error where the form would have
// been, never the yaml and never nothing.
func TestPage_MessageBadForm(t *testing.T) {
	t.Parallel()
	out := string(admin(t).Message("try:\n\n```form\nfields:\n  - key: env\n```\n"))
	assert.Contains(t, out, `<div class="alert bad" role="alert">form: field &#34;env&#34;: it names no component</div>`)
	assert.NotContains(t, out, "<form")
	assert.NotContains(t, out, "key: env")

	out = string(admin(t).Message("```form\nfields: [\n```\n"))
	assert.Contains(t, out, `<div class="alert bad" role="alert">form: `)
}

// A field whose when the defaults do not satisfy starts hidden, and the
// rule is Visible's: the same fields are drawn and shown.
func TestPage_FormOf(t *testing.T) {
	t.Parallel()
	block := blockOf(t, deployForm)
	drawn := admin(t).FormOf(block)
	assert.Equal(t, "Deploy", drawn.Title)
	assert.Equal(t, "Deploy it", drawn.Label)
	assert.Equal(t, "infra", drawn.Folder)
	assert.Equal(t, "deploy", drawn.Job)
	assert.Equal(t, "Deploy {{env}}, ticket {{ticket}}.", drawn.Prompt)
	assert.Len(t, drawn.Fields, 2)
	assert.False(t, drawn.Fields[0].Hidden)
	assert.True(t, drawn.Fields[1].Hidden, "no default picks prod")
	assert.Len(t, block.Form.Visible(map[string]string{}), 1)

	// A default that satisfies the when shows the field it names, as
	// Visible shows it.
	with := blockOf(t, strings.Replace(deployForm, "    component: choice\n", "    component: choice\n    default: prod\n", 1))
	drawn = admin(t).FormOf(with)
	assert.False(t, drawn.Fields[1].Hidden)
	assert.Len(t, with.Form.Visible(map[string]string{"env": "prod"}), 2)
	assert.True(t, drawn.Fields[0].Options[0].Selected)
}

// A block that failed to parse holds its error and no field.
func TestFormBlock(t *testing.T) {
	t.Parallel()
	drawn := admin(t).FormOf(blockOf(t, "```form\n```\n"))
	assert.Equal(t, "form: it declares nothing", drawn.Err)
	assert.Len(t, drawn.Fields, 0)
}

// Each component draws the control the theme has for it, the field
// naming its key, its component and its when.
func TestFormField(t *testing.T) {
	t.Parallel()
	body := "```form\nfields:\n" +
		"  - {key: name, label: Name, component: text, placeholder: web, required: true}\n" +
		"  - {key: why, label: Why, component: textarea, default: because}\n" +
		"  - {key: count, label: Count, component: number, default: \"3\"}\n" +
		"  - {key: force, label: Force, component: bool, default: \"yes\"}\n" +
		"  - {key: batch, label: Batch, component: hidden, default: b1}\n" +
		"  - {key: env, label: Env, component: choice, options: [{value: prod, label: Production, next: why}]}\n" +
		"  - {key: tags, label: Tags, component: multi, options: [{value: a}, {value: b}]}\n" +
		"  - {key: place, label: Place, component: folder}\n" +
		"  - {key: who, label: Who, component: role}\n" +
		"submit:\n  prompt: go {{name}}\n```"
	out := string(view.FormHTML(admin(t).FormOf(blockOf(t, body))))
	assert.Contains(t, out, `<div class="field" data-key="name" data-component="text" data-required>`)
	assert.Contains(t, out, `<input type="text" name="name" value="" aria-label="Name" placeholder="web" required>`)
	assert.Contains(t, out, `<textarea name="why" rows="4" spellcheck="false" aria-label="Why">because</textarea>`)
	assert.Contains(t, out, `<input type="number" name="count" value="3" aria-label="Count">`)
	assert.Contains(t, out, `<label class="check"><input type="checkbox" name="force" data-type="toggle" checked> Force</label>`)
	assert.Contains(t, out, `<input type="hidden" name="batch" value="b1" data-key="batch" data-component="hidden">`)
	assert.Contains(t, out, `<select name="env" aria-label="Env"><option value=""></option><option value="prod" data-next="why">Production</option></select>`)
	assert.Contains(t, out, `<label class="check"><input type="checkbox" name="tags" value="a" data-type="set"> a</label>`)
	assert.Contains(t, out, `<select name="place" aria-label="Place"><option value=""></option><option value="infra">infra</option></select>`)
	assert.Contains(t, out, `<option value="admin">admin</option><option value="runner">runner</option><option value="viewer">viewer</option>`)
	assert.Contains(t, out, `<div class="form-foot"><span class="faint form-state"></span><button class="btn btn-primary" type="submit">Run</button></div>`)
	assert.NotContains(t, out, `data-component="hidden" hidden`)
}

// An option with no label of its own reads as its value, and the one the
// default names is selected.
func TestFormOption(t *testing.T) {
	t.Parallel()
	body := "```form\nfields:\n  - {key: env, component: choice, default: stage, options: [{value: prod, label: Production}, {value: stage}]}\nsubmit: {prompt: \"{{env}}\"}\n```"
	drawn := admin(t).FormOf(blockOf(t, body))
	assert.Equal(t, []view.FormOption{{Value: "prod", Label: "Production"}, {Value: "stage", Label: "stage", Selected: true}}, drawn.Fields[0].Options)
	assert.Equal(t, "env", drawn.Fields[0].Label, "a field with no label reads as its key")
}

// A form fence renders as one form, the markup of the components the theme
// already has, with the submit on the button.
func TestFormHTML(t *testing.T) {
	t.Parallel()
	out := string(view.FormHTML(admin(t).FormOf(blockOf(t, deployForm))))
	assert.Equal(t, `<form class="note-form run-form" data-run-form data-folder="infra" data-job="deploy" data-prompt="Deploy {{env}}, ticket {{ticket}}.">
<p class="form-title">Deploy</p>
<div class="field" data-key="env" data-component="choice" data-required>
<span class="label">Environment</span>
<select name="env" aria-label="Environment" required><option value="prod" data-next="ticket">Production</option><option value="stage">Staging</option></select>
<small class="faint">Where it goes</small>
</div>
<div class="field" data-key="ticket" data-component="text" data-when="env=prod" hidden>
<span class="label">Change ticket</span>
<input type="text" name="ticket" value="" aria-label="Change ticket" placeholder="OPS-1">
</div>
<div class="form-foot"><span class="faint form-state"></span><button class="btn btn-primary" type="submit">Deploy it</button></div>
</form>`, out)
}

// A source names where a field's options come from: the one it declares,
// else the picker component it is.
func TestFormSourceOf(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "role", view.FormSourceOf(form.Field{Component: "choice", Source: "role"}))
	assert.Equal(t, "folder", view.FormSourceOf(form.Field{Component: "folder"}))
	assert.Equal(t, "workspace", view.FormSourceOf(form.Field{Component: "workspace"}))
	assert.Equal(t, "", view.FormSourceOf(form.Field{Component: "text"}))
}

// The sources the page holds: the folders it sees, the roles, the users
// and the workspaces in the sidebar's order. A source it does not hold
// has no options.
func TestPage_FormSource(t *testing.T) {
	t.Parallel()
	p := admin(t)
	p.Roles = map[string]string{"ops": model.RoleRunner, "admin": model.RoleAdmin}
	p.Workspaces = map[string]int{"malta": 1, "internal": 0}
	assert.Equal(t, []view.FormOption{{Value: "infra", Label: "infra"}}, p.FormSource("folder"))
	assert.Equal(t, []view.FormOption{{Value: "admin", Label: "admin"}, {Value: "runner", Label: "runner"}, {Value: "viewer", Label: "viewer"}}, p.FormSource("role"))
	assert.Equal(t, []view.FormOption{{Value: "admin", Label: "admin"}, {Value: "ops", Label: "ops"}}, p.FormSource("user"))
	assert.Equal(t, []view.FormOption{{Value: "internal", Label: "internal"}, {Value: "malta", Label: "malta"}}, p.FormSource("workspace"))
	assert.Nil(t, p.FormSource("job"))
	assert.Nil(t, p.FormSource("key"))
	assert.Nil(t, p.FormSource(""))
}
