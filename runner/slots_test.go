package runner_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

func TestSlots(t *testing.T) {
	t.Parallel()
	s := runner.NewSlots(2)
	assert.Equal(t, 2, s.Limit())
	assert.True(t, s.Take("a"))
	assert.True(t, s.Take("a"), "a run holding a slot keeps it")
	assert.False(t, s.Full())
	assert.True(t, s.Take("b"))
	assert.True(t, s.Full())

	assert.False(t, s.Take("c"), "the third run waits")
	assert.False(t, s.Take("d"))
	assert.False(t, s.Take("c"), "a waiting run keeps its place")
	assert.Equal(t, 2, s.Going())
	assert.Equal(t, 2, s.Waiting())
	assert.Equal(t, 1, s.Position("c"))
	assert.Equal(t, 2, s.Position("d"))
	assert.Equal(t, 0, s.Position("a"))

	assert.Equal(t, []string{"c"}, s.Free("a"), "the first in line takes the freed slot")
	assert.Equal(t, 0, s.Position("c"))
	assert.Equal(t, 1, s.Position("d"))
	assert.False(t, s.Take("e"), "a run that comes waits behind the line")
	assert.Equal(t, []string{"d"}, s.Free("b"))
	assert.Len(t, s.Free("missing"), 0, "a full instance frees nothing for a run it never had")
}

func TestSlots_Drop(t *testing.T) {
	t.Parallel()
	s := runner.NewSlots(1)
	assert.True(t, s.Take("a"))
	assert.False(t, s.Take("b"))
	assert.False(t, s.Take("c"))
	assert.True(t, s.Drop("b"), "a stopped run leaves the line")
	assert.False(t, s.Drop("b"))
	assert.False(t, s.Drop("a"), "a run going does not wait")
	assert.Equal(t, 1, s.Position("c"))
	assert.Equal(t, []string{"c"}, s.Free("a"))
	assert.False(t, s.Take("d"))
	assert.Equal(t, []string{"d"}, s.Clear())
	assert.Equal(t, 0, s.Waiting())
}

func TestNewSlots(t *testing.T) {
	t.Parallel()
	for _, limit := range []int{0, -1} {
		s := runner.NewSlots(limit)
		assert.Equal(t, 1, s.Limit(), "one run goes at the least")
		assert.True(t, s.Take("a"))
		assert.False(t, s.Take("b"))
	}
}

func TestTakesSlot(t *testing.T) {
	t.Parallel()
	f := &folder.Folder{ID: "web", Dir: "/srv/web"}
	assert.False(t, runner.TakesSlot(f.ShellJob()), "@shell")
	assert.False(t, runner.TakesSlot(f.ClaudeJob()), "@claude")
	interactive := &folder.Job{ID: "console", Command: []string{"atkins", "console"}}
	interactive.Interactive = true
	assert.False(t, runner.TakesSlot(interactive))
	assert.True(t, runner.TakesSlot(&folder.Job{ID: "lint", Command: []string{"atkins", "lint"}}))
	assert.True(t, runner.TakesSlot(f.PromptJob()), "a prompt does work")
	assert.False(t, runner.TakesSlot(nil))
}
