package view

import (
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/skills"
)

// PipelineError is one row of a page's pipeline alert: a finding of
// atkins' linter and the folder that holds its pipeline.
type PipelineError struct {
	Folder  *folder.Folder
	Finding skills.Finding
}
