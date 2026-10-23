package api_test

import (
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestWindow_Apply(t *testing.T) {
	t.Parallel()

	// A window that may run an interactive job types into it, as before.
	shell := &terminal{}
	w := &api.Window{Terminal: shell, Interactive: true}
	assert.Nil(t, w.Apply([]byte(`{"type":"input","message":"ls\r"}`)))
	assert.Equal(t, "ls\r", shell.input.String())

	// A job that takes no input drops the keystrokes until a user with
	// the input permission takes it, and again after release.
	build := &terminal{}
	w = &api.Window{Terminal: build, Taker: "black"}
	w.Apply([]byte(`{"type":"input","message":"no"}`))
	assert.Equal(t, "", build.input.String())
	assert.Equal(t, `{"type":"input","state":"on"}`, string(w.Apply([]byte(`{"type":"take"}`))))
	w.Apply([]byte(`{"type":"input","message":"yes\r"}`))
	assert.Equal(t, "yes\r", build.input.String())
	assert.Equal(t, []string{"black"}, build.taken)
	assert.Equal(t, `{"type":"input","state":"off"}`, string(w.Apply([]byte(`{"type":"release"}`))))
	w.Apply([]byte(`{"type":"input","message":"late"}`))
	assert.Equal(t, "yes\r", build.input.String())

	// It takes input when the window could not otherwise end the run.
	watched := &terminal{}
	w = &api.Window{Terminal: watched, ReadOnly: true, Taker: "root"}
	w.Apply([]byte(`{"type":"take"}`))
	w.Apply([]byte(`{"type":"input","message":"y"}`))
	assert.Equal(t, "y", watched.input.String())
	w.Apply([]byte(`{"type":"stop"}`))
	assert.False(t, watched.stopped, "a read-only window does not end the run")
}

func TestWindow_Apply_refused(t *testing.T) {
	t.Parallel()
	build := &terminal{}
	w := &api.Window{Terminal: build}
	reply := string(w.Apply([]byte(`{"type":"take"}`)))
	assert.Contains(t, reply, `"state":"off"`)
	assert.Contains(t, reply, "input permission")
	w.Apply([]byte(`{"type":"input","message":"yes\r"}`))
	assert.Equal(t, "", build.input.String())
	assert.Len(t, build.taken, 0)

	// An interactive job a window may not run stays read-only.
	shell := &terminal{}
	w = &api.Window{Terminal: shell, Interactive: true, ReadOnly: true}
	w.Apply([]byte(`{"type":"input","message":"ls"}`))
	w.Apply([]byte(`not json`))
	assert.Equal(t, "", shell.input.String())
}
