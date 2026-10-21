package model

import (
	"path/filepath"
)

// Dir returns the directory of the workspace under the root: the root
// itself for the default workspace, else the folder below it.
func (w *Workspace) Dir(root string) string {
	if w.Folder == Default {
		return root
	}
	return filepath.Join(root, filepath.FromSlash(w.Folder))
}
