package api

import (
	"github.com/titpetric/task-ui/skills"
)

// FolderView is a folder as the jobs listing shows it: its title and
// description, and its jobs for the user asking.
type FolderView struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Jobs        []JobView `json:"jobs"`
	// Error says why reading the folder's files failed, and Errors are
	// the findings of atkins' linter in a pipeline it refuses, which
	// otherwise lists as a folder with no jobs.
	Error  string           `json:"error,omitempty"`
	Errors []skills.Finding `json:"errors,omitempty"`
}
