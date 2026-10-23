package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/sshkey"
)

// A folder inside a checkout pulls the checkout, but the keys of the
// checkout's .ssh/ are not loaded: with no policy the run has no agent,
// and the server's own SSH_AUTH_SOCK is not passed on.
func TestManager_CheckoutKeys(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	dir := checkout(t)
	sub := filepath.Join(dir, "terraform")
	assert.NoError(t, os.MkdirAll(sub, 0o755))
	_, err := (&sshkey.Store{Dir: filepath.Join(dir, ".ssh")}).Generate("rsa", "deploy")
	assert.NoError(t, err)

	fold := &folder.Folder{ID: "repo/terraform", Dir: sub}
	job := &folder.Job{Folder: fold.ID, ID: "plan", Kind: "script", Command: []string{"sh", "-c", `echo "sock=[$SSH_AUTH_SOCK]"; cat ../README`}}
	job.Name = "plan"
	fold.Jobs = []*folder.Job{job}

	session, err := f.manager.Start(context.Background(), runner.Request{Folder: fold, Job: job, User: &model.User{ID: "u1", Name: "ops", Role: model.RoleAdmin}, Cols: 80, Rows: 24})
	assert.NoError(t, err)
	out, status := collect(t, session.Attach(false))
	assert.Equal(t, model.StateOK, status.State, out)
	assert.Contains(t, out, "git pull --ff-only")
	assert.Contains(t, out, "ssh-agent: (none)")
	assert.NotContains(t, out, "id_rsa")
	assert.Contains(t, out, "sock=[]")
	assert.Contains(t, out, "hello")
}

// An instance key reaches the runs of the roles and the users it links,
// and no other run: ssh-add lists it in a linked user's run, and an
// unlinked user's run has no agent.
func TestManager_LinkedKeys(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.keys.Link(ctx, "task_ssh_ed25519", []string{model.RoleRunner}, nil))
	script := `ssh-add -l && echo agent-ok; echo "sock=[$SSH_AUTH_SOCK]"`

	runner := f.shellJob("policy-runner", script, false)
	runner.User = &model.User{ID: "u1", Name: "ops", Role: model.RoleRunner}
	session, err := f.manager.Start(ctx, runner)
	assert.NoError(t, err)
	out, status := collect(t, session.Attach(false))
	assert.Equal(t, model.StateOK, status.State, out)
	assert.Contains(t, out, "ssh-agent: task_ssh_ed25519")
	assert.Contains(t, out, "task_ssh_ed25519 (ED25519)")
	assert.Contains(t, out, "agent-ok")

	admin := f.shellJob("policy-admin", script, false)
	admin.User = &model.User{ID: "u2", Name: "root", Role: model.RoleAdmin}
	session, err = f.manager.Start(ctx, admin)
	assert.NoError(t, err)
	out, status = collect(t, session.Attach(false))
	assert.Equal(t, model.StateOK, status.State, out)
	assert.Contains(t, out, "ssh-agent: (none)")
	assert.NotContains(t, out, "agent-ok")
	assert.Contains(t, out, "sock=[]")

	// Linked to the admin by name, the key reaches their runs too.
	assert.NoError(t, f.keys.Link(ctx, "task_ssh_ed25519", []string{model.RoleRunner}, []string{"root"}))
	admin = f.shellJob("policy-admin-named", script, false)
	admin.User = &model.User{ID: "u2", Name: "root", Role: model.RoleAdmin}
	session, err = f.manager.Start(ctx, admin)
	assert.NoError(t, err)
	out, _ = collect(t, session.Attach(false))
	assert.Contains(t, out, "ssh-agent: task_ssh_ed25519")
	assert.Contains(t, out, "agent-ok")
	entries, _ := os.ReadDir(f.keys.Tmp)
	assert.Len(t, entries, 0, "no key, no socket on disk after the runs")
}

// A run gets its workspace's SSH keys beside the instance ones: a key
// without a link goes to every run in the workspace, a key linked to
// admins skips a runner's run, and the encryption key goes to none. A
// run in _ gets _'s keys, and a run in another workspace none of these.
func TestManager_WorkspaceKeys(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.keys.Link(ctx, "task_ssh_ed25519", []string{model.RoleRunner, model.RoleAdmin}, nil))
	assert.NoError(t, f.workspaces.Create(ctx, &model.Workspace{Folder: "ws"}))
	assert.NoError(t, f.workspaces.EnsureDefault(ctx))
	ws, err := f.workspaces.GetByFolder(ctx, "ws")
	assert.NoError(t, err)
	scope := keystore.ScopeWorkspace(ws.ID)
	encrypt, err := f.keys.Generate(ctx, keystore.Spec{Name: "workspace_ws_encryption", Purpose: keystore.PurposeEncrypt, Scope: scope})
	assert.NoError(t, err)
	deploy, err := f.keys.Generate(ctx, keystore.Spec{Name: "workspace_ws_deploy", Scope: scope})
	assert.NoError(t, err)
	admin, err := f.keys.Generate(ctx, keystore.Spec{Name: "workspace_ws_admin", Scope: scope})
	assert.NoError(t, err)
	assert.NoError(t, f.keys.Link(ctx, admin.Name, []string{model.RoleAdmin}, nil))
	script := `ssh-add -l; ssh-add -L`

	run := func(id, role string) string {
		t.Helper()
		req := f.shellJob(id, script, false)
		req.User = &model.User{ID: "u1", Name: "ops", Role: role}
		session, err := f.manager.Start(ctx, req)
		assert.NoError(t, err)
		out, _ := collect(t, session.Attach(false))
		return out
	}
	pub := func(k *keystore.Key) string { return strings.Fields(k.PublicKey)[1] }

	out := run("ws/runner", model.RoleRunner)
	assert.Contains(t, out, "ssh-agent: task_ssh_ed25519, workspace_ws_deploy_ed25519\x1b")
	assert.Contains(t, out, pub(deploy))
	assert.NotContains(t, out, pub(admin))
	assert.NotContains(t, out, pub(encrypt))

	out = run("ws/admin", model.RoleAdmin)
	assert.Contains(t, out, "ssh-agent: task_ssh_ed25519, workspace_ws_admin_ed25519, workspace_ws_deploy_ed25519\x1b")
	assert.NotContains(t, out, pub(encrypt))

	// The default workspace's own keys, and none of ws's.
	root, err := f.workspaces.GetByFolder(ctx, model.Default)
	assert.NoError(t, err)
	_, err = f.keys.Generate(ctx, keystore.Spec{Name: "workspace_default_root", Scope: keystore.ScopeWorkspace(root.ID)})
	assert.NoError(t, err)
	out = run("plain", model.RoleViewer)
	assert.Contains(t, out, "ssh-agent: workspace_default_root_ed25519\x1b")
	assert.NotContains(t, out, pub(deploy))
}
