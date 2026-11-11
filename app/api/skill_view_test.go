package api_test

import (
	"encoding/json"
	"testing"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
)

// A skill of the bar encodes as the skill, its jobs with can_run.
func TestSkillView(t *testing.T) {
	t.Parallel()
	job := &folder.Job{ID: "@git:status", Folder: "infra"}
	sv := api.SkillView{Skill: &folder.Skill{ID: "git", Dir: "", Jobs: []*folder.Job{job}}, Jobs: []api.JobView{{Job: job, CanRun: true}}}
	data, err := json.Marshal(sv)
	assert.NoError(t, err)
	var back map[string]any
	assert.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, "git", back["id"])
	jobs := back["jobs"].([]any)
	assert.Len(t, jobs, 1)
	assert.Equal(t, true, jobs[0].(map[string]any)["can_run"])
}
