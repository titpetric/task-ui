package forgejo_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// hookURL is where the fake instance posts pushes.
const hookURL = "https://task.example/api/webhooks/forgejo"

// instance returns a fake Forgejo with four organisations: Malta, new,
// with its home repository, two projects, an archived repository and
// one named src; Acme, a workspace already, without a home repository
// and with a project row of a repository Forgejo lacks; _, no folder name;
// and Taken, whose name a project at the root holds. Its users: black,
// with a home repository, a project and an archived repository; nobody,
// without a repository; Bad.Name!, no folder name; and Jane, with one
// repository the token does not administer.
func instance(t *testing.T) (*fakeForgejo, *fakeWorkspaces, *fakeProjects) {
	t.Helper()
	fake := newFakeForgejo("t0k3n")
	fake.orgs = []forgejo.Org{{Name: "Malta", FullName: "Malta d.o.o."}, {Name: "Acme"}, {Name: "_"}, {Name: "Taken"}}
	fake.users = []forgejo.User{{Login: "black", FullName: "Tit Petric"}, {Login: "nobody"}, {Login: "Bad.Name!"}, {Login: "Jane", FullName: "Jane Doe"}}
	fake.userRepos["black"] = []forgejo.Repo{
		repo("black", "black", "main", false),
		repo("black", "dotfiles", "main", false),
		repo("black", "old", "main", true),
	}
	fake.userRepos["nobody"] = nil
	notes := repo("Jane", "notes", "main", false)
	notes.Permissions.Admin = false
	fake.userRepos["Jane"] = []forgejo.Repo{notes}
	fake.repos["Malta"] = []forgejo.Repo{
		repo("Malta", "malta", "main", false),
		repo("Malta", "apps", "main", false),
		repo("Malta", "deploy", "dev", false),
		repo("Malta", "old-thing", "main", true),
		repo("Malta", "src", "main", false),
	}
	fake.repos["Acme"] = []forgejo.Repo{repo("Acme", "web", "main", false)}
	fake.repos["_"] = nil
	fake.repos["Taken"] = nil
	workspaces := &fakeWorkspaces{rows: map[string]*model.Workspace{"acme": {ID: "w-acme", Folder: "acme", Title: "Acme"}}}
	projects := &fakeProjects{rows: map[string]*model.Project{
		"acme/vhosts": {Folder: "acme/vhosts", SourceURL: "ssh://git@git.example:2222/Acme/vhosts.git"},
		"taken":       {Folder: "taken", SourceURL: "https://elsewhere/taken.git"},
	}}
	return fake, workspaces, projects
}

// sync returns a Sync of the fake over a fresh root.
func newSync(t *testing.T, fake *fakeForgejo, workspaces *fakeWorkspaces, projects *fakeProjects, root string) (*forgejo.Sync, *fakeGit) {
	t.Helper()
	g := &fakeGit{}
	return &forgejo.Sync{
		Client: fake.client("t0k3n"), Git: g, Workspaces: workspaces, Projects: projects,
		Root: root, HookURL: hookURL, Secret: "s3",
	}, g
}

// A first sync makes the workspaces of the organisations and the users,
// clones the homes and the projects and installs the hooks where the
// token is admin; it skips what it must, a user without a repository
// among it, and reports the project whose repository Forgejo lacks. A
// second sync changes nothing and says so.
func TestSync(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake, workspaces, projects := instance(t)
	root := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Join(root, "acme", "home"), 0o755))
	assert.NoError(t, os.MkdirAll(filepath.Join(root, "acme", "src"), 0o755))
	s, g := newSync(t, fake, workspaces, projects, root)

	var out bytes.Buffer
	assert.NoError(t, s.Run(ctx, &out))
	report := out.String()
	for _, line := range []string{
		"# org Malta",
		"  + workspace malta: created",
		"  + home Malta/malta: cloned main",
		"  + project malta/apps: cloned main",
		"  + project malta/deploy: cloned dev",
		"    skipped Malta/old-thing: archived",
		`    skipped Malta/src: "src" is no folder name`,
		"  + hook Malta/malta: installed",
		"  + hook Malta/apps: installed",
		"  + hook Malta/deploy: installed",
		"# org Acme",
		"    workspace acme: ok",
		"    home acme: no Acme/Acme repository: home is empty",
		"  + project acme/web: cloned main",
		"  + hook Acme/web: installed",
		"# org _",
		`    skipped _: "_" is no folder name`,
		"# org Taken",
		"    skipped Taken: a project named taken is at the root",
		"# user black",
		"  + workspace black: created",
		"  + home black/black: cloned main",
		"  + project black/dotfiles: cloned main",
		"    skipped black/old: archived",
		"  + hook black/black: installed",
		"  + hook black/dotfiles: installed",
		"# user nobody",
		"    skipped nobody: no repositories",
		`    skipped Bad.Name!: "bad.name!" is no folder name`,
		"# user Jane",
		"  + workspace jane: created",
		"    home jane: no Jane/Jane repository: home is empty",
		"  + project jane/notes: cloned main",
		"    skipped hook Jane/notes: the token has no admin access",
		"! orphan acme/vhosts: not on Forgejo, left alone",
		"",
		"Applied: 3 workspaces created, 2 homes cloned, 5 projects cloned, 6 hooks installed, 3 unchanged, 8 skipped, 1 left alone, over 4 organisations and 4 users.",
	} {
		assert.Contains(t, report, line+"\n")
	}
	assert.NotContains(t, report, "hook Malta/src")
	assert.NotContains(t, report, "users skipped")
	assert.NotContains(t, report, "Plan:")
	assert.NotContains(t, report, "\x1b[")
	assert.True(t, strings.Index(report, "# user black") > strings.Index(report, "skipped Taken"), "organisations first, then users")
	assert.True(t, strings.HasSuffix(report, "4 users.\n"), "the footer ends the report")

	malta := workspaces.rows["malta"]
	assert.NotNil(t, malta)
	assert.Equal(t, "Malta d.o.o.", malta.Title)
	assert.Equal(t, "ssh://git@git.example:2222/Malta/malta.git", malta.HomeURL)
	assert.Equal(t, "main", malta.HomeBranch)
	assert.True(t, g.IsRepo(filepath.Join(root, "malta", "home")))
	assert.True(t, g.IsRepo(filepath.Join(root, "malta", "src", "apps")))
	assert.True(t, g.IsRepo(filepath.Join(root, "acme", "src", "web")))
	assert.Equal(t, "ssh://git@git.example:2222/Malta/deploy.git", projects.rows["malta/deploy"].SourceURL)
	assert.Equal(t, "dev", projects.rows["malta/deploy"].Branch)
	assert.Nil(t, projects.rows["malta/old-thing"])
	assert.Nil(t, projects.rows["malta/src"])
	assert.Equal(t, "", workspaces.rows["acme"].HomeURL)
	black := workspaces.rows["black"]
	assert.NotNil(t, black)
	assert.Equal(t, "Tit Petric", black.Title)
	assert.Equal(t, "ssh://git@git.example:2222/black/black.git", black.HomeURL)
	assert.True(t, g.IsRepo(filepath.Join(root, "black", "home")))
	assert.True(t, g.IsRepo(filepath.Join(root, "black", "src", "dotfiles")))
	assert.Nil(t, projects.rows["black/old"])
	assert.Nil(t, workspaces.rows["nobody"])
	assert.Nil(t, workspaces.rows["bad.name!"])
	assert.Equal(t, "Jane Doe", workspaces.rows["jane"].Title)
	assert.Len(t, fake.hooksOf("Jane/notes"), 0)
	assert.Len(t, g.clones, 7)
	assert.Equal(t, 6, fake.created)
	assert.Equal(t, "s3", fake.hooksOf("Malta/apps")[0].Config["secret"])
	assert.Equal(t, "s3", fake.hooksOf("black/dotfiles")[0].Config["secret"])

	// Again: everything is there, the home pulls, no hook goes in twice.
	out.Reset()
	assert.NoError(t, s.Run(ctx, &out))
	report = out.String()
	for _, line := range []string{
		"    workspace malta: ok",
		"  ~ home Malta/malta: pulled",
		"    project malta/apps: exists",
		"    hook Malta/malta: ok",
		"    hook Acme/web: ok",
		"    workspace black: ok",
		"  ~ home black/black: pulled",
		"    project black/dotfiles: exists",
		"    hook black/dotfiles: ok",
		"    skipped hook Jane/notes: the token has no admin access",
		"Applied: 0 workspaces created, 0 homes cloned, 0 projects cloned, 0 hooks installed, 2 updated, 17 unchanged, 8 skipped, 1 left alone, over 4 organisations and 4 users.",
	} {
		assert.Contains(t, report, line+"\n")
	}
	assert.Equal(t, 6, fake.created)
	assert.Equal(t, 0, fake.updated)
	assert.Len(t, g.clones, 7)
	assert.Equal(t, []string{filepath.Join(root, "malta", "home"), filepath.Join(root, "black", "home")}, g.pulls)
}

// A dry run reports what it would do and touches nothing.
func TestSync_DryRun(t *testing.T) {
	t.Parallel()
	fake, workspaces, projects := instance(t)
	root := t.TempDir()
	s, g := newSync(t, fake, workspaces, projects, root)
	s.DryRun = true

	var out bytes.Buffer
	assert.NoError(t, s.Run(context.Background(), &out))
	report := out.String()
	for _, line := range []string{
		"# org Malta",
		"  + workspace malta: would create",
		"  + home Malta/malta: would clone main",
		"  + project malta/apps: would clone main",
		"  + hook Malta/apps: would install",
		"  + project acme/web: would clone main",
		"# user black",
		"  + workspace black: would create",
		"  + home black/black: would clone main",
		"  + project black/dotfiles: would clone main",
		"  + hook black/dotfiles: would install",
		"    skipped nobody: no repositories",
		"    skipped hook Jane/notes: the token has no admin access",
		"! orphan acme/vhosts: not on Forgejo, left alone",
		"Plan: 3 workspaces to create, 2 homes to clone, 5 projects to clone, 6 hooks to install, 3 unchanged, 8 skipped, 1 left alone, over 4 organisations and 4 users.",
	} {
		assert.Contains(t, report, line+"\n")
	}
	assert.NotContains(t, report, "Applied:")
	_, err := os.Stat(filepath.Join(root, "malta"))
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(filepath.Join(root, "black"))
	assert.True(t, os.IsNotExist(err))
	assert.Nil(t, workspaces.rows["malta"])
	assert.Nil(t, workspaces.rows["black"])
	assert.Nil(t, projects.rows["malta/apps"])
	assert.Nil(t, projects.rows["black/dotfiles"])
	assert.Len(t, g.clones, 0)
	assert.Equal(t, 0, fake.created)
}

// Without a public URL no hook goes in, and the report says why; a hook
// that posts elsewhere stays, one of ours that is off comes back on; a
// home that is another repository's checkout stays, as does a project
// directory that is no checkout; a project row's URL follows Forgejo.
func TestSync_Edges(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake, workspaces, projects := instance(t)
	root := t.TempDir()
	s, g := newSync(t, fake, workspaces, projects, root)
	s.HookURL = ""
	var out bytes.Buffer
	assert.NoError(t, s.Run(ctx, &out))
	assert.Contains(t, out.String(), "hooks skipped: the instance has no public URL (TASK_UI_URL)\n")
	assert.NotContains(t, out.String(), "hook ")
	assert.Equal(t, 0, fake.created)

	// The second instance: the acme home is a checkout of something else,
	// malta/apps is a plain directory, malta/deploy a row with an old URL,
	// and the hooks are half there.
	fake, workspaces, projects = instance(t)
	root = t.TempDir()
	s, g = newSync(t, fake, workspaces, projects, root)
	workspaces.rows["acme"].HomeURL = "ssh://git@git.example:2222/Acme/workspace.git"
	fake.repos["Acme"] = append(fake.repos["Acme"], repo("Acme", "acme", "main", false))
	assert.NoError(t, os.MkdirAll(filepath.Join(root, "acme", "home", ".git"), 0o755))
	assert.NoError(t, os.MkdirAll(filepath.Join(root, "malta", "src", "apps"), 0o755))
	assert.NoError(t, os.WriteFile(filepath.Join(root, "malta", "src", "apps", "README"), []byte("x"), 0o644))
	projects.rows["malta/deploy"] = &model.Project{Folder: "malta/deploy", Title: "Deploy", SourceURL: "https://old.example/deploy.git"}
	fake.hooks["Malta/deploy"] = []forgejo.Hook{{ID: 7, Config: map[string]string{"url": hookURL, "content_type": "json"}, Events: []string{"push"}, Active: false}}
	fake.hooks["Acme/web"] = []forgejo.Hook{{ID: 1, Config: map[string]string{"url": "https://ci.example/hook"}, Events: []string{"push"}, Active: true}}

	out.Reset()
	assert.NoError(t, s.Run(ctx, &out))
	report := out.String()
	for _, line := range []string{
		"  ! home Acme/acme: the home is a checkout of ssh://git@git.example:2222/Acme/workspace.git, left alone: replace it on the workspace's settings page",
		"  ! project malta/apps: src/apps exists and is no checkout: left alone",
		"  ~ project malta/deploy: updated url",
		"  ~ hook Malta/deploy: updated",
		"  + hook Acme/web: installed",
		"Applied: 3 workspaces created, 2 homes cloned, 3 projects cloned, 5 hooks installed, 2 updated, 2 unchanged, 8 skipped, 3 left alone, over 4 organisations and 4 users.",
	} {
		assert.Contains(t, report, line+"\n")
	}
	assert.Equal(t, "ssh://git@git.example:2222/Acme/workspace.git", workspaces.rows["acme"].HomeURL)
	assert.Equal(t, "Deploy", projects.rows["malta/deploy"].Title, "the row keeps what the admin set")
	assert.Equal(t, "ssh://git@git.example:2222/Malta/deploy.git", projects.rows["malta/deploy"].SourceURL)
	assert.Nil(t, projects.rows["malta/apps"])
	assert.Equal(t, 1, fake.updated)
	assert.True(t, fake.hooksOf("Malta/deploy")[0].Active)
	assert.Len(t, fake.hooksOf("Acme/web"), 2)
	assert.False(t, strings.Contains(strings.Join(g.clones, "\n"), "Acme/acme"))
	assert.False(t, strings.Contains(strings.Join(g.pulls, "\n"), "acme"))
}

// A sync that cannot list the organisations fails outright; https
// clones take the https URL; a token's user in no organisation syncs
// nothing.
func TestSync_Refused(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake, workspaces, projects := instance(t)
	s, _ := newSync(t, fake, workspaces, projects, t.TempDir())
	s.Client = fake.client("wrong")
	var out bytes.Buffer
	err := s.Run(ctx, &out)
	assert.ErrorIs(t, err, forgejo.ErrStatus)
	assert.Contains(t, err.Error(), "listing the organisations")

	fake, workspaces, projects = instance(t)
	fake.orgs = fake.orgs[:1]
	s, g := newSync(t, fake, workspaces, projects, t.TempDir())
	s.HTTPS = true
	out.Reset()
	assert.NoError(t, s.Run(ctx, &out))
	assert.Contains(t, strings.Join(g.clones, "\n"), "https://git.example/Malta/apps.git main")
	assert.Equal(t, "https://git.example/Malta/malta.git", workspaces.rows["malta"].HomeURL)

	fake.orgs = nil
	out.Reset()
	assert.NoError(t, s.Run(ctx, &out))
	assert.Contains(t, out.String(), "the token's user is in no organisation\n")
	assert.Contains(t, out.String(), "    workspace black: ok\n")
}

// A token without read:user syncs the organisations alone, with one
// line on the users and no error; a repository of an organisation the
// token does not administer gets no hook either.
func TestSync_Users(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	fake, workspaces, projects := instance(t)
	fake.noUserScope = true
	web := fake.repos["Acme"][0]
	web.Permissions.Admin = false
	fake.repos["Acme"][0] = web
	s, g := newSync(t, fake, workspaces, projects, t.TempDir())
	var out bytes.Buffer
	assert.NoError(t, s.Run(ctx, &out))
	report := out.String()
	for _, line := range []string{
		"users skipped: the token lacks read:user",
		"  + workspace malta: created",
		"    skipped hook Acme/web: the token has no admin access",
		"Applied: 1 workspaces created, 1 homes cloned, 3 projects cloned, 3 hooks installed, 2 unchanged, 5 skipped, 1 left alone, over 4 organisations and 0 users.",
	} {
		assert.Contains(t, report, line+"\n")
	}
	assert.Equal(t, 1, strings.Count(report, "users skipped"))
	assert.NotContains(t, report, "black")
	assert.Nil(t, workspaces.rows["black"])
	assert.Len(t, fake.hooksOf("Acme/web"), 0)
	assert.Len(t, g.clones, 4)
}

// fakeGit records the clones and pulls: a clone makes the directory
// with a .git in it, and a pull does nothing.
type fakeGit struct {
	mu            sync.Mutex
	clones, pulls []string
}

func (g *fakeGit) IsRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil && info.IsDir()
}

func (g *fakeGit) Clone(_ context.Context, url, dir, branch string) (string, error) {
	if _, err := os.Stat(dir); err == nil {
		return "", fmt.Errorf("%s exists already", filepath.Base(dir))
	}
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		return "", err
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.clones = append(g.clones, url+" "+branch)
	return "Cloning into " + dir, nil
}

func (g *fakeGit) Pull(_ context.Context, dir string, w io.Writer) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.pulls = append(g.pulls, dir)
	fmt.Fprintln(w, "Already up to date.")
	return nil
}

// errNoRow stands in for storage.ErrNotFound.
var errNoRow = errors.New("not found")

// fakeWorkspaces holds workspace rows by folder.
type fakeWorkspaces struct {
	rows map[string]*model.Workspace
}

func (f *fakeWorkspaces) GetByFolder(_ context.Context, folder string) (*model.Workspace, error) {
	if row, ok := f.rows[folder]; ok {
		copied := *row
		return &copied, nil
	}
	return nil, errNoRow
}

func (f *fakeWorkspaces) Create(_ context.Context, row *model.Workspace) error {
	if _, ok := f.rows[row.Folder]; ok {
		return errors.New("exists")
	}
	row.ID = "w-" + row.Folder
	copied := *row
	f.rows[row.Folder] = &copied
	return nil
}

func (f *fakeWorkspaces) Update(_ context.Context, row *model.Workspace) error {
	if _, ok := f.rows[row.Folder]; !ok {
		return errNoRow
	}
	copied := *row
	f.rows[row.Folder] = &copied
	return nil
}

// fakeProjects holds project rows by folder.
type fakeProjects struct {
	rows map[string]*model.Project
}

func (f *fakeProjects) GetByFolder(_ context.Context, folder string) (*model.Project, error) {
	if row, ok := f.rows[folder]; ok {
		copied := *row
		return &copied, nil
	}
	return nil, errNoRow
}

func (f *fakeProjects) Upsert(_ context.Context, row *model.Project) error {
	copied := *row
	f.rows[row.Folder] = &copied
	return nil
}

func (f *fakeProjects) List(context.Context) ([]model.Project, error) {
	var rows []model.Project
	for _, row := range f.rows {
		rows = append(rows, *row)
	}
	return rows, nil
}
