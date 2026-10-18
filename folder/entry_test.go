package folder

import (
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/titpetric/task-ui/git"
)

func TestEntries(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "tools", "Makefile"), "build:\n\tgo build\n")
	write(t, filepath.Join(root, "tools", "package.json"), `{"scripts": {"dev": "vite"}}`)
	write(t, filepath.Join(root, "docs", "README.md"), "nothing to run")
	write(t, filepath.Join(root, "broken", "atkins.yml"), "jobs: [")
	write(t, filepath.Join(root, ".git", "HEAD"), "")

	repo := filepath.Join(root, "repo")
	write(t, filepath.Join(repo, "atkins.yml"), "jobs:\n  a:\n    desc: A\n")
	for _, args := range [][]string{{"init", "-q"}, {"add", "."}, {"-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "first"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	entries := Entries(root, &git.Git{})
	byID := map[string]Entry{}
	for _, entry := range entries {
		byID[entry.ID] = entry
	}
	if len(entries) != 4 {
		t.Fatalf("entries: %+v", entries)
	}
	tools := byID["tools"]
	if !slices.Equal(tools.Sources, []string{"Makefile", "package.json"}) || tools.Jobs != 2 || tools.Checkout != nil {
		t.Fatalf("tools: %+v", tools)
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
