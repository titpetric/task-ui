//go:build integration

package folder

import (
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/titpetric/task-ui/git"
)

func TestEntries(t *testing.T) {
	top := t.TempDir()
	root := filepath.Join(top, "src")
	write(t, filepath.Join(root, "tools", "Makefile"), "build:\n\tgo build\n")
	write(t, filepath.Join(root, "tools", "package.json"), `{"scripts": {"dev": "vite"}}`)
	write(t, filepath.Join(root, "docs", "README.md"), "nothing to run")
	write(t, filepath.Join(root, "broken", "atkins.yml"), "jobs: [")
	write(t, filepath.Join(root, ".git", "HEAD"), "")
	write(t, filepath.Join(top, "home", "atkins.yml"), "jobs:\n  setup: echo setup\n")
	// A project whose jobs are all in its subfolders.
	write(t, filepath.Join(root, "apps", "README.md"), "apps")
	write(t, filepath.Join(root, "apps", "blog", "atkins.yml"), "jobs:\n  serve: phpscript serve\n  test: go test ./...\n")
	write(t, filepath.Join(root, "apps", "booking", "atkins.yml"), "jobs:\n  build: go build\n")

	repo := filepath.Join(root, "repo")
	write(t, filepath.Join(repo, "atkins.yml"), "jobs:\n  a:\n    desc: A\n")
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "first"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	entries := Entries(top, &git.Git{}, nil)
	byID := map[string]Entry{}
	for _, entry := range entries {
		byID[entry.ID] = entry
	}
	if len(entries) != 5 {
		t.Fatalf("entries: the folders of src, not the home: %+v", entries)
	}
	tools := byID["tools"]
	// The Makefile reads tools; its package.json is the Makefile's business.
	if !slices.Equal(tools.Sources, []string{"Makefile"}) || tools.Jobs != 1 || tools.Checkout != nil {
		t.Fatalf("tools: %+v", tools)
	}
	if apps := byID["apps"]; apps.Jobs != 3 {
		t.Fatalf("apps counts the jobs of its subfolders: %+v", apps)
	}
	if docs := byID["docs"]; len(docs.Sources) != 0 || docs.Jobs != 0 {
		t.Fatalf("docs: %+v", docs)
	}
	if broken := byID["broken"]; broken.Error == "" {
		t.Fatalf("broken: %+v", broken)
	}
	if entry := byID["repo"]; entry.Checkout == nil || entry.Checkout.Subject != "first" || entry.Jobs != 1 {
		t.Fatalf("repo: %+v", entry)
	}
}

// A workspace's checkouts under src list as entries of their own,
// ws/name; its home, with its jobs, is no project and does not.
func TestEntries_Workspace(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	ws := filepath.Join(root, "malta", "src")
	home := filepath.Join(root, "malta", "home")
	gitInit(t, home)
	write(t, filepath.Join(home, "atkins.yml"), "jobs:\n  setup: echo setup\n")
	write(t, filepath.Join(home, "policies", "README.md"), "policies")
	project := filepath.Join(ws, "dockerfiles")
	gitInit(t, project)
	write(t, filepath.Join(project, "atkins.yml"), "jobs:\n  build: echo build\n")
	write(t, filepath.Join(project, "task", "atkins.yml"), "jobs:\n  push: echo push\n")

	byID := map[string]Entry{}
	ids := []string{}
	for _, entry := range Entries(root, &git.Git{}, []string{"malta"}) {
		byID[entry.ID] = entry
		ids = append(ids, entry.ID)
	}
	if want := []string{"malta", "malta/dockerfiles"}; !slices.Equal(ids, want) {
		t.Fatalf("entries: got %v, want %v", ids, want)
	}
	if entry := byID["malta/dockerfiles"]; entry.Dir != project || entry.Jobs != 2 || entry.Checkout == nil {
		t.Fatalf("malta/dockerfiles: %+v", entry)
	}
	if entry := byID["malta"]; entry.Jobs != 2 {
		t.Fatalf("malta counts its projects' jobs: %+v", entry)
	}
	if ids := Entries(root, &git.Git{}, nil); len(ids) != 0 {
		t.Fatalf("without the workspace: %+v", ids)
	}
}
