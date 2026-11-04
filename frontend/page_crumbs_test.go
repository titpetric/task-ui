package frontend

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The default workspace is spelled internal in an ID and said as Internal in a
// crumb, the way the Keys and Workspaces pages call it. Every other folder
// says what it is.
func TestCrumbName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Internal", crumbName(model.Default))
	assert.Equal(t, "malta", crumbName("malta"))
	assert.Equal(t, "", crumbName(""))
}

// A nested ID becomes one step per segment, and the last step has no link,
// because it is the page the reader is on.
func TestCrumbsOf(t *testing.T) {
	t.Parallel()

	named := map[string]int{"malta": 1}

	out := crumbsOf("malta/src/deploy", "", named)
	assert.Equal(t, 3, len(out))
	assert.Equal(t, "malta", out[0].Text)
	assert.Equal(t, "/f/malta", out[0].Href)
	assert.Equal(t, "src", out[1].Text)
	assert.Equal(t, "/f/malta/src", out[1].Href)
	assert.Equal(t, "deploy", out[2].Text)
	assert.Equal(t, "", out[2].Href, "the page the reader is on is no link")
}

// A last step given is appended with no link, and then the step before it
// keeps its own: a job's page links the folder above it.
func TestCrumbsOf_Last(t *testing.T) {
	t.Parallel()

	out := crumbsOf("malta/src/deploy", "build", map[string]int{"malta": 1})
	assert.Equal(t, 4, len(out))
	assert.Equal(t, "deploy", out[2].Text)
	assert.Equal(t, "/f/malta/src/deploy", out[2].Href, "the folder is a link again")
	assert.Equal(t, "build", out[3].Text)
	assert.Equal(t, "", out[3].Href)
}

// A folder in no named workspace is a project at the root, which lives under
// the default workspace: the walk starts at Internal, so the crumb above a
// project is the page that lists them.
func TestCrumbsOf_Default(t *testing.T) {
	t.Parallel()

	out := crumbsOf("infra", "", nil)
	assert.Equal(t, 2, len(out))
	assert.Equal(t, "Internal", out[0].Text)
	assert.Equal(t, view.FolderHref(model.Default, nil), out[0].Href)
	assert.Equal(t, "infra", out[1].Text)
	assert.Equal(t, "", out[1].Href)

	// The default workspace's own page is one step, not two: it is not
	// nested under itself.
	own := crumbsOf(model.Default, "", nil)
	assert.Equal(t, 1, len(own))
	assert.Equal(t, "Internal", own[0].Text)
	assert.Equal(t, "", own[0].Href)
}

// A named workspace does not pass through Internal, because it is not under
// it: the two trees are siblings and the crumbs say so.
func TestCrumbsOf_Named(t *testing.T) {
	t.Parallel()

	out := crumbsOf("malta", "", map[string]int{"malta": 1})
	assert.Equal(t, 1, len(out))
	assert.Equal(t, "malta", out[0].Text)

	// A workspace the layout no longer names reads as a project under
	// Internal, which is where its folders are served from once it is gone:
	// the trail follows the links and not the layout it was written against.
	moved := crumbsOf("malta", "", nil)
	assert.Equal(t, 2, len(moved))
	assert.Equal(t, "Internal", moved[0].Text)
}
