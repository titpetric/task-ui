package hostrun_test

import (
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// root is a catalogue of the root folder alone.
type root struct{ f *folder.Folder }

func (r root) Find(id string) *folder.Folder {
	if id == model.Default {
		return r.f
	}
	return nil
}
