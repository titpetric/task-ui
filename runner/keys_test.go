package runner_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
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

// An instance key reaches the runs of the roles and the users its policy
// names, and no other run: ssh-add lists it in a mapped user's run, and
// an unmapped user's run has no agent.
func TestManager_PolicyKeys(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	assert.NoError(t, f.policies.Set(context.Background(), "id_ed25519", []string{model.RoleRunner}))
	script := `ssh-add -l && echo agent-ok; echo "sock=[$SSH_AUTH_SOCK]"`

	runner := f.shellJob("policy-runner", script, false)
	runner.User = &model.User{ID: "u1", Name: "ops", Role: model.RoleRunner}
	session, err := f.manager.Start(context.Background(), runner)
	assert.NoError(t, err)
	out, status := collect(t, session.Attach(false))
	assert.Equal(t, model.StateOK, status.State, out)
	assert.Contains(t, out, "ssh-agent: id_ed25519")
	assert.Contains(t, out, "id_ed25519 (ED25519)")
	assert.Contains(t, out, "agent-ok")

	admin := f.shellJob("policy-admin", script, false)
	admin.User = &model.User{ID: "u2", Name: "root", Role: model.RoleAdmin}
	session, err = f.manager.Start(context.Background(), admin)
	assert.NoError(t, err)
	out, status = collect(t, session.Attach(false))
	assert.Equal(t, model.StateOK, status.State, out)
	assert.Contains(t, out, "ssh-agent: (none)")
	assert.NotContains(t, out, "agent-ok")
	assert.Contains(t, out, "sock=[]")

	// Given to the admin by name, the key reaches their runs too.
	assert.NoError(t, f.policies.SetUsers(context.Background(), "id_ed25519", []string{"root"}))
	admin = f.shellJob("policy-admin-named", script, false)
	admin.User = &model.User{ID: "u2", Name: "root", Role: model.RoleAdmin}
	session, err = f.manager.Start(context.Background(), admin)
	assert.NoError(t, err)
	out, _ = collect(t, session.Attach(false))
	assert.Contains(t, out, "ssh-agent: id_ed25519")
	assert.Contains(t, out, "agent-ok")
}

// A run gets its workspace's SSH keys beside the instance ones, each as
// ws:<name>: a key without a grant goes to every run in the workspace, a
// key granted to admins skips a runner's run, and the encryption key
// goes to none. A run in _ gets _'s keys, and a run in another
// workspace none of these.
func TestManager_WorkspaceKeys(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.policies.Set(ctx, "id_ed25519", []string{model.RoleRunner, model.RoleAdmin}))
	assert.NoError(t, f.workspaces.Create(ctx, &model.Workspace{Folder: "ws"}))
	assert.NoError(t, f.workspaces.EnsureDefault(ctx))
	ws, err := f.workspaces.GetByFolder(ctx, "ws")
	assert.NoError(t, err)
	_, err = f.keys.Ensure(ws.ID, "encrypt")
	assert.NoError(t, err)
	encrypt := f.keys.EncryptKey(ws.ID)
	deploy, err := f.keys.SSH(ws.ID).GenerateAs("ed25519", "deploy", "deploy")
	assert.NoError(t, err)
	admin, err := f.keys.SSH(ws.ID).GenerateAs("ed25519", "admin", "admin")
	assert.NoError(t, err)
	assert.NoError(t, f.grants.Set(ctx, ws.ID, admin.Name, []string{model.RoleAdmin}, nil))
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

	out := run("ws/runner", model.RoleRunner)
	assert.Contains(t, out, "ssh-agent: id_ed25519, ws:id_ed25519_deploy\x1b")
	assert.Contains(t, out, "id_ed25519 (ED25519)")
	assert.Contains(t, out, deploy.Fingerprint)
	assert.NotContains(t, out, admin.Fingerprint)
	assert.NotContains(t, out, encrypt.Fingerprint)

	out = run("ws/admin", model.RoleAdmin)
	assert.Contains(t, out, "ssh-agent: id_ed25519, ws:id_ed25519_deploy, ws:id_ed25519_admin\x1b")
	assert.Contains(t, out, admin.Fingerprint)
	assert.NotContains(t, out, encrypt.Fingerprint)

	// The default workspace's own keys, and none of ws's.
	root, err := f.workspaces.GetByFolder(ctx, model.Default)
	assert.NoError(t, err)
	_, err = f.keys.SSH(root.ID).GenerateAs("ed25519", "root", "root")
	assert.NoError(t, err)
	out = run("plain", model.RoleViewer)
	assert.Contains(t, out, "ssh-agent: ws:id_ed25519_root\x1b")
	assert.NotContains(t, out, deploy.Fingerprint)
}
