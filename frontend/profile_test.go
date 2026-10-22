package frontend_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/permission"
)

// A profile lists what a user did that the one who looks may see: runs
// and prompts in folders they see, notes with note/view.
func TestProfile(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	ops := f.users[asOps]
	for _, folder := range []string{"infra", "private"} {
		assert.NoError(t, f.opts.Runs.Create(ctx, &model.Run{Folder: folder, Job: "deploy-" + folder, UserID: ops.ID, UserName: ops.Name}))
	}
	assert.NoError(t, f.opts.Runs.Create(ctx, &model.Run{Folder: "infra", Job: "by-admin", UserName: asAdmin}))
	notes := f.opts.Board.Notes
	note := &model.Note{UserName: ops.Name, Body: "rotate the certificates"}
	assert.NoError(t, notes.Create(ctx, note))
	seen := &model.Note{UserName: ops.Name, Body: "what builds here", WorkspaceID: "ws", Dir: "infra/docs", State: model.NoteDone}
	assert.NoError(t, notes.Create(ctx, seen))
	assert.NoError(t, notes.Create(ctx, &model.Note{UserName: ops.Name, Body: "a private question", WorkspaceID: "ws", Dir: "private"}))

	body := f.page(t, asAdmin, "/users/ops")
	assert.Contains(t, body, `<h2>ops <span class="badge">runner</span></h2>`)
	assert.Contains(t, body, "deploy-infra")
	assert.Contains(t, body, "deploy-private")
	assert.NotContains(t, body, "infra / by-admin</a>")
	assert.Contains(t, body, `href="/notes#note-`+note.ID+`">rotate the certificates</a>`)
	assert.Contains(t, body, `href="/f/_/infra/docs#note-`+seen.ID+`">what builds here</a>`)
	assert.Contains(t, body, "a private question")
	assert.Contains(t, body, `href="/runs?user=ops"`)
	assert.Contains(t, body, `data-api="/api/users/`+ops.ID+`" data-method="PUT"`)
	assert.Contains(t, body, "MCP access")

	// limited sees users, infra only, and no notes.
	policies := f.opts.Users.Policies()
	for _, path := range []string{"user/view", "prompt/view"} {
		assert.NoError(t, policies.Put(ctx, permission.KindUser, asLimited, path, permission.Allow))
	}
	assert.NoError(t, policies.Put(ctx, permission.KindUser, asLimited, "note/view", permission.Deny))
	limited := f.page(t, asLimited, "/users/ops")
	assert.Contains(t, limited, "deploy-infra")
	assert.NotContains(t, limited, "deploy-private")
	assert.NotContains(t, limited, "rotate the certificates")
	assert.Contains(t, limited, "what builds here")
	assert.NotContains(t, limited, "a private question")
	assert.NotContains(t, limited, `name="folders"`, "the folders change with user/edit")

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/users/nobody").Status)
	assert.Equal(t, http.StatusForbidden, f.get(t, asViewer, "/users/ops").Status)
	assert.Contains(t, f.page(t, asViewer, "/users/viewer"), "Nothing you can see yet.")
}

// A profile's matrix is the user's own policy with what they may do; it
// changes with user/edit, never an admin's, and not your own without
// user/edit through your role.
func TestProfile_Permissions(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Users.Policies().Put(ctx, permission.KindUser, asOps, "shell", permission.Deny))

	body := f.page(t, asAdmin, "/users/ops")
	assert.Contains(t, body, `id="permissions"`)
	assert.Contains(t, body, `data-api="/api/permissions/users/ops" data-method="PUT" data-next="reload" data-matrix`)
	assert.Contains(t, body, `data-tri="deny" data-title="Enter Shell" title="Enter Shell: deny"`)
	assert.Contains(t, body, `<input type="hidden" name="shell" value="deny" data-was="deny" disabled>`)
	assert.Contains(t, body, `<span class="eff ok" title="Result: allowed">✓</span>`)
	assert.Contains(t, body, `href="/settings/permissions#role-runner"`)

	own := f.page(t, asOps, "/users/ops")
	assert.Contains(t, own, `id="permissions"`)
	assert.NotContains(t, own, "data-matrix")
	assert.Contains(t, own, `title="Enter Shell: deny" aria-label="Enter Shell: deny" disabled>`)

	admin := f.page(t, asAdmin, "/users/admin")
	assert.Contains(t, admin, "Admins can do everything;")
	assert.NotContains(t, admin, "data-matrix")
	assert.NotContains(t, admin, "Delete user")
}
