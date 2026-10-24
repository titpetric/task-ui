package folder

import (
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

// Ignored names what a checkout ignores, asked of a directory with a
// .git alone: an ignored build copy below a checkout stays out.
func TestDiscovery_Ignored(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "app", ".git", "HEAD"), "ref: refs/heads/main\n")
	write(t, filepath.Join(root, "app", "Taskfile.yml"), "tasks:\n  build:\n    desc: Build\n")
	write(t, filepath.Join(root, "app", "dist", "Taskfile.yml"), "tasks:\n  copy:\n    desc: Copy\n")
	write(t, filepath.Join(root, "app", "tools", "Taskfile.yml"), "tasks:\n  lint:\n    desc: Lint\n")

	var asked []string
	d := &Discovery{Root: root, Discover: true, Ignored: func(dir string) []string {
		asked = append(asked, dir)
		return []string{"dist"}
	}}
	ids := []string{}
	for _, f := range d.List() {
		ids = append(ids, f.ID)
	}
	slices.Sort(ids)
	if want := []string{"app", "app/tools"}; !slices.Equal(ids, want) {
		t.Fatalf("folders: got %v, want %v", ids, want)
	}
	if want := []string{filepath.Join(root, "app")}; !slices.Equal(asked, want) {
		t.Fatalf("asked: got %v, want %v", asked, want)
	}
}

func folderIDs(list []*Folder) []string {
	result := []string{}
	for _, f := range list {
		result = append(result, f.ID)
	}
	slices.Sort(result)
	return result
}

// everyTree is a root with a workspace whose checkout ignores a deep
// checkout with jobs, a shallow project with a nested folder and a
// folder with nothing to run.
func everyTree(t *testing.T) (string, func(string) []string) {
	t.Helper()
	root := t.TempDir()
	const task = "tasks:\n  build:\n    desc: Build\n"
	write(t, filepath.Join(root, "ws", ".git", "HEAD"), "ref: main\n")
	write(t, filepath.Join(root, "ws", "Taskfile.yml"), task)
	write(t, filepath.Join(root, "ws", "lib", "app", "Taskfile.yml"), task)
	write(t, filepath.Join(root, "ws", "docs", "README.md"), "nothing to run")
	write(t, filepath.Join(root, "proj", "Taskfile.yml"), task)
	write(t, filepath.Join(root, "proj", "sub", "deep", "Taskfile.yml"), task)
	ignored := func(dir string) []string {
		if dir == filepath.Join(root, "ws") {
			return []string{"lib/app"}
		}
		return nil
	}
	return root, ignored
}

func TestDiscoveryEvery(t *testing.T) {
	t.Parallel()
	root, ignored := everyTree(t)
	config := []Config{{ID: "proj", Shallow: true, Allow: []string{"ops"}}}
	workspaces := []Config{{ID: "ws"}}

	sidebar := &Discovery{Root: root, Discover: true, Config: config, Workspaces: workspaces, Ignored: ignored}
	if got, want := folderIDs(sidebar.List()), []string{"proj", "ws"}; !slices.Equal(got, want) {
		t.Fatalf("settings listing: got %v, want %v", got, want)
	}

	every := &Discovery{Root: root, Discover: true, Config: config, Workspaces: workspaces, Ignored: ignored, Every: true}
	if got, want := folderIDs(every.List()), []string{"proj", "proj/sub/deep", "ws", "ws/lib/app"}; !slices.Equal(got, want) {
		t.Fatalf("every listing: got %v, want %v", got, want)
	}
	app := every.Find("ws/lib/app")
	if app == nil || app.Job("build") == nil {
		t.Fatalf("ignored checkout's jobs: %+v", app)
	}
	if proj := every.Find("proj"); proj == nil || !slices.Equal(proj.Allow, []string{"ops"}) {
		t.Fatalf("project configuration not applied: %+v", proj)
	}
	if every.Find("ws/docs") != nil {
		t.Fatal("a folder without jobs lists")
	}
}
