package view_test

import (
	"bytes"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A home without a repository is empty, a nil home too.
func TestWorkspaceHome_Empty(t *testing.T) {
	t.Parallel()
	var none *view.WorkspaceHome
	assert.True(t, none.Empty())
	assert.True(t, (&view.WorkspaceHome{Folder: "malta/home"}).Empty())
	assert.False(t, (&view.WorkspaceHome{Folder: "malta/home", URL: "git@example.com:org/workspace.git"}).Empty())
}

// A workspace's settings page shows its home: an empty one with Clone,
// a checkout with its repository, Replace and Remove.
func TestWorkspaceEdit_Home(t *testing.T) {
	t.Parallel()
	v := shared(t)
	p := admin(t)
	p.Workspace = &model.Workspace{Folder: "malta", Title: "Malta"}
	p.Workspaces = map[string]int{"malta": 1}
	p.Form = map[string]string{"title": "Malta"}
	p.Home = &view.WorkspaceHome{Folder: "malta/home", Dir: "/data/malta/home"}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "workspace_edit", p))
	body := buf.String()
	assert.Contains(t, body, `<h3>Home repository <small>malta/home/, the HOME of the workspace&#39;s runs</small></h3>`)
	assert.Contains(t, body, `<p class="empty-line">The home is an empty directory.</p>`)
	assert.Contains(t, body, `data-dialog-open="home-replace-dialog">`)
	assert.Contains(t, body, `>Clone</button>`)
	assert.NotContains(t, body, `>Remove</button>`)
	assert.Contains(t, body, `data-api="/api/workspaces/malta/home" data-method="PUT"`)

	p.Home.URL, p.Home.Branch = "git@example.com:malta/workspace.git", "main"
	p.Home.Checkout = &git.Info{Remote: p.Home.URL, Branch: "main", Commit: "abc1234", Subject: "Skills"}
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "workspace_edit", p))
	body = buf.String()
	assert.NotContains(t, body, `The home is an empty directory.`)
	assert.Contains(t, body, `git@example.com:malta/workspace.git`)
	assert.Contains(t, body, `href="/f/malta/home"`)
	assert.Contains(t, body, `abc1234`)
	assert.Contains(t, body, `>Replace</button>`)
	assert.Contains(t, body, `>Remove</button>`)
	assert.Contains(t, body, `data-api="/api/workspaces/malta/home" data-method="DELETE"`)
}
