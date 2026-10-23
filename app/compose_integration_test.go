//go:build integration

package app_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
)

// ComposeProject reads the compose label of the container named by the
// host name; a failing docker, or no label, is none.
func TestComposeProject(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	host, err := os.Hostname()
	assert.NoError(t, err)
	write := func(name, body string) string {
		path := filepath.Join(dir, name)
		assert.NoError(t, os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755))
		return path
	}
	labelled := write("labelled", `[ "$1" = inspect ] && [ "$4" = "`+host+`" ] && echo task`)
	bare := write("bare", `echo '<no value>'`)
	failing := write("failing", `exit 1`)
	assert.Equal(t, "task", app.ComposeProject(labelled))
	assert.Equal(t, "", app.ComposeProject(bare))
	assert.Equal(t, "", app.ComposeProject(failing))
	assert.Equal(t, "", app.ComposeProject(filepath.Join(dir, "missing")))
}
