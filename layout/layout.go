package layout

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/titpetric/task-ui/model"
)

// Home is the directory of a workspace's home under its directory, and
// the last segment of the home's folder ID.
const Home = "home"

// Src is the directory of a workspace's projects under its directory.
const Src = "src"

// Reserved reports a name that is task-ui's own: home, src, internal or _.
func Reserved(name string) bool {
	return name == Home || name == Src || model.IsReserved(name)
}

// WorkspaceDir returns the directory of a workspace: the root for _,
// else its folder below it.
func WorkspaceDir(root, workspace string) string {
	if workspace == "" || workspace == model.Default {
		return root
	}
	return filepath.Join(root, filepath.FromSlash(workspace))
}

// HomeDir returns the home of a workspace, the HOME of its runs.
func HomeDir(root, workspace string) string {
	return filepath.Join(WorkspaceDir(root, workspace), Home)
}

// SrcDir returns the directory that holds a workspace's projects.
func SrcDir(root, workspace string) string {
	return filepath.Join(WorkspaceDir(root, workspace), Src)
}

// ProjectDir returns the directory of a project, or a folder inside one, of a workspace.
func ProjectDir(root, workspace, project string) string {
	return filepath.Join(SrcDir(root, workspace), filepath.FromSlash(project))
}

// HomeID returns the folder ID of a workspace's home: home for _, else
// <workspace>/home.
func HomeID(workspace string) string {
	if workspace == "" || workspace == model.Default {
		return Home
	}
	return workspace + "/" + Home
}

// IsHome reports whether a folder ID is a workspace's home or a directory inside it.
func IsHome(named []string, id string) bool {
	_, rel := Split(named, id)
	return rel == Home || strings.HasPrefix(rel, Home+"/")
}

// Split returns the workspace of a folder ID and the ID's path relative
// to it: "" for the workspace itself. An ID whose first segment is no
// named workspace belongs to _, where the whole ID is the path.
func Split(named []string, id string) (string, string) {
	id = strings.Trim(id, "/")
	if id == "" || id == model.Default {
		return model.Default, ""
	}
	top, rest, _ := strings.Cut(id, "/")
	if slices.Contains(named, top) {
		return top, rest
	}
	return model.Default, id
}

// Join returns the folder ID of a path relative to a workspace: the
// workspace's own ID for "".
func Join(workspace, rel string) string {
	if workspace == "" || workspace == model.Default {
		if rel == "" {
			return model.Default
		}
		return rel
	}
	if rel == "" {
		return workspace
	}
	return workspace + "/" + rel
}

// Dir returns the directory of a folder ID among the named workspaces:
// the home for <workspace>/home, the src directory for a workspace's own
// ID, and the project's directory under src for the rest.
func Dir(root string, named []string, id string) string {
	workspace, rel := Split(named, id)
	switch {
	case rel == "":
		return SrcDir(root, workspace)
	case rel == Home || strings.HasPrefix(rel, Home+"/"):
		return filepath.Join(WorkspaceDir(root, workspace), filepath.FromSlash(rel))
	}
	return ProjectDir(root, workspace, rel)
}

// ID returns the folder ID of a directory under the root, Dir's inverse:
// false for a directory outside the root, the root itself, a
// workspace's directory, or one that is neither in a home nor in src.
func ID(root string, named []string, dir string) (string, bool) {
	rel, err := filepath.Rel(root, dir)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	workspace := model.Default
	if slices.Contains(named, parts[0]) {
		workspace, parts = parts[0], parts[1:]
	}
	if len(parts) == 0 {
		return "", false
	}
	switch parts[0] {
	case Home:
		return Join(workspace, strings.Join(parts, "/")), true
	case Src:
		if len(parts) == 1 {
			return Join(workspace, ""), true
		}
		return Join(workspace, strings.Join(parts[1:], "/")), true
	}
	return "", false
}

// Workspace returns the workspace a directory under the root belongs
// to, by its folder: a named one whose directory holds dir, else _.
// False for a directory outside the root.
func Workspace(root string, named []string, dir string) (string, bool) {
	rel, err := filepath.Rel(root, dir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	top, _, _ := strings.Cut(filepath.ToSlash(rel), "/")
	if slices.Contains(named, top) {
		return top, true
	}
	return model.Default, true
}

// Ensure makes the home and src directories of a workspace.
func Ensure(root, workspace string) error {
	for _, dir := range []string{HomeDir(root, workspace), SrcDir(root, workspace)} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// AtkinsDB returns the ATKINS_DB of a run with HOME at home: the home's
// history, then the instance's mirror under serverHome, when there is one.
func AtkinsDB(home, serverHome string) string {
	own := filepath.Join(home, ".atkins", "atkins.db")
	if serverHome == "" {
		return own
	}
	return own + string(os.PathListSeparator) + filepath.Join(serverHome, ".atkins", "atkins.db")
}
