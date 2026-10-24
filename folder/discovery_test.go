package folder

import (
	"path/filepath"
	"slices"
	"testing"
)

// src is the src directory of the default workspace under a root.
func src(root string, path ...string) string {
	return filepath.Join(append([]string{root, "src"}, path...)...)
}

func TestDiscovery(t *testing.T) {
	root := t.TempDir()
	write(t, src(root, "infra", "atkins.yml"), "name: Infra\njobs:\n  plan: echo plan\n")
	// Atkins files win over Taskfiles in the same folder.
	write(t, src(root, "infra", "Taskfile.yml"), "tasks:\n  other:\n    desc: Other\n")
	write(t, src(root, "web site", "Taskfile.yml"), "tasks:\n  build:\n    desc: Build\n")
	write(t, src(root, ".hidden", "Taskfile.yml"), "tasks:\n  x:\n    desc: X\n")
	write(t, src(root, "empty", "README.md"), "nothing")
	write(t, src(root, "elsewhere", "jobs.yml"), "jobs:\n  sync:\n    desc: Sync\n")
	// The root holds nothing to run of its own: its home and src are the folders.
	write(t, filepath.Join(root, "Taskfile.yml"), "tasks:\n  root:\n    desc: Root\n")

	d := &Discovery{
		Root:     root,
		Discover: true,
		Config: []Config{
			{ID: "web", Path: "src/web site", Title: "Website", Allow: []string{"ops"}, Command: "make", Jobs: map[string][]string{"build": {"admin"}}},
			{ID: "sync", Path: "src/elsewhere/jobs.yml"},
			{ID: "ghost", Path: "ghost"},
			{ID: "holder", Path: "src/empty"},
		},
	}
	list := d.List()

	ids := []string{}
	for _, f := range list {
		ids = append(ids, f.ID)
	}
	slices.Sort(ids)
	if want := []string{"ghost", "holder", "infra", "sync", "web"}; !slices.Equal(ids, want) {
		t.Fatalf("folders: got %v, want %v", ids, want)
	}

	infra := d.Find("infra")
	if len(infra.Sources) != 1 || infra.Sources[0].Kind != "atkins" || infra.Job("plan") == nil || infra.Job("other") != nil {
		t.Fatalf("infra: %+v", infra)
	}
	if infra.Dir != src(root, "infra") {
		t.Fatalf("infra lives under src: %s", infra.Dir)
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

// A project row names its directory by its ID alone: a project of _
// under <root>/src, one of a named workspace under <root>/<ws>/src.
func TestDiscovery_ConfiguredDirs(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, src(root, "infra", "atkins.yml"), "jobs:\n  plan: echo plan\n")
	write(t, filepath.Join(root, "malta", "src", "deploy", "atkins.yml"), "jobs:\n  up: echo up\n")

	d := &Discovery{Root: root, Discover: false, Workspaces: []Config{{ID: "_"}, {ID: "malta"}}, Config: []Config{{ID: "infra", Title: "Infra"}, {ID: "malta/deploy", Title: "Deploy"}}}
	if f := d.Find("infra"); f == nil || f.Dir != src(root, "infra") || f.Job("plan") == nil {
		t.Fatalf("infra: %+v", f)
	}
	if f := d.Find("malta/deploy"); f == nil || f.Dir != filepath.Join(root, "malta", "src", "deploy") || f.Job("up") == nil {
		t.Fatalf("malta/deploy: %+v", f)
	}
}

// Discovery goes deep: it lists a nested folder with a pipeline file
// at its path, in tree order. Under a pipeline folder the build files of
// a source tree are not folders; under a plain folder they are.
func TestDiscovery_Tree(t *testing.T) {
	root := t.TempDir()
	write(t, src(root, "services", "api", "atkins.yml"), "jobs:\n  deploy: echo api\n")
	write(t, src(root, "services", "api", "edge", "atkins.yml"), "jobs:\n  deploy: echo edge\n")
	write(t, src(root, "services", "api", "internal", "Makefile"), "build:\n\tgo build\n")
	write(t, src(root, "services", "api", "cmd", "tool", "Makefile"), "build:\n\tgo build\n")
	write(t, src(root, "node_modules", "dep", "atkins.yml"), "jobs:\n  x: echo x\n")

	d := &Discovery{Root: root, Discover: true}
	ids := []string{}
	for _, f := range d.List() {
		ids = append(ids, f.ID)
	}
	// services has nothing to run, but holds folders that do: the listing
	// shows it as their top level folder, the sidebar's section for them.
	if want := []string{"services", "services/api", "services/api/edge"}; !slices.Equal(ids, want) {
		t.Fatalf("tree: got %v, want %v", ids, want)
	}
	edge := d.Find("services/api/edge")
	if edge == nil || edge.Name() != "edge" || edge.Depth() != 2 || edge.Parent() != "services/api" || edge.Job("deploy") == nil {
		t.Fatalf("edge: %+v", edge)
	}
	if api := d.Find("services/api"); api.Depth() != 1 || api.Parent() != "services" || api.Name() != "api" {
		t.Fatalf("api: %+v", api)
	}
	if top := d.Find("services"); top.Depth() != 0 || top.Parent() != "" || top.Dir != src(root, "services") {
		t.Fatalf("services: %+v", top)
	}

	// A plain project: the build files are folders, as deep as they go.
	plain := t.TempDir()
	write(t, src(plain, "lib", "Makefile"), "build:\n\tgo build\n")
	write(t, src(plain, "lib", "sub", "scripts", "run.sh"), "# Run\n")
	write(t, src(plain, "lib", "sub", "deeper", "atkins.yml"), "jobs:\n  x: echo x\n")
	ids = ids[:0]
	for _, f := range (&Discovery{Root: plain, Discover: true}).List() {
		ids = append(ids, f.ID)
	}
	if want := []string{"lib", "lib/sub", "lib/sub/deeper"}; !slices.Equal(ids, want) {
		t.Fatalf("plain tree: got %v, want %v", ids, want)
	}
}

// Each workspace's home is a folder, home for _ and <ws>/home, listed
// with or without something to run; its projects live under src and
// take the ID <ws>/<project>. Nothing beside home and src lists.
func TestDiscovery_Workspaces(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, filepath.Join(root, "home", "README.md"), "the default workspace's home, empty")
	write(t, src(root, "infra", "atkins.yml"), "jobs:\n  plan: echo plan\n")
	write(t, filepath.Join(root, "malta", "home", "atkins.yml"), "jobs:\n  setup: echo setup\n")
	write(t, filepath.Join(root, "malta", "home", ".atkins", "skills", "git.yml"), "jobs:\n  pull: git pull\n")
	write(t, filepath.Join(root, "malta", "src", "deploy", "atkins.yml"), "jobs:\n  up: echo up\n")
	write(t, filepath.Join(root, "malta", "src", "deploy", "web", "atkins.yml"), "jobs:\n  ship: echo ship\n")
	write(t, filepath.Join(root, "malta", "notes", "atkins.yml"), "jobs:\n  x: echo x\n")
	write(t, filepath.Join(root, "other", "atkins.yml"), "jobs:\n  x: echo x\n")

	d := &Discovery{Root: root, Discover: true, Workspaces: []Config{{ID: "_"}, {ID: "malta"}}}
	ids := []string{}
	for _, f := range d.List() {
		ids = append(ids, f.ID)
	}
	if want := []string{"home", "infra", "malta", "malta/deploy", "malta/deploy/web", "malta/home"}; !slices.Equal(ids, want) {
		t.Fatalf("folders: got %v, want %v", ids, want)
	}
	if home := d.Find("malta/home"); home == nil || home.Dir != filepath.Join(root, "malta", "home") || home.Job("setup") == nil || home.Depth() != 1 {
		t.Fatalf("malta/home: %+v", home)
	}
	if home := d.Find("home"); home == nil || home.Dir != filepath.Join(root, "home") || len(home.Jobs) != 0 {
		t.Fatalf("home: %+v", home)
	}
	if ws := d.Find("malta"); ws == nil || ws.Dir != filepath.Join(root, "malta", "src") {
		t.Fatalf("malta: %+v", ws)
	}
	if deploy := d.Find("malta/deploy"); deploy == nil || deploy.Dir != filepath.Join(root, "malta", "src", "deploy") || deploy.Depth() != 1 {
		t.Fatalf("malta/deploy: %+v", deploy)
	}

	// A shallow workspace lists its projects and no folder inside them.
	shallow := &Discovery{Root: root, Discover: true, Workspaces: []Config{{ID: "malta", Shallow: true}}}
	ids = ids[:0]
	for _, f := range shallow.List() {
		ids = append(ids, f.ID)
	}
	if want := []string{"home", "infra", "malta", "malta/deploy", "malta/home"}; !slices.Equal(ids, want) {
		t.Fatalf("shallow: got %v, want %v", ids, want)
	}

	// Every lists a home only when it has a job, as any folder.
	every := &Discovery{Root: root, Discover: true, Workspaces: []Config{{ID: "malta"}}, Every: true}
	ids = ids[:0]
	for _, f := range every.List() {
		ids = append(ids, f.ID)
	}
	if want := []string{"infra", "malta/deploy", "malta/deploy/web", "malta/home"}; !slices.Equal(ids, want) {
		t.Fatalf("every: got %v, want %v", ids, want)
	}
}

// A directory of src named internal or _ lists nothing, nor do the
// folders inside it: the names are task-ui's own. Deeper down they are
// ordinary ones.
func TestDiscovery_SkipsInternal(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write(t, src(root, "internal", "Taskfile.yml"), "tasks:\n  x:\n    desc: X\n")
	write(t, src(root, "internal", "tools", "Taskfile.yml"), "tasks:\n  y:\n    desc: Y\n")
	write(t, src(root, "_", "Taskfile.yml"), "tasks:\n  u:\n    desc: U\n")
	write(t, src(root, "_", "web", "Taskfile.yml"), "tasks:\n  w:\n    desc: W\n")
	write(t, src(root, "deploy", "internal", "Taskfile.yml"), "tasks:\n  z:\n    desc: Z\n")
	write(t, src(root, "deploy", "_", "Taskfile.yml"), "tasks:\n  v:\n    desc: V\n")

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
	write(t, src(root, "app", ".git", "HEAD"), "ref: refs/heads/main\n")
	write(t, src(root, "app", "Taskfile.yml"), "tasks:\n  build:\n    desc: Build\n")
	write(t, src(root, "app", "dist", "Taskfile.yml"), "tasks:\n  copy:\n    desc: Copy\n")
	write(t, src(root, "app", "tools", "Taskfile.yml"), "tasks:\n  lint:\n    desc: Lint\n")

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
	if want := []string{src(root, "app")}; !slices.Equal(asked, want) {
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

// everyTree is a root with a workspace whose project ignores a deep
// checkout with jobs, a shallow project with a nested folder and a
// folder with nothing to run.
func everyTree(t *testing.T) (string, func(string) []string) {
	t.Helper()
	root := t.TempDir()
	const task = "tasks:\n  build:\n    desc: Build\n"
	write(t, filepath.Join(root, "ws", "src", "tooling", ".git", "HEAD"), "ref: main\n")
	write(t, filepath.Join(root, "ws", "src", "tooling", "Taskfile.yml"), task)
	write(t, filepath.Join(root, "ws", "src", "tooling", "lib", "app", "Taskfile.yml"), task)
	write(t, filepath.Join(root, "ws", "src", "tooling", "docs", "README.md"), "nothing to run")
	write(t, src(root, "proj", "Taskfile.yml"), task)
	write(t, src(root, "proj", "sub", "deep", "Taskfile.yml"), task)
	ignored := func(dir string) []string {
		if dir == filepath.Join(root, "ws", "src", "tooling") {
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
	if got, want := folderIDs(sidebar.List()), []string{"proj", "ws", "ws/tooling"}; !slices.Equal(got, want) {
		t.Fatalf("settings listing: got %v, want %v", got, want)
	}

	every := &Discovery{Root: root, Discover: true, Config: config, Workspaces: workspaces, Ignored: ignored, Every: true}
	if got, want := folderIDs(every.List()), []string{"proj", "proj/sub/deep", "ws/tooling", "ws/tooling/lib/app"}; !slices.Equal(got, want) {
		t.Fatalf("every listing: got %v, want %v", got, want)
	}
	app := every.Find("ws/tooling/lib/app")
	if app == nil || app.Job("build") == nil {
		t.Fatalf("ignored checkout's jobs: %+v", app)
	}
	if proj := every.Find("proj"); proj == nil || !slices.Equal(proj.Allow, []string{"ops"}) {
		t.Fatalf("project configuration not applied: %+v", proj)
	}
	if every.Find("ws/tooling/docs") != nil {
		t.Fatal("a folder without jobs lists")
	}
}
