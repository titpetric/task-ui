package view_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestProfile_Activity(t *testing.T) {
	t.Parallel()
	p := &view.Profile{}
	assert.False(t, p.Activity())
	p.Prompts = []model.Note{{Body: "why"}}
	assert.True(t, p.Activity())
}

// A profile is a page to read: the name and when they joined, Edit top
// right for your own or with user/edit, then Runs, Workflow runs, Notes
// and Prompts in that order, the notes as plain text; no form, no role,
// no matrix. Each empty section says so in one line.
func TestView_RenderProfile(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	render := func(p *view.Page) string {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, "profile", p))
		return buf.String()
	}
	p := admin(t)
	p.Profile.CanEdit = true
	p.Profile.FlowRuns = []model.WorkflowRun{{ID: "w1", WorkflowID: "wf", Name: "release", State: model.StateOK}}
	p.Profile.Notes = []model.Note{{ID: "n2", Body: "**bold** # head [l](u)"}}
	out := render(p)
	assert.Contains(t, out, `<h2>ops</h2>`)
	assert.Contains(t, out, `<a class="btn btn-primary" href="/users/ops/edit"><svg class="ic"><use href="#i-edit"/></svg>Edit</a>`)
	assert.NotContains(t, out, `<form class="form"`)
	assert.NotContains(t, out, `data-api="/api/users/`)
	assert.NotContains(t, out, "data-matrix")
	assert.NotContains(t, out, "<h3>Roles")
	assert.NotContains(t, out, "<h3>Permissions")
	assert.NotContains(t, out, "MCP access")
	assert.NotContains(t, out, `class="badge">runner`)
	order := []string{`<h3>Runs `, `<h3>Workflow runs `, `<h3>Notes `, `<h3>Prompts `}
	last := -1
	for _, head := range order {
		at := strings.Index(out, head)
		assert.True(t, at > last, head)
		last = at
	}
	assert.Contains(t, out, `<a class="excerpt" href="/notes#note-n2">bold # head l</a>`)
	assert.Contains(t, out, `<a href="/workflows/wf"><svg class="ic lead"><use href="#i-flow"/></svg>release</a>`)
	assert.Contains(t, out, `href="/workflows/wf/runs/w1">View</a>`)
	assert.Contains(t, out, `<table class="table cards">`)

	// A viewer with user/view alone reads it without Edit; nothing done
	// yet reads as one line a section.
	p = admin(t)
	p.Profile = &view.Profile{}
	out = render(p)
	assert.NotContains(t, out, `href="/users/ops/edit"`)
	for _, line := range []string{"No runs yet.", "No workflow runs yet.", "No notes yet.", "No prompts yet."} {
		assert.Contains(t, out, `<p class="empty-line">`+line+`</p>`)
	}
}

// The edit page holds the forms of a user, Roles first, then Folders,
// Permissions, Keys and MCP access, each returning to the edit page,
// with View profile top right and Cancel back to the profile.
func TestView_RenderUserEdit(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "user_edit", admin(t)))
	out := buf.String()
	assert.Contains(t, out, `<h2>Edit ops</h2>`)
	assert.Contains(t, out, `<a class="btn" href="/users/ops"><svg class="ic"><use href="#i-users"/></svg>View profile</a>`)
	order := []string{`id="roles"`, `id="folders"`, `id="permissions"`, `id="keys"`, `id="mcp"`}
	last := -1
	for _, id := range order {
		at := strings.Index(out, id)
		assert.True(t, at > last, id)
		last = at
	}
	assert.Contains(t, out, `data-api="/api/users/u1" data-method="PUT" data-next="/users/ops/edit?notice=Roles+saved#roles"`)
	assert.Contains(t, out, `data-next="/users/ops/edit?notice=Folders+saved#folders"`)
	assert.Contains(t, out, `<a class="btn" href="/users/ops">Cancel</a>`)
	assert.Contains(t, out, "data-matrix")
	assert.Contains(t, out, `<a class="btn" href="/users/ops/edit#mcp">Done</a>`)
}

// The users page links a name to the profile, and the row's menu opens
// with Edit, the edit page.
func TestView_RenderUsers(t *testing.T) {
	t.Parallel()

	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Users = []model.User{{ID: "u1", Name: "ops", Role: model.RoleRunner}}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "users", p))
	out := buf.String()
	assert.Contains(t, out, `<a class="user-name" href="/users/ops">ops</a>`)
	edit := strings.Index(out, `href="/users/ops/edit"`)
	assert.True(t, edit > 0)
	assert.True(t, edit < strings.Index(out, `>Profile</span>`))
	assert.NotContains(t, out, "#permissions")
}
