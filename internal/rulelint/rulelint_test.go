package rulelint_test

import (
	"testing"

	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/rulelint"
)

func TestLint(t *testing.T) {
	t.Parallel()
	perms, err := config.Parse([]byte("rules: {all: All, view: View, edit: Edit, delete: Delete}\nsections: {note: Notes}\npermissions: {run: Run, shell: Shell, hosts: Hosts}\n"))
	assert.NoError(t, err)

	findings, err := rulelint.Lint("testdata/app", perms)
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"app.go:15: unknown rule \"note/write\"",
		"app.go:16: rule \"note/all\" is a policy's shorthand, no check names it",
		"app.go:17: the rule of can is no string literal",
		"config: rule \"hosts\" never checked",
		"page.tpl:2: unknown rule \"secret/view\"",
	}, findings)

	_, err = rulelint.Lint("testdata/none", perms)
	assert.Error(t, err)
}

// The code of task-ui checks every rule of its configuration, and only those.
func TestLint_taskUI(t *testing.T) {
	t.Parallel()
	findings, err := rulelint.Lint("../..", config.Default())
	assert.NoError(t, err)
	assert.Len(t, findings, 0, findings)
	assert.True(t, rulelint.Checks["Allowed"])
}
