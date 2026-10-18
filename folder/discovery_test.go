package folder

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
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
		},
	}
	list := d.List()

	ids := []string{}
	for _, f := range list {
		ids = append(ids, f.ID)
	}
	slices.Sort(ids)
	if want := []string{filepath.Base(root), "ghost", "infra", "sync", "web"}; !slices.Equal(ids, want) {
		t.Fatalf("folders: got %v, want %v", ids, want)
	}

	infra := d.Find("infra")
	if len(infra.Sources) != 2 || infra.Sources[0].Kind != "atkins" || infra.Job("atkins:plan") == nil {
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
		t.Fatalf("configured folder with nothing to run: %+v", ghost)
	}
	if d.Find("nope") != nil {
		t.Fatal("missing folder found")
	}

	quiet := &Discovery{Root: root, Discover: false}
	if len(quiet.List()) != 0 {
		t.Fatal("discovery off still listed folders")
	}
}

// Discovery goes deep: a nested folder with a pipeline file is listed
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
	// services has nothing to run, but holds folders that do: it is listed
	// as their top level folder, the sidebar's section for them.
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
// package.json without scripts) is left out as well.
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
// still the top level folder those apps are listed under.
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
