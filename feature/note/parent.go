package note

import (
	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/storage"
)

// parentOf returns the run a thread's next run sits under, and its kind:
// the run working the thread, else the workflow run or the run the thread
// asks about, and none for a thread a person opened.
func parentOf(thread *model.Note) (string, string) {
	switch {
	case thread == nil:
		return "", ""
	case thread.RunID != "":
		return thread.RunID, storage.ParentRun
	case thread.AboutWorkflowRunID != "":
		return thread.AboutWorkflowRunID, storage.ParentWorkflowRun
	case thread.AboutRunID != "":
		return thread.AboutRunID, storage.ParentRun
	}
	return "", ""
}
