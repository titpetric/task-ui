package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The header's prompt asks in the page's folder and directory, about a
// page's run, in a workspace's settings in its folder, and at the root
// anywhere else.
func TestPage_HeaderPrompt(t *testing.T) {
	t.Parallel()
	malta := &folder.Folder{ID: "malta"}
	workspaces := map[string]int{"malta": 0}
	root := view.HeaderPrompt{Folder: "_", Label: "_", Href: "/f/_"}
	for _, c := range []struct {
		name string
		page view.Page
		want view.HeaderPrompt
	}{
		{"nowhere", view.Page{}, root},
		{"folder", view.Page{Folders: []*folder.Folder{malta}, Workspaces: workspaces, Folder: malta}, view.HeaderPrompt{Folder: "malta", Label: "malta", Href: "/f/malta"}},
		{"directory", view.Page{Folders: []*folder.Folder{malta}, Workspaces: workspaces, Folder: malta, Browse: "dockerfiles/"}, view.HeaderPrompt{Folder: "malta", Dir: "dockerfiles", Label: "malta/dockerfiles", Href: "/f/malta/dockerfiles"}},
		{"job of a directory", view.Page{Folders: []*folder.Folder{malta}, Workspaces: workspaces, Folder: malta, Job: &folder.Job{Dir: "dockerfiles"}}, view.HeaderPrompt{Folder: "malta", Dir: "dockerfiles", Label: "malta/dockerfiles", Href: "/f/malta/dockerfiles"}},
		{"workspace settings", view.Page{Folders: []*folder.Folder{malta}, Workspaces: workspaces, Workspace: &model.Workspace{Folder: "malta"}}, view.HeaderPrompt{Folder: "malta", Label: "malta", Href: "/f/malta"}},
		{"default workspace", view.Page{Workspace: &model.Workspace{Folder: "_"}}, root},
		{"internal", view.Page{Folder: &folder.Folder{ID: model.Internal}}, root},
		{"run", view.Page{Folders: []*folder.Folder{malta}, Workspaces: workspaces, Run: &view.Run{Run: model.Run{ID: "06gp9t19", Folder: "malta", Dir: "dockerfiles", Job: "build"}}}, view.HeaderPrompt{Folder: "malta", Dir: "dockerfiles", Run: "06gp9t19", Label: "run 06gp…", Href: "/runs/06gp9t19"}},
		{"prompt's run", view.Page{Folders: []*folder.Folder{malta}, Workspaces: workspaces, Run: &view.Run{Run: model.Run{ID: "06gp9t19", Folder: "malta", Job: model.PromptJob}}}, view.HeaderPrompt{Folder: "malta", Label: "malta", Href: "/f/malta"}},
		{"internal run", view.Page{Run: &view.Run{Run: model.Run{ID: "06gp9t19", Folder: model.Internal, Job: "claude"}}}, root},
	} {
		assert.Equal(t, c.want, c.page.HeaderPrompt(), c.name)
	}
}
