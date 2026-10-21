package model

import (
	"strings"
)

// FolderRef is a row that names a folder by its ID: a schedule, a
// secret, an env file or a workflow step. A project that moves leaves
// these behind under its old ID.
type FolderRef struct {
	// Kind is schedule, secret, env_file or workflow_step.
	Kind string `db:"kind" json:"kind"`
	// ID is the row's: the schedule, the secret, the env file, or the
	// workflow the step is in.
	ID     string `db:"id" json:"id"`
	Folder string `db:"folder" json:"folder"`
	// Name is the schedule's job, the secret's name, the env file's path,
	// or the workflow's name and the step's job.
	Name string `db:"name" json:"name"`
}

// String names the row for a person: its kind, folder and name.
func (r FolderRef) String() string {
	return strings.ReplaceAll(r.Kind, "_", " ") + " " + r.Folder + " · " + r.Name
}

// Under reports whether a folder ID is id or a folder inside it.
func Under(folder, id string) bool {
	return folder == id || strings.HasPrefix(folder, id+"/")
}

// Moved returns a folder ID under from as it reads under to, and false
// for one that is not under from.
func Moved(folder, from, to string) (string, bool) {
	if !Under(folder, from) {
		return folder, false
	}
	return to + strings.TrimPrefix(folder, from), true
}
