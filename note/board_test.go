package note_test

import (
	"sort"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/note"
)

// Claude replies to a prompt under its own name.
func TestBoard_Answerer(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "claude", (&note.Board{}).Answerer())
}

// catalogue is a fixed set of folders.
type catalogue map[string]*folder.Folder

func (c catalogue) Find(id string) *folder.Folder { return c[id] }

func (c catalogue) List() []*folder.Folder {
	ids := make([]string, 0, len(c))
	for id := range c {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	list := make([]*folder.Folder, 0, len(ids))
	for _, id := range ids {
		list = append(list, c[id])
	}
	return list
}
