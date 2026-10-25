package view

import (
	"sort"
	"strings"
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// SideLimits is how many projects the sidebar lists for each workspace,
// by the workspace's folder, _ for the root: the most recently committed
// first, a project with no commit last, ties by name. A workspace not
// in it, or with 0, lists every project. Apply cuts the sidebar's
// folders to that: the projects past a limit go with the folders inside
// them, the workspaces' own rows stay, and the folders keep their order,
// which the sidebar sets itself.
type SideLimits map[string]int

// Apply returns folders less the projects past each workspace's limit,
// by when, the time of a project's last commit.
func (l SideLimits) Apply(folders []*folder.Folder, workspaces map[string]int, when func(id string) time.Time) []*folder.Folder {
	dropped := map[string]bool{}
	for workspace, limit := range l {
		if limit <= 0 {
			continue
		}
		var projects []string
		seen := map[string]bool{}
		for _, f := range folders {
			ws, project := projectOf(f.ID, workspaces)
			if ws == workspace && project != "" && !seen[project] {
				seen[project] = true
				projects = append(projects, project)
			}
		}
		at := map[string]time.Time{}
		for _, project := range projects {
			at[project] = when(project)
		}
		sort.SliceStable(projects, func(i, j int) bool {
			a, b := at[projects[i]], at[projects[j]]
			if a.Equal(b) {
				return strings.ToLower(projects[i]) < strings.ToLower(projects[j])
			}
			if a.IsZero() || b.IsZero() {
				return b.IsZero()
			}
			return a.After(b)
		})
		for _, project := range projects[min(limit, len(projects)):] {
			dropped[project] = true
		}
	}
	if len(dropped) == 0 {
		return folders
	}
	kept := make([]*folder.Folder, 0, len(folders))
	for _, f := range folders {
		if _, project := projectOf(f.ID, workspaces); !dropped[project] {
			kept = append(kept, f)
		}
	}
	return kept
}

// projectOf returns the workspace a folder is in, _ for the root, and
// the project, the folder at the top of the workspace's src that holds
// it; "" for the workspace's own folder.
func projectOf(id string, workspaces map[string]int) (string, string) {
	parts := strings.Split(id, "/")
	if _, named := workspaces[parts[0]]; named {
		if len(parts) == 1 {
			return parts[0], ""
		}
		return parts[0], parts[0] + "/" + parts[1]
	}
	if parts[0] == model.Default {
		return model.Default, ""
	}
	return model.Default, parts[0]
}
