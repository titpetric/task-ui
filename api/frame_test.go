package api_test

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A run whose session is gone still has a last screen: the one its
// recording ends on.
func TestFrame_FromRecording(t *testing.T) {
	t.Parallel()
	f := newFixture(t)

	id := model.NewID()
	rel := filepath.Join("infra", "plan", id+".cast")
	path := filepath.Join(f.opts.Recordings.Root, rel)
	assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	assert.NoError(t, os.WriteFile(path, []byte("{\"version\":2,\"width\":40,\"height\":5}\n[0.1,\"o\",\"planned, from the recording\"]\n"), 0o644))
	run := &model.Run{ID: id, Folder: "infra", Job: "plan", State: model.StateOK, UserName: "ops", CastPath: rel, TermCols: 40, TermRows: 5}
	assert.NoError(t, f.opts.Runs.Create(context.Background(), run))

	res := f.call(t, asOps, http.MethodGet, "/runs/"+id+"/frame.txt", nil)
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
	assert.Contains(t, string(res.Raw), "planned, from the recording")
	svg := f.call(t, asOps, http.MethodGet, "/runs/"+id+"/frame.svg", nil)
	assert.Equal(t, http.StatusOK, svg.Status)
	assert.Contains(t, svg.Headers.Get("Content-Type"), "image/svg+xml")

	// ?all=1 is the whole output, where the last screen of a 5 row
	// terminal holds only the last lines.
	long := "{\"version\":2,\"width\":40,\"height\":5}\n"
	for i := 0; i < 20; i++ {
		long += "[0.1,\"o\",\"output line " + strconv.Itoa(i) + "\\r\\n\"]\n"
	}
	assert.NoError(t, os.WriteFile(path, []byte(long), 0o644))
	screen := f.call(t, asOps, http.MethodGet, "/runs/"+id+"/frame.txt", nil)
	assert.NotContains(t, string(screen.Raw), "output line 0\n")
	all := f.call(t, asOps, http.MethodGet, "/runs/"+id+"/frame.txt?all=1", nil)
	assert.Equal(t, http.StatusOK, all.Status)
	assert.Contains(t, string(all.Raw), "output line 0\n")
	assert.Contains(t, string(all.Raw), "output line 19")

	// No session and no recording: nothing to show.
	bare := &model.Run{Folder: "infra", Job: "plan", State: model.StateOK, UserName: "ops"}
	assert.NoError(t, f.opts.Runs.Create(context.Background(), bare))
	assert.Equal(t, http.StatusNotFound, f.call(t, asOps, http.MethodGet, "/runs/"+bare.ID+"/frame.txt", nil).Status)
}
