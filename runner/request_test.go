package runner

import (
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// request is a request for the job of a folder, as a dispatcher builds one.
func request() Request {
	job := &folder.Job{Folder: "infra", ID: "deploy"}
	job.Name = "deploy"
	return Request{Folder: &folder.Folder{ID: "infra"}, Job: job, User: &model.User{ID: "u1", Name: "ops"}}
}

// The row of a run carries the parent of the request that started it, and
// a run a person started has none.
func TestRequest(t *testing.T) {
	req := request()
	req.ParentID, req.ParentKind = "wr1", storage.ParentWorkflowRun
	run := req.newRun()
	if run.ParentID != "wr1" || run.ParentKind != storage.ParentWorkflowRun {
		t.Fatalf("the workflow run is the parent: %q %q", run.ParentID, run.ParentKind)
	}

	own := request().newRun()
	if own.ParentID != "" || own.ParentKind != "" {
		t.Fatalf("a run of a person has no parent: %q %q", own.ParentID, own.ParentKind)
	}
}

// A rerun under the ID of the run its thread holds is not its own parent.
func TestRequest_SelfParent(t *testing.T) {
	req := request()
	req.ID, req.ParentID = "r1", "r1"
	run := req.newRun()
	if run.ID != "r1" || run.ParentID != "" || run.ParentKind != "" {
		t.Fatalf("a run is no parent of itself: %q %q", run.ParentID, run.ParentKind)
	}
}
