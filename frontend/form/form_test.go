package form_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/form"
	"github.com/titpetric/task-ui/internal/assert"
)

// sample is the dialect as the documentation writes it.
const sample = `title: Pick a target
table: run
fields:
  - key: folder
    label: Folder
    component: folder
    required: true
    help: Where it runs
    placeholder: malta/apps
    default: ""
    source: folders
  - key: mode
    label: Mode
    component: choice
    options:
      - {value: plan, label: Plan only}
      - {value: apply, label: Apply, next: confirm}
  - key: hosts
    label: Hosts
    component: multi
    source: hosts
  - key: confirm
    label: Type the folder to confirm
    component: text
    when: mode=apply
submit:
  label: Run
  job: "@claude-prompt"
  folder: malta/apps
  dir: ""
  prompt: "deploy {{folder}} in {{mode}} mode"
`

func TestParse(t *testing.T) {
	t.Parallel()
	f, err := form.Parse([]byte(sample))
	assert.NoError(t, err)
	assert.Equal(t, "Pick a target", f.Title)
	assert.Equal(t, "run", f.Table)
	assert.Len(t, f.Fields, 4)
	assert.Equal(t, form.Field{
		Key:         "folder",
		Label:       "Folder",
		Component:   "folder",
		Help:        "Where it runs",
		Placeholder: "malta/apps",
		Source:      "folders",
		Required:    true,
	}, f.Fields[0])
	assert.Equal(t, []form.Option{
		{Value: "plan", Label: "Plan only"},
		{Value: "apply", Label: "Apply", Next: "confirm"},
	}, f.Fields[1].Options)
	assert.Equal(t, "multi", f.Fields[2].Component)
	assert.Equal(t, "mode=apply", f.Fields[3].When)
	assert.Equal(t, form.Submit{
		Label:  "Run",
		Job:    "@claude-prompt",
		Folder: "malta/apps",
		Prompt: "deploy {{folder}} in {{mode}} mode",
	}, f.Submit)
}

func TestParse_Label(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"no submit":   "fields:\n  - {key: a, component: text}\n",
		"no label":    "fields:\n  - {key: a, component: text}\nsubmit: {job: \"@shell\"}\n",
		"empty label": "fields:\n  - {key: a, component: text}\nsubmit: {label: \"\"}\n",
	} {
		f, err := form.Parse([]byte(source))
		assert.NoError(t, err, name)
		assert.Equal(t, "Run", f.Submit.Label, name)
	}
}

func TestParse_Errors(t *testing.T) {
	t.Parallel()
	for name, test := range map[string]struct {
		source string
		want   string
	}{
		"empty":              {"", "declares nothing"},
		"not yaml":           {"fields: [", "form:"},
		"an unknown key":     {"nope: 1\nfields:\n  - {key: a, component: text}\n", "nope"},
		"an unknown field":   {"fields:\n  - {key: a, component: text, colour: red}\n", "colour"},
		"no fields":          {"title: a\n", "declares no field"},
		"an empty list":      {"fields: []\n", "declares no field"},
		"no key":             {"fields:\n  - {component: text}\n", "field 1: it has no key"},
		"a key in caps":      {"fields:\n  - {key: Folder, component: text}\n", `key "Folder"`},
		"a key with a dot":   {"fields:\n  - {key: a.b, component: text}\n", `key "a.b"`},
		"a key twice":        {"fields:\n  - {key: a, component: text}\n  - {key: a, component: bool}\n", `field "a": declared twice`},
		"no component":       {"fields:\n  - {key: a}\n", `field "a": it names no component`},
		"a bad component":    {"fields:\n  - {key: a, component: slider}\n", `field "a": no component "slider"`},
		"options on a text":  {"fields:\n  - {key: a, component: text, options: [{value: x}]}\n", `field "a": options are for choice and multi`},
		"options and source": {"fields:\n  - {key: a, component: choice, source: modes, options: [{value: x}]}\n", `field "a": source "modes" supplies the options`},
		"a choice with none": {"fields:\n  - {key: a, component: choice}\n", `field "a": a choice needs options or a source`},
		"a multi with none":  {"fields:\n  - {key: a, component: multi}\n", `field "a": a multi needs options or a source`},
		"an option with no value": {"fields:\n  - {key: a, component: choice, options: [{label: x}]}\n",
			`field "a": option 1: it has no value`},
		"an option twice": {"fields:\n  - {key: a, component: choice, options: [{value: x}, {value: x}]}\n",
			`field "a": option "x" twice`},
		"a next to nowhere": {"fields:\n  - {key: a, component: choice, options: [{value: x, next: nope}]}\n",
			`field "a": option "x": next names no field "nope"`},
		"a when with no value": {"fields:\n  - {key: a, component: text}\n  - {key: b, component: text, when: a}\n",
			`field "b": when "a" is no <key>=<value>`},
		"an empty when value": {"fields:\n  - {key: a, component: text}\n  - {key: b, component: text, when: \"a=\"}\n",
			`field "b": when "a=" is no <key>=<value>`},
		"a when to nowhere": {"fields:\n  - {key: a, component: text}\n  - {key: b, component: text, when: nope=1}\n",
			`field "b": when names no field "nope"`},
		"a when off the options": {"fields:\n  - {key: a, component: choice, options: [{value: x}]}\n  - {key: b, component: text, when: a=y}\n",
			`field "b": when "y" is no option of field "a"`},
		"a prompt to nowhere": {"fields:\n  - {key: a, component: text}\nsubmit: {prompt: \"run {{nope}}\"}\n",
			`submit: prompt {{nope}} names no field "nope"`},
	} {
		f, err := form.Parse([]byte(test.source))
		assert.Nil(t, f, name)
		assert.ErrorContains(t, err, test.want, name)
	}
}

func TestForm_Visible(t *testing.T) {
	t.Parallel()
	f, err := form.Parse([]byte("fields:\n" +
		"  - {key: mode, component: choice, options: [{value: plan}, {value: apply, next: confirm}]}\n" +
		"  - {key: confirm, component: choice, when: mode=apply, options: [{value: yes, next: reason}, {value: no}]}\n" +
		"  - {key: reason, component: text, when: confirm=yes}\n"))
	assert.NoError(t, err)

	for name, test := range map[string]struct {
		values map[string]string
		want   []string
	}{
		"nothing picked":          {nil, []string{"mode"}},
		"the first choice":        {map[string]string{"mode": "plan"}, []string{"mode"}},
		"the next choice":         {map[string]string{"mode": "apply"}, []string{"mode", "confirm"}},
		"the whole chain":         {map[string]string{"mode": "apply", "confirm": "yes"}, []string{"mode", "confirm", "reason"}},
		"a chain cut at the root": {map[string]string{"mode": "plan", "confirm": "yes"}, []string{"mode"}},
		"a value with spaces":     {map[string]string{"mode": " apply "}, []string{"mode", "confirm"}},
	} {
		var keys []string
		for _, field := range f.Visible(test.values) {
			keys = append(keys, field.Key)
		}
		assert.Equal(t, test.want, keys, name)
	}
}

// A when over a bool reads both spellings of a truth: a toggle posts 1 or
// 0 and Bind writes true or false.
func TestForm_VisibleBool(t *testing.T) {
	t.Parallel()
	f, err := form.Parse([]byte("fields:\n" +
		"  - {key: deploy, component: bool}\n" +
		"  - {key: reason, component: text, when: deploy=true}\n"))
	assert.NoError(t, err)

	for name, test := range map[string]struct {
		values map[string]string
		want   []string
	}{
		"a toggle that posts 1":   {map[string]string{"deploy": "1"}, []string{"deploy", "reason"}},
		"a toggle that posts on":  {map[string]string{"deploy": "on"}, []string{"deploy", "reason"}},
		"the bound value":         {map[string]string{"deploy": "true"}, []string{"deploy", "reason"}},
		"a toggle that posts 0":   {map[string]string{"deploy": "0"}, []string{"deploy"}},
		"a toggle nobody touched": {nil, []string{"deploy"}},
	} {
		var keys []string
		for _, field := range f.Visible(test.values) {
			keys = append(keys, field.Key)
		}
		assert.Equal(t, test.want, keys, name)
	}

	// A when over a bool that names no truth is a parse error.
	_, err = form.Parse([]byte("fields:\n" +
		"  - {key: deploy, component: bool}\n" +
		"  - {key: reason, component: text, when: deploy=maybe}\n"))
	assert.ErrorContains(t, err, "is no yes or no")
}

func TestForm_Render(t *testing.T) {
	t.Parallel()
	f, err := form.Parse([]byte("fields:\n" +
		"  - {key: folder, component: folder, source: folders}\n" +
		"  - {key: mode, component: choice, options: [{value: apply}]}\n" +
		"  - {key: hosts, component: multi, source: hosts}\n" +
		"submit:\n  prompt: \"deploy {{folder}} in {{ mode }} mode on {{hosts}}, {{folder}} again\"\n"))
	assert.NoError(t, err)

	s, err := f.Bind(map[string]string{"folder": "malta/apps", "mode": "apply", "hosts": "one, two\nthree"})
	assert.NoError(t, err)
	assert.Equal(t, "deploy malta/apps in apply mode on one, two, three, malta/apps again", f.Render(s))
	assert.Equal(t, "deploy  in  mode on ,  again", f.Render(nil))
}
