package runner

import (
	"time"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/model"
)

// Request is what starts a run.
type Request struct {
	Folder *folder.Folder
	Job    *folder.Job
	User   *model.User
	Force  bool
	// Reason says why the request forces a run out of hours; the run keeps it.
	Reason    string
	Scheduled bool
	Cols      int
	Rows      int
	// Env adds NAME=value pairs to the run's environment, after the
	// secrets and before the TASK_UI_ variables, which win.
	Env []string
	// Workspace is the directory of the workspace the run belongs to, the
	// root for the default one. Empty takes the workspace of the folder's
	// first segment, or the default workspace for a folder in none.
	Workspace string
	// WorkspaceOf is a folder ID whose workspace the run takes when
	// Workspace is empty: for an internal job run for a folder, such as a
	// prompt in it. Empty takes the workspace of the run's own folder.
	WorkspaceOf string
	// Dir is the directory the job runs in, absolute: a directory inside
	// the folder, which the caller holds to the folder's rules, for a job
	// of Folder.Sub. Empty runs in the folder's own directory.
	Dir string
	// Args follow the job's argv after --, as argv and never through a
	// shell: only for a job that takes args.
	Args []string
	// Target is the host a run of @host runs on, which the row keeps;
	// two runs of a job with different targets run at once.
	Target string
	// BatchID groups the run with others started together, the ID of
	// the batch's first run; the row keeps it.
	BatchID string
	// Timeout ends the run, killed, once it runs that long. Zero lets
	// it run.
	Timeout time.Duration
	// ID runs the request again under the ID of a run of the same job,
	// in the same directory: the new attempt replaces the run's row,
	// recording, summary and artifacts. Empty takes a new ID.
	ID string
}

// newRun is the row of the run a request starts, under its ID or, when
// it has none, a new one the storage sets.
func (req Request) newRun() *model.Run {
	run := &model.Run{
		ID:       req.ID,
		Folder:   req.Folder.ID,
		Job:      req.Job.ID,
		Dir:      req.Job.Dir,
		Args:     model.EncodeList(req.Args),
		Target:   req.Target,
		BatchID:  req.BatchID,
		TermCols: int64(req.Cols),
		TermRows: int64(req.Rows),
	}
	if req.User != nil {
		run.UserID, run.UserName = req.User.ID, req.User.Name
	}
	if req.Force {
		run.Forced = 1
	}
	if req.Scheduled {
		run.Scheduled = 1
	}
	if run.TermCols < 20 || run.TermCols > 500 {
		run.TermCols = 120
	}
	if run.TermRows < 5 || run.TermRows > 200 {
		run.TermRows = 30
	}
	return run
}
