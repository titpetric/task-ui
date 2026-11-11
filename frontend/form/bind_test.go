package form_test

import (
	"testing"

	"github.com/titpetric/task-ui/form"
	"github.com/titpetric/task-ui/internal/assert"
)

// bindForm is a form of one field per thing Bind does to a value.
func bindForm() *form.Form {
	return &form.Form{
		Title: "Deploy",
		Fields: []form.Field{
			{Key: "folder", Label: "Folder", Component: "folder", Source: "folders", Required: true},
			{Key: "count", Label: "Count", Component: "number", Default: "1"},
			{Key: "dry", Label: "Dry run", Component: "bool"},
			{Key: "mode", Label: "Mode", Component: "choice", Options: []form.Option{{Value: "plan"}, {Value: "apply"}}},
			{Key: "team", Label: "Team", Component: "choice", Source: "teams"},
			{Key: "hosts", Label: "Hosts", Component: "multi", Source: "hosts"},
			{Key: "note", Label: "Note", Component: "textarea"},
		},
		Submit: form.Submit{Label: "Run"},
	}
}

func TestForm_Bind(t *testing.T) {
	t.Parallel()
	f := bindForm()
	s, err := f.Bind(map[string]string{
		"folder": "  malta/apps  ",
		"dry":    "on",
		"mode":   "apply",
		"team":   "anything the renderer listed",
		"hosts":  " a, b\nc ",
		"note":   " hi ",
		"gone":   "a value of no field",
	})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{
		"folder": "malta/apps",
		"count":  "1",
		"dry":    "true",
		"mode":   "apply",
		"team":   "anything the renderer listed",
		"hosts":  "a, b, c",
		"note":   "hi",
	}, s.Values)
	assert.Equal(t, map[string][]string{"hosts": {"a", "b", "c"}}, s.List)
	assert.Equal(t, map[string]string{}, s.Errors)
}

func TestForm_Bind_Coercion(t *testing.T) {
	t.Parallel()
	f := bindForm()
	for name, test := range map[string]struct {
		key   string
		value string
		want  string
	}{
		"a number with spaces": {"count", "  7 ", "7"},
		"a number signed":      {"count", "+3", "3"},
		"a number trailing":    {"count", "2.50", "2.5"},
		"a number zero":        {"count", "0", "0"},
		"a number empty":       {"count", "", "1"},
		"a box unticked":       {"dry", "", "false"},
		"a box ticked":         {"dry", "on", "true"},
		"a box as one":         {"dry", "1", "true"},
		"a box in caps":        {"dry", "YES", "true"},
		"a box as no":          {"dry", "no", "false"},
		"a multi with commas":  {"hosts", "a,b", "a, b"},
		"a multi with breaks":  {"hosts", "a\nb", "a, b"},
		"a multi half empty":   {"hosts", ",, a ,,b,", "a, b"},
		"a multi empty":        {"hosts", "", ""},
		"a textarea trimmed":   {"note", "  two words  ", "two words"},
	} {
		s, err := f.Bind(map[string]string{"folder": "malta/apps", test.key: test.value})
		assert.NoError(t, err, name)
		assert.Equal(t, test.want, s.Values[test.key], name)
	}
}

func TestForm_Bind_Errors(t *testing.T) {
	t.Parallel()
	f := bindForm()
	for name, test := range map[string]struct {
		values map[string]string
		want   map[string]string
	}{
		"a required field empty": {
			map[string]string{"folder": "   "},
			map[string]string{"folder": "a value is needed"},
		},
		"a required field missing": {
			nil,
			map[string]string{"folder": "a value is needed"},
		},
		"a number that is none": {
			map[string]string{"folder": "f", "count": "ten"},
			map[string]string{"count": `"ten" is no number`},
		},
		"a box that is neither": {
			map[string]string{"folder": "f", "dry": "maybe"},
			map[string]string{"dry": `"maybe" is no yes or no`},
		},
		"a choice off its options": {
			map[string]string{"folder": "f", "mode": "destroy"},
			map[string]string{"mode": `"destroy" is no option of it`},
		},
		"two fields at once": {
			map[string]string{"count": "ten"},
			map[string]string{"folder": "a value is needed", "count": `"ten" is no number`},
		},
	} {
		s, err := f.Bind(test.values)
		assert.ErrorIs(t, err, form.ErrInvalid, name)
		assert.Equal(t, test.want, s.Errors, name)
		for key, message := range test.want {
			assert.Contains(t, err.Error(), key+": "+message, name)
			assert.Equal(t, "", s.Values[key], name)
		}
	}
}

func TestForm_Bind_Options(t *testing.T) {
	t.Parallel()
	f := &form.Form{Fields: []form.Field{
		{Key: "hosts", Component: "multi", Required: true, Options: []form.Option{{Value: "one"}, {Value: "two"}}},
	}}
	s, err := f.Bind(map[string]string{"hosts": "two,one"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"two", "one"}, s.List["hosts"])

	s, err = f.Bind(map[string]string{"hosts": "one,three"})
	assert.ErrorIs(t, err, form.ErrInvalid)
	assert.Equal(t, `"three" is no option of it`, s.Errors["hosts"])

	s, err = f.Bind(nil)
	assert.ErrorIs(t, err, form.ErrInvalid)
	assert.Equal(t, "a value is needed", s.Errors["hosts"])
}

func TestForm_Bind_Hidden(t *testing.T) {
	t.Parallel()
	f, err := form.Parse([]byte("fields:\n" +
		"  - {key: mode, component: choice, options: [{value: plan}, {value: apply, next: confirm}]}\n" +
		"  - {key: confirm, component: text, when: mode=apply, required: true}\n"))
	assert.NoError(t, err)

	s, err := f.Bind(map[string]string{"mode": "plan", "confirm": "typed anyway"})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"mode": "plan"}, s.Values)

	s, err = f.Bind(map[string]string{"mode": "apply"})
	assert.ErrorIs(t, err, form.ErrInvalid)
	assert.Equal(t, map[string]string{"confirm": "a value is needed"}, s.Errors)

	s, err = f.Bind(map[string]string{"mode": "apply", "confirm": "malta/apps"})
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"mode": "apply", "confirm": "malta/apps"}, s.Values)
}
