package folder

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/titpetric/task-ui/model"
)

func TestDiscovery(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "Taskfile.yml"), "tasks:\n  root:\n    desc: Root\n")
	write(t, filepath.Join(root, "infra", "atkins.yml"), "name: Infra\njobs:\n  plan: echo plan\n")
	// Atkins files win over Taskfiles in the same folder.
	write(t, filepath.Join(root, "infra", "Taskfile.yml"), "tasks:\n  other:\n    desc: Other\n")
	write(t, filepath.Join(root, "web site", "Taskfile.yml"), "tasks:\n  build:\n    desc: Build\n")
	write(t, filepath.Join(root, ".hidden", "Taskfile.yml"), "tasks:\n  x:\n    desc: X\n")
	write(t, filepath.Join(root, "empty", "README.md"), "nothing")
	write(t, filepath.Join(root, "elsewhere", "jobs.yml"), "jobs:\n  sync:\n    desc: Sync\n")

	d := &Discovery{
		Root:     root,
		Discover: true,
		Config: []Config{
			{ID: "web", Path: "web site", Title: "Website", Allow: []string{"ops"}, Command: "make", Jobs: map[string][]string{"build": {"admin"}}},
			{ID: "sync", Path: "elsewhere/jobs.yml"},
			{ID: "ghost", Path: "ghost"},
			{ID: "holder", Path: "empty"},
		},
	}
	list := d.List()

	ids := []string{}
	for _, f := range list {
		ids = append(ids, f.ID)
	}
	slices.Sort(ids)
	if want := []string{filepath.Base(root), "ghost", "holder", "infra", "sync", "web"}; !slices.Equal(ids, want) {
		t.Fatalf("folders: got %v, want %v", ids, want)
	}

	infra := d.Find("infra")
	if len(infra.Sources) != 1 || infra.Sources[0].Kind != "atkins" || infra.Job("plan") == nil || infra.Job("other") != nil {
		t.Fatalf("infra: %+v", infra)
	}
	web := d.Find("web")
	if web.Title != "Website" || !slices.Equal(web.Allow, []string{"ops"}) || !slices.Equal(web.Job("build").Command, []string{"make", "build"}) {
		t.Fatalf("web overrides not applied: %+v", web)
	}
	if !slices.Equal(web.Job("build").Allow, []string{"admin"}) {
		t.Fatalf("job allow not applied: %+v", web.Job("build"))
	}
	if sync := d.Find("sync"); sync == nil || sync.Job("sync") == nil {
		t.Fatalf("configured file folder missing: %+v", sync)
	}
	if ghost := d.Find("ghost"); ghost == nil || ghost.Error == "" {
		t.Fatalf("configured folder that is not there: %+v", ghost)
	}
	// A directory with nothing to run of its own holds folders: no error.
	if holder := d.Find("holder"); holder == nil || holder.Error != "" {
		t.Fatalf("configured directory with nothing to run: %+v", holder)
	}
	if d.Find("nope") != nil {
		t.Fatal("missing folder found")
	}

	quiet := &Discovery{Root: root, Discover: false}
	if len(quiet.List()) != 0 {
		t.Fatal("discovery off still listed folders")
	}
}

// Discovery goes deep: it lists a nested folder with a pipeline file
// at its path, in tree order. Under a pipeline folder the build files of
// a source tree are not folders; under a plain folder they are.
func TestDiscovery_Tree(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "atkins.yml"), "name: Root\njobs:\n  all: echo all\n")
	write(t, filepath.Join(root, "services", "api", "atkins.yml"), "jobs:\n  deploy: echo api\n")
	write(t, filepath.Join(root, "services", "api", "edge", "atkins.yml"), "jobs:\n  deploy: echo edge\n")
	write(t, filepath.Join(root, "services", "api", "internal", "Makefile"), "build:\n\tgo build\n")
	write(t, filepath.Join(root, "cmd", "tool", "Makefile"), "build:\n\tgo build\n")
	write(t, filepath.Join(root, "node_modules", "dep", "atkins.yml"), "jobs:\n  x: echo x\n")

	d := &Discovery{Root: root, Discover: true}
	ids := []string{}
	for _, f := range d.List() {
		ids = append(ids, f.ID)
	}
	// services has nothing to run, but holds folders that do: the listing
	// shows it as their top level folder, the sidebar's section for them.
	if want := []string{filepath.Base(root), "services", "services/api", "services/api/edge"}; !slices.Equal(ids, want) {
		t.Fatalf("tree: got %v, want %v", ids, want)
	}
	edge := d.Find("services/api/edge")
	if edge == nil || edge.Name() != "edge" || edge.Depth() != 2 || edge.Parent() != "services/api" || edge.Job("deploy") == nil {
		t.Fatalf("edge: %+v", edge)
	}
	if api := d.Find("services/api"); api.Depth() != 1 || api.Parent() != "services" || api.Name() != "api" {
		t.Fatalf("api: %+v", api)
	}
	if top := d.Find(filepath.Base(root)); top.Depth() != 0 || top.Parent() != "" {
		t.Fatalf("root: %+v", top)
	}

	// A plain root: the build files are folders, as deep as they go.
	plain := t.TempDir()
	write(t, filepath.Join(plain, "lib", "Makefile"), "build:\n\tgo build\n")
	write(t, filepath.Join(plain, "lib", "sub", "scripts", "run.sh"), "# Run\n")
	write(t, filepath.Join(plain, "lib", "sub", "deeper", "atkins.yml"), "jobs:\n  x: echo x\n")
	ids = ids[:0]
	for _, f := range (&Discovery{Root: plain, Discover: true}).List() {
		ids = append(ids, f.ID)
	}
	if want := []string{"lib", "lib/sub", "lib/sub/deeper"}; !slices.Equal(ids, want) {
		t.Fatalf("plain tree: got %v, want %v", ids, want)
	}
}

// A folder git ignores is a build copy or a cache, not a project: the
// sources a dockerfiles build copies into src/ carry their own pipelines.
// A folder with nothing to run (a Taskfile whose tasks have no desc, a
// package.json without scripts) stays out as well.
func TestDiscovery_SkipsIgnoredAndEmpty(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "dockerfiles")
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
	apps := filepath.Join(root, "apps")
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
	for _, name := range []string{"mono", "apps", "ws/inner", "ws/deep"} {
		dir := filepath.Join(root, filepath.FromSlash(name))
		gitInit(t, dir)
		write(t, filepath.Join(dir, "atkins.yml"), "jobs:\n  all: echo all\n")
		write(t, filepath.Join(dir, "blog", "atkins.yml"), "jobs:\n  serve: echo serve\n")
		write(t, filepath.Join(dir, "blog", "admin", "atkins.yml"), "jobs:\n  admin: echo admin\n")
		write(t, filepath.Join(dir, "lib", "web", "atkins.yml"), "jobs:\n  build: echo build\n")
	}

	d := &Discovery{Root: root, Discover: true, Config: []Config{
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

// A directory of the root named internal or _ lists nothing, nor do the
// folders inside it: the names are task-ui's own. Deeper down they are
// ordinary ones.
func TestDiscovery_SkipsInternal(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, filepath.Join(root, "internal", "Taskfile.yml"), "tasks:\n  x:\n    desc: X\n")
	write(t, filepath.Join(root, "internal", "tools", "Taskfile.yml"), "tasks:\n  y:\n    desc: Y\n")
	write(t, filepath.Join(root, "_", "Taskfile.yml"), "tasks:\n  u:\n    desc: U\n")
	write(t, filepath.Join(root, "_", "web", "Taskfile.yml"), "tasks:\n  w:\n    desc: W\n")
	write(t, filepath.Join(root, "deploy", "internal", "Taskfile.yml"), "tasks:\n  z:\n    desc: Z\n")
	write(t, filepath.Join(root, "deploy", "_", "Taskfile.yml"), "tasks:\n  v:\n    desc: V\n")

	d := &Discovery{Root: root, Discover: true}
	ids := []string{}
	for _, f := range d.List() {
		ids = append(ids, f.ID)
	}
	if want := []string{"deploy", "deploy/_", "deploy/internal"}; !slices.Equal(ids, want) {
		t.Fatalf("folders: got %v, want %v", ids, want)
	}
}

// A project cloned inside a workspace lists, with the folders inside it,
// though the workspace's .gitignore ignores it: the walk goes into every
// configured folder.
func TestDiscovery_Workspace(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	ws := filepath.Join(root, "ws")
	gitInit(t, ws)
	write(t, filepath.Join(ws, ".gitignore"), "/project\n/scratch\n")
	write(t, filepath.Join(ws, "atkins.yml"), "jobs:\n  setup: echo setup\n")
	project := filepath.Join(ws, "project")
	gitInit(t, project)
	write(t, filepath.Join(project, "atkins.yml"), "jobs:\n  build: echo build\n")
	write(t, filepath.Join(project, "sub", "atkins.yml"), "jobs:\n  test: echo test\n")
	write(t, filepath.Join(ws, "scratch", "atkins.yml"), "jobs:\n  tmp: echo tmp\n")

	list := func(d *Discovery) []string {
		ids := []string{}
		for _, f := range d.List() {
			ids = append(ids, f.ID)
		}
		return ids
	}
	d := &Discovery{Root: root, Discover: true, Config: []Config{{ID: "ws/project"}}}
	if got, want := list(d), []string{"ws", "ws/project", "ws/project/sub"}; !slices.Equal(got, want) {
		t.Fatalf("folders: got %v, want %v", got, want)
	}
	if sub := d.Find("ws/project/sub"); sub == nil || sub.Job("test") == nil {
		t.Fatalf("ws/project/sub: %+v", sub)
	}
	// Unconfigured, the ignored project stays out.
	if got, want := list(&Discovery{Root: root, Discover: true}), []string{"ws"}; !slices.Equal(got, want) {
		t.Fatalf("folders: got %v, want %v", got, want)
	}
}

// A directory a workspace's checkout ignores, right below it, lists
// when it holds a job source, on top or further in, such as another
// checkout; the folders inside it follow the workspace's Recursive. One
// with no job source stays out, as does an ignored directory deeper
// down. A project row on the directory wins over the discovered one.
func TestDiscovery_IgnoredCheckouts(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	ws := filepath.Join(root, "ws")
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

	// The default workspace is the root, by Path.
	other := t.TempDir()
	gitInit(t, other)
	write(t, filepath.Join(other, ".gitignore"), "/tools\n")
	write(t, filepath.Join(other, "tools", "Makefile"), "## Build\nbuild:\n\techo build\n")
	top := &Discovery{Root: other, Discover: true, Workspaces: WorkspaceConfigsOf([]model.Workspace{{Folder: model.Default}})}
	if got, want := list(top), []string{"tools"}; !slices.Equal(got, want) {
		t.Fatalf("root: got %v, want %v", got, want)
	}
}
