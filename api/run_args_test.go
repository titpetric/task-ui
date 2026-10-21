package api_test

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestRunArgs_UnmarshalJSON(t *testing.T) {
	t.Parallel()
	var req api.RunRequest
	assert.NoError(t, json.Unmarshal([]byte(`{"folder":"malta","dir":"dockerfiles","job":"custom","args":"ps -a '\"; id\"'"}`), &req))
	assert.Equal(t, "dockerfiles", req.Dir)
	assert.Equal(t, api.RunArgs{"ps", "-a", `"; id"`}, req.Args)

	var list api.RunArgs
	assert.NoError(t, json.Unmarshal([]byte(`["logs", "-f", "a b"]`), &list))
	assert.Equal(t, api.RunArgs{"logs", "-f", "a b"}, list)
	assert.NoError(t, json.Unmarshal([]byte(`null`), &list))

	for _, bad := range []string{`"'open"`, `["a\nb"]`, `42`, `[1]`} {
		var args api.RunArgs
		assert.Error(t, json.Unmarshal([]byte(bad), &args), bad)
	}
}

// The words typed for a job that takes args reach the exec: step of its
// atkins job as argv, through atkins: ps -a is two arguments and "; id"
// one, never a command. Needs atkins in PATH.
func TestRun_ArgsThroughAtkins(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("atkins"); err != nil {
		t.Skip("atkins not in PATH")
	}
	f := newFixture(t)
	stack := filepath.Join(f.root, "infra", "stack")
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
