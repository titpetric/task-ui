package frontend_test

import (
	"sort"

	"github.com/titpetric/task-ui/folder"
)

// catalogue is a fixed set of folders.
type catalogue map[string]*folder.Folder

func (c catalogue) Find(id string) *folder.Folder { return c[id] }

func (c catalogue) List() []*folder.Folder {
	ids := make([]string, 0, len(c))
	for id := range c {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	result := make([]*folder.Folder, 0, len(ids))
	for _, id := range ids {
		result = append(result, c[id])
	}
	return result
}
