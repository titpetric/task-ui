package view

import (
	"strings"

	"github.com/titpetric/task-ui/model"
)

// FolderHref returns the path of a folder's page under /f: a folder of a
// named workspace, or of internal, by its ID; the default workspace at
// /f/_; any other folder, a project at the root, under /f/_/.
func FolderHref(id string, workspaces map[string]int) string {
	if id == model.Default {
		return "/f/" + model.Default
	}
	top, _, _ := strings.Cut(id, "/")
	if _, named := workspaces[top]; named || top == model.Internal {
		return "/f/" + id
	}
	return "/f/" + model.Default + "/" + id
}
