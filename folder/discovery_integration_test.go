//go:build integration

package folder

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/titpetric/task-ui/model"
)

// A folder git ignores is a build copy or a cache, not a project: the
// sources a dockerfiles build copies into src/ carry their own pipelines.
// A folder with nothing to run (a Taskfile whose tasks have no desc, a
// package.json without scripts) stays out as well.
func TestDiscovery_SkipsIgnoredAndEmpty(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "src", "dockerfiles")
	gitInit(t, repo)
	write(t, filepath.Join(repo, ".gitignore"), "/app/src\n")
	write(t, filepath.Join(repo, "app", "atkins.yml"), "jobs:\n  build: docker build .\n")
	write(t, filepath.Join(repo, "app", "src", "lib", "atkins.yml"), "jobs:\n  test: go test ./...\n")
	write(t, filepath.Join(repo, "assets", "Taskfile.yml"), "tasks:\n  helper:\n    cmds: [echo]\n")
	write(t, filepath.Join(repo, "tool", "package.json"), "{\"name\": \"tool\"}\n")

	d := &Discovery{Root: root, Discover: true}
	ids := []string{}
	for _, f := range d.List() {
		ids = append(ids, f.ID)
	}
	if want := []string{"dockerfiles", "dockerfiles/app"}; !slices.Equal(ids, want) {
		t.Fatalf("folders: got %v, want %v", ids, want)
	}
}

func gitInit(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", dir, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
}

// A checkout with nothing to run at its top, its apps in subfolders, is
// still the top level folder the listing puts those apps under.
func TestDiscovery_ProjectSection(t *testing.T) {
	root := t.TempDir()
	apps := filepath.Join(root, "src", "apps")
	gitInit(t, apps)
	write(t, filepath.Join(apps, "README.md"), "apps")
	write(t, filepath.Join(apps, "blog", "atkins.yml"), "jobs:\n  serve: phpscript serve\n")
	write(t, filepath.Join(apps, "booking", "atkins.yml"), "tasks:\n  default:\n    steps:\n      - task: go:build\n")

	d := &Discovery{Root: root, Discover: true}
	ids := []string{}
	for _, f := range d.List() {
		ids = append(ids, f.ID)
	}
	if want := []string{"apps", "apps/blog", "apps/booking"}; !slices.Equal(ids, want) {
		t.Fatalf("folders: got %v, want %v", ids, want)
	}
	if top := d.Find("apps"); top.Depth() != 0 || len(top.Jobs) != 0 {
		t.Fatalf("apps: %+v", top)
	}
}

// A shallow project lists itself and its top-level folders alone:
// discovery goes one level into it, counted from the project's own
// folder, for one at the root and one in a workspace alike. Another
// checkout, recursive as projects are by default, lists the folders
// inside it as deep as they go.
func TestDiscovery_Shallow(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"src/mono", "src/apps", "ws/src/inner", "ws/src/deep"} {
		dir := filepath.Join(root, filepath.FromSlash(name))
		gitInit(t, dir)
		write(t, filepath.Join(dir, "atkins.yml"), "jobs:\n  all: echo all\n")
		write(t, filepath.Join(dir, "blog", "atkins.yml"), "jobs:\n  serve: echo serve\n")
		write(t, filepath.Join(dir, "blog", "admin", "atkins.yml"), "jobs:\n  admin: echo admin\n")
		write(t, filepath.Join(dir, "lib", "web", "atkins.yml"), "jobs:\n  build: echo build\n")
	}

	d := &Discovery{Root: root, Discover: true, Workspaces: []Config{{ID: "ws"}}, Config: []Config{
		{ID: "mono", Shallow: true}, {ID: "apps"},
		{ID: "ws/inner", Shallow: true}, {ID: "ws/deep"},
	}}
	ids := []string{}
	for _, f := range d.List() {
		ids = append(ids, f.ID)
	}
	want := []string{
		"apps", "apps/blog", "apps/blog/admin", "apps/lib/web",
		"mono", "mono/blog",
		"ws",
		"ws/deep", "ws/deep/blog", "ws/deep/blog/admin", "ws/deep/lib/web",
		"ws/inner", "ws/inner/blog",
	}
	if !slices.Equal(ids, want) {
		t.Fatalf("folders: got %v, want %v", ids, want)
	}
	if mono := d.Find("mono"); mono == nil || mono.Job("all") == nil {
		t.Fatalf("mono's own jobs: %+v", mono)
	}
	if d.Find("mono/blog/admin") != nil || d.Find("ws/inner/lib/web") != nil {
		t.Fatal("a shallow project lists no folder below its top level")
	}
}

// A workspace is its home, a checkout with its own jobs, and the
// projects under src, each a checkout of its own with the folders inside
// it; a configured folder lists though the project's own .gitignore
// ignores it. A directory under the root that is no named workspace
// lists nothing.
func TestDiscovery_Workspace(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	home := filepath.Join(root, "ws", "home")
	gitInit(t, home)
	write(t, filepath.Join(home, "atkins.yml"), "jobs:\n  setup: echo setup\n")
	project := filepath.Join(root, "ws", "src", "project")
	gitInit(t, project)
	write(t, filepath.Join(project, ".gitignore"), "/vendor\n")
	write(t, filepath.Join(project, "atkins.yml"), "jobs:\n  build: echo build\n")
	write(t, filepath.Join(project, "sub", "atkins.yml"), "jobs:\n  test: echo test\n")
	write(t, filepath.Join(project, "vendor", "atkins.yml"), "jobs:\n  tmp: echo tmp\n")
	write(t, filepath.Join(root, "ws", "scratch", "atkins.yml"), "jobs:\n  x: echo x\n")

	list := func(d *Discovery) []string {
		ids := []string{}
		for _, f := range d.List() {
			ids = append(ids, f.ID)
		}
		return ids
	}
	d := &Discovery{Root: root, Discover: true, Workspaces: []Config{{ID: "ws"}}, Config: []Config{{ID: "ws/project/vendor"}}}
	if got, want := list(d), []string{"ws", "ws/home", "ws/project", "ws/project/sub", "ws/project/vendor"}; !slices.Equal(got, want) {
		t.Fatalf("folders: got %v, want %v", got, want)
	}
	if sub := d.Find("ws/project/sub"); sub == nil || sub.Job("test") == nil {
		t.Fatalf("ws/project/sub: %+v", sub)
	}
	if home := d.Find("ws/home"); home == nil || home.Job("setup") == nil {
		t.Fatalf("ws/home: %+v", home)
	}
	// Unconfigured, the ignored folder stays out.
	plain := &Discovery{Root: root, Discover: true, Workspaces: []Config{{ID: "ws"}}}
	if got, want := list(plain), []string{"ws", "ws/home", "ws/project", "ws/project/sub"}; !slices.Equal(got, want) {
		t.Fatalf("folders: got %v, want %v", got, want)
	}
	// Without the workspace named, nothing under its directory lists.
	if got := list(&Discovery{Root: root, Discover: true}); len(got) != 0 {
		t.Fatalf("folders: got %v, want none", got)
	}
}

// A directory git ignores right below a workspace's src, when src is a
// checkout, lists when it holds a job source, on top or further in, such
// as another checkout; the folders inside it follow the workspace's
// Recursive. One with no job source stays out, as does an ignored
// directory deeper down. A project row on the directory wins over the
// discovered one.
func TestDiscovery_IgnoredCheckouts(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	ws := filepath.Join(root, "ws", "src")
	gitInit(t, ws)
	write(t, filepath.Join(ws, ".gitignore"), "/deploy\n/empty\n/lib/cache\n/ops\n")
	write(t, filepath.Join(ws, "atkins.yml"), "jobs:\n  setup: echo setup\n")
	deploy := filepath.Join(ws, "deploy")
	gitInit(t, deploy)
	write(t, filepath.Join(deploy, "Taskfile.yml"), "tasks:\n  up:\n    desc: Up\n")
	write(t, filepath.Join(deploy, "web", "atkins.yml"), "jobs:\n  ship: echo ship\n")
	write(t, filepath.Join(ws, "empty", "inner", "README.md"), "nothing to run\n")
	ops := filepath.Join(ws, "ops")
	gitInit(t, ops)
	write(t, filepath.Join(ops, "scripts", "atkins.yml"), "jobs:\n  apply: echo apply\n")
	write(t, filepath.Join(ws, "lib", "cache", "atkins.yml"), "jobs:\n  y: echo y\n")

	list := func(d *Discovery) []string {
		ids := []string{}
		for _, f := range d.List() {
			ids = append(ids, f.ID)
		}
		return ids
	}
	d := &Discovery{Root: root, Discover: true, Workspaces: []Config{{ID: "ws"}}}
	if got, want := list(d), []string{"ws", "ws/deploy", "ws/deploy/web", "ws/ops/scripts"}; !slices.Equal(got, want) {
		t.Fatalf("folders: got %v, want %v", got, want)
	}
	if f := d.Find("ws/deploy"); f == nil || f.Job("up") == nil {
		t.Fatalf("ws/deploy: %+v", f)
	}
	if f := d.Find("ws/ops/scripts"); f == nil || f.Job("apply") == nil {
		t.Fatalf("ws/ops/scripts: %+v", f)
	}

	shallow := &Discovery{Root: root, Discover: true, Workspaces: []Config{{ID: "ws", Shallow: true}}}
	if got, want := list(shallow), []string{"ws", "ws/deploy"}; !slices.Equal(got, want) {
		t.Fatalf("shallow: got %v, want %v", got, want)
	}

	configured := &Discovery{Root: root, Discover: true, Workspaces: []Config{{ID: "ws"}}, Config: []Config{{ID: "ws/deploy", Title: "Deploy"}}}
	if f := configured.Find("ws/deploy"); f == nil || f.Title != "Deploy" {
		t.Fatalf("project row: %+v", f)
	}

	// Every folder of a workspace runs in its hours; the default
	// workspace's hours stay with its own folders.
	gated := &Discovery{Root: root, Discover: true, Workspaces: []Config{{ID: "ws", Hours: "mon-fri 9-17"}, {ID: "_", Hours: "sun 1-2"}}}
	for _, f := range gated.List() {
		if f.Hours != "mon-fri 9-17" {
			t.Fatalf("hours of %s: %q", f.ID, f.Hours)
		}
	}
	if gated.HoursOf("_") != "sun 1-2" || gated.HoursOf("nope") != "" {
		t.Fatalf("hours of the workspaces: %q", gated.HoursOf("_"))
	}

	// The default workspace is the root: its projects are under src.
	other := t.TempDir()
	gitInit(t, filepath.Join(other, "src"))
	write(t, filepath.Join(other, "src", ".gitignore"), "/tools\n")
	write(t, filepath.Join(other, "src", "tools", "Makefile"), "## Build\nbuild:\n\techo build\n")
	top := &Discovery{Root: other, Discover: true, Workspaces: WorkspaceConfigsOf([]model.Workspace{{Folder: model.Default}})}
	if got, want := list(top), []string{"tools"}; !slices.Equal(got, want) {
		t.Fatalf("root: got %v, want %v", got, want)
	}
}
