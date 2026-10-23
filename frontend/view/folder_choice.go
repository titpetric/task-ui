package view

import (
	"slices"

	"github.com/titpetric/task-ui/folder"
)

// FolderChoice is a box of the user form's folder list: a folder, and
// whether the user sees it. A folder the user sees that the catalogue lacks keeps a box, checked,
// so saving keeps it.
type FolderChoice struct {
	ID      string
	Checked bool
}

// FolderChoices lists a box for every folder, checked for those in seen, then the rest of seen.
func FolderChoices(folders []*folder.Folder, seen []string) []FolderChoice {
	choices := make([]FolderChoice, 0, len(folders))
	known := map[string]bool{}
	for _, f := range folders {
		known[f.ID] = true
		choices = append(choices, FolderChoice{ID: f.ID, Checked: slices.Contains(seen, f.ID)})
	}
	for _, id := range seen {
		if !known[id] {
			choices = append(choices, FolderChoice{ID: id, Checked: true})
		}
	}
	return choices
}
