package sshprobe_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/sshprobe"
)

// A result carries the host, the login, the exit code and the output; an
// error only when there is one.
func TestResult(t *testing.T) {
	t.Parallel()
	data, err := json.Marshal(sshprobe.Result{Host: "a", Login: "debian@a", Output: "ok", Duration: time.Second})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"host":"a","login":"debian@a","exit_code":0,"output":"ok","duration":1000000000}`, string(data))
}
