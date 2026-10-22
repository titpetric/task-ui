package config_test

import (
	"testing"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestDefault(t *testing.T) {
	t.Parallel()
	p := config.Default()
	assert.Equal(t, "all", p.Rules[0].Key)
	assert.Equal(t, "workspace", p.Sections[0].Key)
	assert.Equal(t, "Workspaces", p.Sections[0].Title)
	assert.Equal(t, "run", p.Permissions[0].Key)
	assert.Contains(t, p.Paths(), "note/all")
	assert.Contains(t, p.Paths(), "secret/view")
	assert.Contains(t, p.Paths(), "hosts")
	assert.Equal(t, len(p.Sections)*len(p.Rules)+len(p.Permissions), len(p.Paths()))
	assert.Equal(t, len(p.Paths())-len(p.Sections), len(p.Checks()))
	for _, path := range p.Checks() {
		assert.True(t, p.Checkable(path), path)
	}
}

func TestPermissions_Valid(t *testing.T) {
	t.Parallel()
	p := config.Default()
	assert.True(t, p.Valid("note/edit"))
	assert.True(t, p.Valid("note/all"))
	assert.True(t, p.Valid("run"))
	assert.False(t, p.Valid("note"))
	assert.False(t, p.Valid("note/write"))
	assert.False(t, p.Valid("run/view"))
	assert.False(t, p.Valid("nothing/view"))
	assert.False(t, p.Valid(""))
	assert.True(t, p.Checkable("note/edit"))
	assert.False(t, p.Checkable("note/all"))
	assert.False(t, p.Checkable("note/write"))
}

func TestParse(t *testing.T) {
	t.Parallel()
	good := "rules: {all: All, view: View, edit: Edit, delete: Delete}\nsections: {note: Notes}\npermissions: {run: Run}\n"
	p, err := config.Parse([]byte(good))
	assert.NoError(t, err)
	assert.Equal(t, []string{"note/all", "note/view", "note/edit", "note/delete", "run"}, p.Paths())

	for name, bad := range map[string]string{
		"not yaml":        "rules: [",
		"no sections":     "rules: {all: All, view: View, edit: Edit, delete: Delete}\npermissions: {run: Run}\n",
		"a rule missing":  "rules: {all: All, view: View, edit: Edit}\nsections: {note: Notes}\npermissions: {run: Run}\n",
		"a list":          "rules: {all: All, view: View, edit: Edit, delete: Delete}\nsections: [note]\npermissions: {run: Run}\n",
		"an empty title":  "rules: {all: All, view: View, edit: Edit, delete: Delete}\nsections: {note: \"\"}\npermissions: {run: Run}\n",
		"a key in caps":   "rules: {all: All, view: View, edit: Edit, delete: Delete}\nsections: {Note: Notes}\npermissions: {run: Run}\n",
		"a key with a /":  "rules: {all: All, view: View, edit: Edit, delete: Delete}\nsections: {a/b: Notes}\npermissions: {run: Run}\n",
		"a section twice": "rules: {all: All, view: View, edit: Edit, delete: Delete}\nsections: {note: Notes}\npermissions: {note: Notes}\n",
	} {
		_, err := config.Parse([]byte(bad))
		assert.Error(t, err, name)
	}
}
