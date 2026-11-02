package app

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/config"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/phpfront"
	"github.com/titpetric/task-ui/runner"
)

// FrontBase is where the front end being ported answers while the Go pages
// still hold the root. It is the one place the prefix is spelled: the cutover
// sets it to "" and deletes the Go pages.
const FrontBase = "/development"

// FrontHost answers phpfront.Host off the instance: the catalogue, the live
// runner, the checkouts, the permission file and the event bus. Every row the
// front end reads it reads itself.
//
// Rules hands over the titles in file order and the path lists the file
// decides, so the decision in code/Task/Access.php is the only part of
// permission/ that is ported.
type FrontHost struct {
	Catalogue folder.Catalogue
	Runner    *runner.Manager
	Git       *git.Git
	Events    *event.Bus

	// Root is the workspace root, which a checkout's status is relative to.
	Root string
}

// Whoami is the name the Guard resolved for this request.
func (h *FrontHost) Whoami(ctx context.Context) string {
	user := access.UserFrom(ctx)
	if user == nil {
		return ""
	}
	return user.Name
}

// Folders is the catalogue: each folder and the jobs it declares.
func (h *FrontHost) Folders(context.Context) []phpfront.Folder {
	if h.Catalogue == nil {
		return nil
	}
	found := h.Catalogue.List()
	out := make([]phpfront.Folder, 0, len(found))
	for _, f := range found {
		jobs := make([]string, 0, len(f.Jobs))
		for _, job := range f.Jobs {
			jobs = append(jobs, job.Name)
		}
		out = append(out, phpfront.Folder{ID: f.ID, Title: f.Title, Dir: f.Dir, Jobs: jobs})
	}
	return out
}

// Running is the runs the runner holds, by the key their job is listed under.
func (h *FrontHost) Running(context.Context) []phpfront.Running {
	if h.Runner == nil {
		return nil
	}
	var out []phpfront.Running
	for _, session := range h.Runner.List() {
		run := session.Snapshot()
		if run.Finished() {
			continue
		}
		out = append(out, phpfront.Running{Key: run.Key(), ID: run.ID, State: string(run.State)})
	}
	return out
}

// Dirty counts the changed and untracked files under a checkout directory.
func (h *FrontHost) Dirty(_ context.Context, dir string) int {
	if h.Git == nil || h.Root == "" || dir == "" {
		return 0
	}
	status := h.Git.Status(dir, h.Root)
	if status == nil {
		return 0
	}
	rel, err := filepath.Rel(filepath.Join(h.Root, filepath.FromSlash(status.Checkout)), dir)
	if err != nil {
		return 0
	}
	return status.Under(strings.TrimPrefix(filepath.ToSlash(rel), "."))
}

// Rules is config/permissions.yml, titles and path lists both.
func (h *FrontHost) Rules(context.Context) phpfront.Rules {
	file := config.Default()
	return phpfront.Rules{
		Rules:       frontRules(file.Rules),
		Sections:    frontRules(file.Sections),
		Permissions: frontRules(file.Permissions),
		Paths:       file.Paths(),
		Checks:      file.Checks(),
		Extras:      frontExtras(file),
	}
}

// Publish sends an event on the stream every live label reads.
func (h *FrontHost) Publish(_ context.Context, topic string, payload map[string]any) error {
	if h.Events == nil {
		return nil
	}
	name, _ := payload["event"].(string)
	h.Events.Publish(topic, name, payload)
	return nil
}

// frontRules is a titled list of the permission file.
func frontRules(entries []config.Entry) []phpfront.Rule {
	out := make([]phpfront.Rule, 0, len(entries))
	for _, entry := range entries {
		out = append(out, phpfront.Rule{Key: entry.Key, Title: entry.Title})
	}
	return out
}

// frontExtras is the paths of the rules a section added of its own, which a
// section's all does not stand for.
func frontExtras(file *config.Permissions) []string {
	out := make([]string, 0, len(file.Extra))
	for _, extra := range file.Extra {
		out = append(out, extra.Path())
	}
	return out
}
