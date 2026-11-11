package hostrun

import (
	"testing"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/folder"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// hostSession is the run of a host, as the starter's first one is.
func hostSession(id string) *runner.Session {
	run := &model.Run{ID: id, Folder: model.Default, Job: model.HostPrefix + "db1"}
	return runner.NewSession(run, &folder.Folder{ID: model.Default}, &folder.Job{ID: run.Job})
}

// The runs of a batch sit under its first run, and under whatever
// dispatched the batch where something did. Nothing precedes the first
// run, so it is the root of the batch.
func TestRequest(t *testing.T) {
	t.Parallel()
	req := Request{Targets: []string{"*"}, Command: "uptime", Batch: true}

	id, kind := req.parent(nil)
	assert.Equal(t, "", id)
	assert.Equal(t, "", kind)

	id, kind = req.parent(hostSession("r1"))
	assert.Equal(t, "r1", id)
	assert.Equal(t, "", kind)

	step := Request{Targets: []string{"*"}, Command: "uptime", Batch: true, ParentID: "wr1", ParentKind: storage.ParentWorkflowRun}
	id, kind = step.parent(hostSession("r1"))
	assert.Equal(t, "wr1", id)
	assert.Equal(t, storage.ParentWorkflowRun, kind)
}
