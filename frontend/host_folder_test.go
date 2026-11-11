package phpfront_test

import (
	"context"

	"github.com/titpetric/task-ui/phpfront"
)

// What the fixture answers of a folder: the catalogue, the checkouts and the
// artifact tree, grouped the way host_folder.go declares the shapes.

func (h *host) Folders(context.Context) []phpfront.Folder {
	return []phpfront.Folder{{
		ID:    "infra",
		Title: "Infra",
		Dir:   "/src/infra",
		Allow: []string{"ops"},
		Jobs:  []phpfront.Job{{ID: "build", Name: "build"}, {ID: "task:test", Name: "test", Allow: []string{"ann"}}},
	}}
}

// The job of a run is the folder, the directory inside it and the job, all
// three: the shell of a directory inside the folder is a different job from
// the shell of the folder, and only this answers either.
func (h *host) Job(_ context.Context, folder, dir, id string) phpfront.Job {
	if folder != "infra" || dir != "api" || id != "@shell" {
		return phpfront.Job{}
	}
	return phpfront.Job{
		Found:       true,
		ID:          "@shell",
		Name:        "shell",
		Dir:         "api",
		Shell:       true,
		Allow:       []string{"admin"},
		FolderAllow: []string{"ops"},
	}
}

func (h *host) Projects(_ context.Context, workspace string) []string {
	if workspace != "malta" {
		return nil
	}
	return []string{"malta/deploy"}
}

func (h *host) FolderDir(_ context.Context, id string, workspaces []string) string {
	if id == "infra" {
		return "/src/infra"
	}
	if len(workspaces) == 0 {
		return ""
	}
	return "/" + workspaces[0] + "/src/" + id
}

func (h *host) FolderDirs(ctx context.Context, ids []string, workspaces []string) map[string]string {
	out := make(map[string]string, len(ids))
	for _, id := range ids {
		if dir := h.FolderDir(ctx, id, workspaces); dir != "" {
			out[id] = dir
		}
	}
	return out
}

func (h *host) IsDir(_ context.Context, id, rel string) bool {
	return id == "infra" && rel == "api"
}

func (h *host) Commit(_ context.Context, dir string) string {
	if dir != "/src/infra" {
		return ""
	}
	return "2024-05-01 10:00:00 +0000 UTC"
}

func (h *host) Dirty(_ context.Context, dir string) int {
	if dir == "/src/infra" {
		return 3
	}
	return 0
}

func (h *host) Tree(_ context.Context, dir string) []phpfront.TreeItem {
	if dir != "lint" {
		return nil
	}
	return []phpfront.TreeItem{
		{Name: "nested", Dir: true, Size: 4096, ModTime: "2026-11-02 10:32:12.414 +0000 UTC"},
		{Name: "out.txt", Size: 31},
	}
}

func (h *host) Space(context.Context) phpfront.Space {
	return phpfront.Space{Measured: true, Files: 2048, History: 512, Ignored: 64}
}

// A directory nothing measured is not in the answer, which is what an
// instance with no meter answers for every one of them.
func (h *host) Spaces(_ context.Context, dirs []string) map[string]phpfront.Space {
	out := map[string]phpfront.Space{}
	for _, dir := range dirs {
		if dir != "/root/src/infra" {
			continue
		}
		out[dir] = phpfront.Space{Measured: true, Files: 1024, History: 256, Ignored: 32}
	}
	return out
}

// The folder pages read the catalogue and the disk, which this fake has none
// of: each answers the nothing an instance with no such folder answers.
func (h *host) Place(context.Context, string, string) phpfront.Place {
	return phpfront.Place{}
}

func (h *host) Groups(context.Context, string, string, []string, bool, bool) []phpfront.Group {
	return nil
}

func (h *host) Listing(context.Context, string, string, []string) []phpfront.Entry {
	return nil
}

func (h *host) RootTree(context.Context, bool, []string) []phpfront.Entry {
	return nil
}

func (h *host) ModTime(context.Context, string) string {
	return ""
}

func (h *host) Status(context.Context, string) phpfront.Status {
	return phpfront.Status{}
}

func (h *host) Checkout(context.Context, string) phpfront.Checkout {
	return phpfront.Checkout{}
}

func (h *host) Hours(context.Context, string) phpfront.Hours {
	return phpfront.Hours{}
}
