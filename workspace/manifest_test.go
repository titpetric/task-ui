package workspace_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/workspace"
)

// A named workspace holds its folder and those inside it; the default
// one every folder outside the named workspaces and task-ui's own.
func TestContains(t *testing.T) {
	t.Parallel()
	named := []string{"acme"}
	for _, c := range []struct {
		workspace, folder string
		want              bool
	}{
		{"acme", "acme", true},
		{"acme", "acme/api", true},
		{"acme", "acmes/api", false},
		{"acme", "web", false},
		{"_", "web", true},
		{"_", "web/docs", true},
		{"_", "acme/api", false},
		{"_", "internal/claude-prompt", false},
	} {
		assert.Equal(t, c.want, workspace.Contains(c.workspace, named, c.folder), c.workspace+" "+c.folder)
	}
}

// A folder goes relative on export and comes back under the target on
// import; the default workspace's folders move under a named target.
func TestRelative(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", workspace.Relative("acme", "acme"))
	assert.Equal(t, "api", workspace.Relative("acme", "acme/api"))
	assert.Equal(t, "web", workspace.Relative("_", "web"))
	assert.Equal(t, "acme2/api", workspace.Absolute("acme2", "api"))
	assert.Equal(t, "acme2", workspace.Absolute("acme2", ""))
	assert.Equal(t, "legacy/web", workspace.Absolute("legacy", workspace.Relative("_", "web")))
	assert.Equal(t, "web", workspace.Absolute("_", "web"))
}

// A shared scope stays shared, under the target's name; a folder scope
// moves like its folder.
func TestRelativeScope(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "workspace:", workspace.RelativeScope("acme", "workspace:acme"))
	assert.Equal(t, "api", workspace.RelativeScope("acme", "acme/api"))
	assert.Equal(t, "workspace:legacy", workspace.AbsoluteScope("legacy", workspace.RelativeScope("_", "workspace:_")))
	assert.Equal(t, "legacy/web", workspace.AbsoluteScope("legacy", workspace.RelativeScope("_", "web")))
}
