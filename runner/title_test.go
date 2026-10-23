package runner_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/runner"
)

func TestCleanTitle(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "building images", runner.CleanTitle("  building\x07 images\x1b \t"))
	assert.Equal(t, "", runner.CleanTitle(" \x01\x02 "))
	assert.Equal(t, "ok", runner.CleanTitle("o\xffk"))
	long := runner.CleanTitle(strings.Repeat("é", 200))
	assert.Equal(t, runner.TitleMax, len([]rune(long)))
}
