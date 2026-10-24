package view_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

// An excerpt is the words of markdown on one line, without its marks,
// cut at a word.
func TestExcerpt(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Skip watering today. Next is the 25th.", view.Excerpt("## Skip **watering** today.\n\n- Next is the `25th`.\n", 80))
	assert.Equal(t, "before after", view.Excerpt("before\n```sh\n```\nafter", 80))
	assert.Equal(t, "Rain: done.", view.Excerpt("Rain:\n\n| Day | mm |\n|---|---|\n| 21 | 3 |\n\ndone.", 80), "a table stays out")
	long := strings.Repeat("rain ", 40)
	cut := view.Excerpt(long, 20)
	assert.Equal(t, "rain rain rain rain…", cut)
	assert.Equal(t, "", view.Excerpt("  \n", 20))
	assert.Equal(t, "čćž…", view.Excerpt("čćžšđ", 3), "a cut counts runes")
}
