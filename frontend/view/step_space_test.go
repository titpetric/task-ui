package view_test

import (
	"bytes"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// spacePage is an admin's page with a folder at the root, infra, and
// one in each of two workspaces, malta ranked before ops.
func spacePage(t *testing.T) *view.Page {
	t.Helper()
	p := admin(t)
	for _, id := range []string{"ops/deploy", "malta/vhosts"} {
		job := &folder.Job{Folder: id, ID: "up"}
		job.Name = "up"
		p.Folders = append(p.Folders, &folder.Folder{ID: id, Jobs: []*folder.Job{job}})
	}
	p.Workspaces = map[string]int{"malta": 0, "ops": 1, "empty": 2}
	p.WorkspaceList = []model.Workspace{{Folder: "malta", Title: "Malta"}}
	return p
}

// The default workspace comes first, then the named ones holding a
// folder the user sees, by rank, each by its title or its folder.
func TestStepSpace(t *testing.T) {
	t.Parallel()
	p := spacePage(t)
	assert.Equal(t, []view.StepSpace{{ID: "_", Title: "Default"}, {ID: "malta", Title: "Malta"}, {ID: "ops", Title: "ops"}}, p.StepSpaces())
	assert.Equal(t, "malta", p.StepSpaceOf("malta/vhosts"))
	assert.Equal(t, "_", p.StepSpaceOf("infra"))
	assert.Equal(t, "_", p.StepSpaceOf(""))

	p.WorkspaceList = append(p.WorkspaceList, model.Workspace{Folder: "_", Title: "Root"})
	assert.Equal(t, "Root", p.StepSpaces()[0].Title)
}

// Each step card picks a workspace first, preselected from the step's
// folder; the folder options carry their workspace for the cascade.
func TestView_RenderStepSpace(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := spacePage(t)
	p.StepWorkflows = []view.Workflow{{Workflow: model.Workflow{ID: "w2", Name: "nightly"}}}
	p.Workflow = &view.Workflow{
		Workflow: model.Workflow{ID: "w1", Name: "release"},
		Steps: []view.Step{
			{Kind: model.StepJob, Folder: "malta/vhosts", Job: "up"},
			{Kind: model.StepWorkflow, WorkflowID: "w2"},
			{Kind: model.StepPrompt, Folder: "ops/deploy", Prompt: "Why?"},
		},
		CanChange: true,
	}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "workflow_edit", p))
	out := buf.String()
	assert.Contains(t, out, `<select data-step-space aria-label="Workspace">`)
	assert.Contains(t, out, `<option value="malta" selected>Malta</option>`, "the command's workspace")
	assert.Contains(t, out, `<option value="workflow:" selected>Workflows</option>`, "the workflow step's")
	assert.Contains(t, out, `<option value="ops" selected>ops</option>`, "the prompt's")
	assert.Contains(t, out, `<option value="malta/vhosts" data-workspace="malta" selected>malta/vhosts</option>`)
	assert.Contains(t, out, `<option value="infra" data-workspace="_">infra</option>`)
	assert.Contains(t, out, `<option value="workflow:w2" data-workspace="workflow:" selected>nightly</option>`)
	assert.Contains(t, out, `<option value="_" data-workspace="_">the root</option>`)
	assert.Contains(t, out, `<option value="ops/deploy" data-workspace="ops" selected>ops/deploy</option>`)
	assert.Contains(t, out, `data-step-folder-label>Workflow</span>`)
	assert.NotContains(t, out, `value="empty"`, "a workspace without a folder the user sees")
}
