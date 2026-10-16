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
