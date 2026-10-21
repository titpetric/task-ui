package api

import (
	"github.com/titpetric/task-ui/folder"
)

// SkillView is a skill of the bar, each job as the jobs listing shows it.
type SkillView struct {
	*folder.Skill
	Jobs []JobView `json:"jobs"`
}
