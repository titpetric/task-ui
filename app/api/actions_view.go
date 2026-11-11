package api

import (
	"github.com/titpetric/task-ui/git"
)

// ActionsView is the actions bar of a folder, or of a directory inside
// it: the git line of its checkout, the skills atkins applies
// there with their jobs, and the place's own jobs.
type ActionsView struct {
	Folder string      `json:"folder"`
	Dir    string      `json:"dir"`
	Git    *git.Status `json:"git"`
	Skills []SkillView `json:"skills"`
	Jobs   []JobView   `json:"jobs"`
}
