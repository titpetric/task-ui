package workflow_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/workflow"
)

// Expand walks the steps once per project: a step in a project of the
// workspace moves to each project at the same path, a step elsewhere and
// a workflow step stay as they are.
func TestExpand(t *testing.T) {
	t.Parallel()
	steps := []model.WorkflowStep{
		{Kind: model.StepJob, Folder: "malta/apps", Job: "@git:pull"},
		{Kind: model.StepJob, Folder: "malta/apps/cdn", Job: "@script", Args: `["go get -u ./..."]`},
		{Kind: model.StepPrompt, Folder: "malta/apps", Prompt: "Review the changes."},
		{Kind: model.StepJob, Folder: "infra", Job: "notify"},
		{Kind: model.StepWorkflow, WorkflowRef: "w1"},
	}
	expanded, groups := workflow.Expand(steps, "malta", []string{"apps", "deploy"})
	assert.Len(t, expanded, 10)
	assert.Equal(t, []int{0, 0, 0, 0, 0, 1, 1, 1, 1, 1}, groups)
	assert.Equal(t, "malta/apps", expanded[0].Folder)
	assert.Equal(t, "malta/deploy", expanded[5].Folder)
	assert.Equal(t, "malta/deploy/cdn", expanded[6].Folder, "the path inside the project stays")
	assert.Equal(t, `["go get -u ./..."]`, expanded[6].Args)
	assert.Equal(t, "malta/deploy", expanded[7].Folder, "a prompt asks in each project")
	assert.Equal(t, "infra", expanded[8].Folder, "a folder of another workspace stays")
	assert.Equal(t, "w1", expanded[9].WorkflowRef)
	assert.Equal(t, "malta/apps", steps[0].Folder, "the workflow's own steps stay as they are")

	// The default workspace names its projects without a prefix.
	root, _ := workflow.Expand([]model.WorkflowStep{{Kind: model.StepJob, Folder: "apps", Job: "build"}, {Kind: model.StepPrompt, Folder: model.Default}}, model.Default, []string{"apps", "web"})
	assert.Equal(t, "web", root[2].Folder)
	assert.Equal(t, model.Default, root[3].Folder, "the root is no project")

	none, groups := workflow.Expand(steps, "malta", nil)
	assert.Len(t, none, 0)
	assert.Len(t, groups, 0)
}

// projects lists the projects of the workspace ws.
type projects []*folder.Folder

func (p projects) WorkspaceProjects(_ context.Context, workspace string) []*folder.Folder {
	if workspace != "ws" {
		return nil
	}
	return p
}

// A workflow's projects come from the workspace it runs over.
func TestProjects(t *testing.T) {
	t.Parallel()
	var list workflow.Projects = projects{{ID: "ws/a"}}
	assert.Len(t, list.WorkspaceProjects(context.Background(), "ws"), 1)
	assert.Len(t, list.WorkspaceProjects(context.Background(), "other"), 0)
}
