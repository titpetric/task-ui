//go:build integration

package api_test

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The words typed for a job that takes args reach the exec: step of its
// atkins job as argv, through atkins: ps -a is two arguments and "; id"
// one, never a command. Needs atkins in PATH.
func TestRun_ArgsThroughAtkins(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("atkins"); err != nil {
		t.Skip("atkins not in PATH")
	}
	f := newFixture(t)
	stack := filepath.Join(f.root, "src", "infra", "stack")
	assert.NoError(t, os.MkdirAll(stack, 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(stack, "atkins.yml"), []byte("jobs:\n  custom:\n    desc: Custom\n    args: true\n    steps:\n      - exec: [sh, -c, 'printf \"[%s]\" \"$@\" > out.txt', sh]\n"), 0o644))

	res := f.call(t, asOps, http.MethodPost, "/runs", map[string]any{"folder": "infra", "dir": "stack", "job": "custom", "args": `ps -a "; id"`, "cols": 80, "rows": 24})
	assert.Equal(t, http.StatusCreated, res.Status, string(res.Raw))
	run := f.finished(t, res.str("id"))
	assert.Equal(t, model.StateOK, run.State)
	assert.Equal(t, []string{"ps", "-a", "; id"}, run.ArgList())
	out, err := os.ReadFile(filepath.Join(stack, "out.txt"))
	assert.NoError(t, err)
	assert.Equal(t, "[ps][-a][; id]", string(out))
}
