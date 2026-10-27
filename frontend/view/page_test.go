package view_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/permission"
)

func TestPage_CanRun(t *testing.T) {
	t.Parallel()
	p := admin(t)
	f, job := p.Folder, p.Job

	assert.True(t, p.CanRun(f, job))
	assert.False(t, p.CanRun(nil, job))
	assert.False(t, p.CanRun(f, nil))

	signIn(p, &model.User{Name: "guest", Role: model.RoleViewer}, nil)
	assert.False(t, p.CanRun(f, job))

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"run": permission.Allow})
	assert.True(t, p.CanRun(f, job))
	job.Allow = []string{"admin"}
	assert.False(t, p.CanRun(f, job))
}

// signIn puts a user on a page, with what the user's role allows.
func signIn(p *view.Page, user *model.User, role permission.Policy) {
	p.User, p.Permissions = user, permission.Set{User: user, Role: role}
}

// |> Summarize takes Summaries, on a job that ends on its own.
func TestPage_Summarizes(t *testing.T) {
	t.Parallel()
	f := &folder.Folder{ID: "infra", Dir: t.TempDir()}
	plan := &folder.Job{ID: "plan"}
	p := &view.Page{}
	assert.False(t, p.Summarizes(plan))
	p.Summaries = true
	assert.True(t, p.Summarizes(plan))
	assert.False(t, p.Summarizes(nil))
	assert.False(t, p.Summarizes(f.ShellJob()))
	assert.False(t, p.Summarizes(f.PromptJob()))
	interactive := &folder.Job{ID: "console"}
	interactive.Interactive = true
	assert.False(t, p.Summarizes(interactive))
}

// A job's dot is how its last run ended, pending for none.
func TestPage_LastState(t *testing.T) {
	t.Parallel()
	job := &folder.Job{Folder: "infra", ID: "plan"}
	p := &view.Page{Last: map[string]view.Run{job.Key(): {Run: model.Run{State: model.StateFailed}}}}
	assert.Equal(t, model.StateFailed, p.LastState(job))
	assert.Equal(t, "pending", p.LastState(&folder.Job{Folder: "infra", ID: "apply"}))
	assert.Equal(t, "pending", p.LastState(nil))
}

func TestPage_Runs(t *testing.T) {
	t.Parallel()
	p := admin(t)
	key := p.Recents[0].Entries[0].Run.Key()

	assert.NotNil(t, p.RunningRun(key))
	assert.Equal(t, "ops", p.RunningRun(key).UserName)
	assert.Nil(t, p.RunningRun("infra/other"))

	assert.Nil(t, p.LastRun(key))
	p.Last = map[string]view.Run{key: *p.Recents[0].Entries[0].Run}
	assert.NotNil(t, p.LastRun(key))
}

func TestPage_Paths(t *testing.T) {
	t.Parallel()
	p := &view.Page{Path: "/settings/keys"}

	assert.True(t, p.IsPath("/settings/keys"))
	assert.False(t, p.IsPath("/settings"))
	assert.True(t, p.Under("/settings"))
	assert.True(t, p.Under("/settings/keys"))
	assert.False(t, p.Under("/set"))
}

func TestPage_InHours(t *testing.T) {
	t.Parallel()

	p := &view.Page{Now: time.Date(2026, time.October, 16, 22, 0, 0, 0, time.UTC)}
	assert.False(t, p.InHours("mon-fri 9-17"))
	assert.True(t, p.InHours(""))
	assert.True(t, (&view.Page{}).InHours("daily 0-24"))
}

func TestPage_WhoCanRun(t *testing.T) {
	t.Parallel()
	p := admin(t)
	f, job := p.Folder, p.Job

	assert.Equal(t, "admin, runner", p.WhoCanRun(f, job))
	f.Allow = []string{"ops"}
	job.Allow = []string{"release"}
	assert.Equal(t, "admin, ops, release", p.WhoCanRun(f, job))
	assert.Equal(t, []string{"ops"}, f.Allow)

	var _ *folder.Folder = f
}

func TestPage_CanRunJob(t *testing.T) {
	t.Parallel()
	p := admin(t)

	assert.True(t, p.CanRunAny(p.Folder))
	assert.True(t, p.CanRunJob("infra", "", "plan"))
	assert.False(t, p.CanRunJob("infra", "", "nothing"))
	assert.False(t, p.CanRunJob("nowhere", "", "plan"))

	// A run of a directory's job names the directory: its job is one of
	// that directory's own targets.
	p.Folder.Dir = t.TempDir()
	images := filepath.Join(p.Folder.Dir, "images")
	assert.NoError(t, os.MkdirAll(images, 0o755))
	// A Taskfile, which task-ui parses itself: the atkins binary lists an
	// atkins pipeline, a process this unit test does not start.
	assert.NoError(t, os.WriteFile(filepath.Join(images, "Taskfile.yml"), []byte("tasks:\n  build:\n    desc: Build the images\n"), 0o644))
	assert.True(t, p.CanRunJob("infra", "images", "build"))
	assert.False(t, p.CanRunJob("infra", "", "build"))
	assert.False(t, p.CanRunJob("infra", "images", "plan"))
	assert.False(t, p.CanRunJob("infra", "..", "plan"))

	p.User = &model.User{Name: "guest", Role: model.RoleViewer}
	assert.False(t, p.CanRunAny(p.Folder))
	assert.False(t, p.CanRunJob("infra", "", "plan"))
	assert.False(t, p.CanRunJob("infra", "images", "build"))
}

// A job's page is under its folder's, after its directory for a job of
// a directory inside the folder.
func TestPage_JobHref(t *testing.T) {
	t.Parallel()
	p := &view.Page{Workspaces: map[string]int{"malta": 0}}
	f := &folder.Folder{ID: "malta"}
	job := &folder.Job{ID: "build"}
	assert.Equal(t, "/f/malta/build", p.JobHref(f, job))
	job.Dir = "dockerfiles/blog"
	assert.Equal(t, "/f/malta/dockerfiles/blog/build", p.JobHref(f, job))
	assert.Equal(t, "/f/internal/acme/@git:status", p.JobHref(&folder.Folder{ID: "acme"}, &folder.Job{ID: "@git:status"}))
	assert.Equal(t, "", p.JobHref(nil, job))
}

// The sidebar's groups come from the folders it lists and the
// workspaces: a workspace holds its projects.
func TestPage_Groups(t *testing.T) {
	t.Parallel()
	p := &view.Page{SideFolders: []*folder.Folder{{ID: "malta"}, {ID: "malta/deploy"}}}
	assert.Equal(t, []string{"malta/deploy"}, idsOf(p.Groups()[0].Nested))

	p.Workspaces = map[string]int{"malta": 0}
	groups := p.Groups()
	assert.Len(t, groups, 1)
	assert.True(t, groups[0].Workspace)
	assert.Equal(t, "malta/deploy", groups[0].Projects[0].Top.ID)
}

// A page links a folder among its workspaces: a named one by its folder,
// a project at the root under /f/internal/.
func TestPage_Href(t *testing.T) {
	t.Parallel()
	p := &view.Page{Workspaces: map[string]int{"acme": 0}}
	assert.Equal(t, "/f/acme/api", p.Href("acme/api"))
	assert.Equal(t, "/f/internal/infra", p.Href("infra"))
	assert.Equal(t, "/f/internal/infra", (&view.Page{}).Href("infra"))
}

// A run links its place and its job's page; a prompt's, the Notes
// agent's and an older run of the internal folder have no job page, and
// internal no folder page.
func TestPage_RunJobHref(t *testing.T) {
	t.Parallel()
	p := &view.Page{Workspaces: map[string]int{"acme": 0}}
	assert.Equal(t, "/f/acme/dockerfiles/build", p.RunJobHref("acme", "dockerfiles", "build"))
	assert.Equal(t, "/f/internal/infra/@claude", p.RunJobHref("infra", "", "@claude"))
	assert.Equal(t, "", p.RunJobHref("acme", "dockerfiles", "@claude-prompt"))
	assert.Equal(t, "", p.RunJobHref("internal", "", "@claude-note"))

	assert.Equal(t, "/f/acme/dockerfiles", p.RunFolderHref("acme", "dockerfiles"))
	assert.Equal(t, "/f/internal", p.RunFolderHref("internal", ""))
}

func TestPage_Links(t *testing.T) {
	t.Parallel()
	p := &view.Page{Workspaces: map[string]int{"acme": 0}}
	links := p.Links([]view.Step{{Folder: "acme/api", Job: "build"}, {Folder: "infra", Job: "plan"}})
	assert.Equal(t, "/f/acme/api/build", links[0].Href)
	assert.Equal(t, "/f/internal/infra/plan", links[1].Href)
}

func TestPage_Can(t *testing.T) {
	t.Parallel()
	p := admin(t)
	assert.True(t, p.Can("secret/delete"))

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"note/all": permission.Allow})
	assert.True(t, p.Can("note/edit"))
	assert.False(t, p.Can("secret/view"))

	// Permissions of someone else are none of the user's.
	p.User = &model.User{Name: "dev", Role: model.RoleRunner}
	assert.False(t, p.Can("note/edit"))
}

func TestPage_CanSettings(t *testing.T) {
	t.Parallel()
	p := admin(t)
	assert.True(t, p.CanSettings())

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"run": permission.Allow})
	assert.False(t, p.CanSettings())
	signIn(p, p.User, permission.Policy{"key/view": permission.Allow})
	assert.True(t, p.CanSettings())
}

func TestPage_Deletes(t *testing.T) {
	t.Parallel()
	p := admin(t)
	mine := model.Note{UserName: "ops"}
	prompt := model.Note{UserName: "ops", WorkspaceID: "w1"}
	assert.True(t, p.Deletes(mine), "an admin deletes anyone's")

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"note/delete": permission.Allow})
	assert.True(t, p.Deletes(mine))
	assert.False(t, p.Deletes(model.Note{UserName: "dev"}))
	assert.False(t, p.Deletes(prompt), "a prompt takes prompt/delete")
	reply := prompt
	reply.ParentID = "n1"
	assert.True(t, p.Deletes(reply), "a reply to a prompt is a note")

	signIn(p, p.User, permission.Policy{"prompt/delete": permission.Allow})
	assert.True(t, p.Deletes(prompt))
	assert.False(t, p.Deletes(mine))
}

func TestPage_Edits(t *testing.T) {
	t.Parallel()
	p := admin(t)
	thread := view.Thread{Note: model.Note{ID: "n1", UserName: "ops"}}
	assert.True(t, p.Edits(thread.Note, thread))

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, nil)
	assert.False(t, p.Edits(thread.Note, thread), "without note/edit")
	signIn(p, p.User, permission.Policy{"note/edit": permission.Allow})
	assert.True(t, p.Edits(thread.Note, thread))
	assert.False(t, p.Edits(model.Note{ID: "n2", UserName: "dev"}, thread))
}

func TestPage_CanProfile(t *testing.T) {
	t.Parallel()
	p := admin(t)
	p.Roles = map[string]string{"ops": model.RoleRunner, "admin": model.RoleAdmin}
	assert.True(t, p.CanProfile("ops"))
	assert.False(t, p.CanProfile("claude"), "no user by that name")

	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"run": permission.Allow})
	assert.True(t, p.CanProfile("ops"), "your own")
	assert.False(t, p.CanProfile("admin"))
	signIn(p, p.User, permission.Policy{"user/view": permission.Allow})
	assert.True(t, p.CanProfile("admin"))
	assert.False(t, (&view.Page{}).CanProfile("ops"))
}

func TestPage_UserNames(t *testing.T) {
	t.Parallel()
	p := &view.Page{Roles: map[string]string{"ops": model.RoleRunner, "admin": model.RoleAdmin}}
	assert.Equal(t, []string{"admin", "ops"}, p.UserNames())
}

// A prompt's place links under the folder that holds it, the deepest
// one the page lists; the root is _.
func TestPage_PlaceHref(t *testing.T) {
	t.Parallel()
	p := &view.Page{Folders: []*folder.Folder{{ID: "malta"}, {ID: "malta/vhosts"}}, Workspaces: map[string]int{"malta": 0}}
	assert.Equal(t, "/f/internal", p.PlaceHref("internal"))
	assert.Equal(t, "/f/malta", p.PlaceHref("malta"))
	assert.Equal(t, "/f/malta/vhosts/www", p.PlaceHref("malta/vhosts/www"))
	assert.Equal(t, "/f/malta/docs", p.PlaceHref("malta/docs"))
	assert.Equal(t, "", p.PlaceHref("private/x"))
}

func TestPage_CanPermissions(t *testing.T) {
	t.Parallel()
	p := admin(t)
	p.Roles = map[string]string{"ops": model.RoleRunner}
	assert.True(t, p.CanPermissions("ops"))
	assert.False(t, p.CanPermissions("claude"))
	signIn(p, &model.User{Name: "ops", Role: model.RoleRunner}, permission.Policy{"user/view": permission.Allow})
	assert.False(t, p.CanPermissions("ops"))
}

// Summarize stands on a run that ended, of a job that Summarizes, with
// no summary yet nor one under way, and on none of Claude's runs.
func TestPage_SummarizesRun(t *testing.T) {
	t.Parallel()
	job := &folder.Job{ID: "plan"}
	run := view.Run{Run: model.Run{ID: "r1", Folder: "infra", Job: "plan", State: model.StateOK}}
	p := &view.Page{}
	assert.False(t, p.SummarizesRun(job, run), "no Summaries")
	p.Summaries = true
	assert.True(t, p.SummarizesRun(job, run))
	running := run
	running.State = model.StateRunning
	assert.False(t, p.SummarizesRun(job, running))
	prompt := run
	prompt.Job = folder.PromptID
	assert.False(t, p.SummarizesRun(job, prompt))
	p.RunReports = map[string]*note.Report{"r1": {RunID: "r1", Pending: true}}
	assert.False(t, p.SummarizesRun(job, run))
	assert.True(t, p.RunReport("r1").Pending)
	assert.Nil(t, p.RunReport("r2"))
}

// The selects of the filters and the page size carry an id their label
// names, which names the dropdown that stands in for each; a
// row of choices under a label is a group the label names.
func TestPage_FieldNames(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Instance = &view.Instance{ParentEnv: "none"}
	for _, c := range []struct{ page, want string }{
		{"runs", `<label for="runs-folder">Folder`},
		{"runs", `<select id="runs-folder" name="folder">`},
		{"runs", `<select id="runs-state" name="state">`},
		{"runs", `<select id="runs-user" name="user">`},
		{"runs", `<label class="page-size" for="page-size">Page size`},
		{"runs", `<select id="page-size" name="n">`},
		{"artifacts", `<label for="art-folder">Folder`},
		{"artifacts", `<select id="art-folder" name="folder">`},
		{"instance", `<div class="field" role="radiogroup" aria-labelledby="parent-env-label">`},
		{"instance", `<span class="label" id="parent-env-label">Forward`},
		{"key_generate", `<div class="field" role="radiogroup" aria-labelledby="key-purpose-label">`},
		{"key_generate", `<span class="label" id="key-purpose-label">Purpose</span>`},
		{"user_form", `<div class="field" data-folder-pick role="group" aria-labelledby="folders-label">`},
		{"user_form", `<span class="label" id="folders-label">Folders</span>`},
	} {
		var buf bytes.Buffer
		assert.NoError(t, v.Render(&buf, c.page, p), c.page)
		assert.Contains(t, buf.String(), c.want, c.page)
	}

	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "project_edit", projectPage(t, &model.User{Name: "admin", Role: model.RoleAdmin}, nil)))
	assert.Contains(t, buf.String(), `<div class="field" role="group" aria-labelledby="pe-allow-label">`)
	assert.Contains(t, buf.String(), `<span class="label" id="pe-allow-label">Who can run</span>`)
}
