package api_test

import (
	"net/http"
	"sort"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// catalogue is a fixed set of folders.
type catalogue map[string]*folder.Folder

func (c catalogue) Find(id string) *folder.Folder { return c[id] }

func (c catalogue) List() []*folder.Folder {
	ids := make([]string, 0, len(c))
	for id := range c {
		// The root holds @host alone here, and lists as no folder.
		if id != model.Default {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	result := make([]*folder.Folder, 0, len(ids))
	for _, id := range ids {
		result = append(result, c[id])
	}
	return result
}

// response is a decoded reply.
type response struct {
	Status  int
	Headers http.Header
	Body    map[string]any
	Raw     []byte
}

// get returns a nested value of the body by path, or nil.
func (r response) get(path ...string) any {
	var v any = r.Body
	for _, key := range path {
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		v = m[key]
	}
	return v
}

// str returns a string value of the body, or "".
func (r response) str(path ...string) string {
	s, _ := r.get(path...).(string)
	return s
}

// list returns a list value of the body.
func (r response) list(key string) []any {
	l, _ := r.Body[key].([]any)
	return l
}
