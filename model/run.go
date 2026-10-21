package model

import (
	"time"
)

// Key is the identity the run lock uses: one run per job per folder
// and directory inside it.
func (r *Run) Key() string {
	return RunKey(r.Folder, r.Dir, r.Job)
}

// RunKey is the key of a job: folder/job, or folder/dir/job for a job
// of a directory inside the folder.
func RunKey(folder, dir, job string) string {
	if dir != "" {
		return folder + "/" + dir + "/" + job
	}
	return folder + "/" + job
}

// Where is what a title names the run by before its job: the folder,
// or folder/dir for a job of a directory inside it.
func (r Run) Where() string {
	if r.Dir != "" {
		return r.Folder + "/" + r.Dir
	}
	return r.Folder
}

// ArgList returns the arguments the run passed the job, after --.
func (r *Run) ArgList() []string {
	return DecodeList(r.Args)
}

// Finished reports whether the run ended.
func (r *Run) Finished() bool {
	return r != nil && Finished(r.State)
}

// Elapsed is how long the run took, or has been going.
func (r *Run) Elapsed() time.Duration {
	if r == nil || r.StartedAt == nil {
		return 0
	}
	if r.FinishedAt != nil {
		return r.FinishedAt.Sub(*r.StartedAt)
	}
	return time.Since(*r.StartedAt)
}
