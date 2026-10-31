package view_test

import (
	"bytes"
	"regexp"
	"strings"
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
	p.StepFolders = append(p.StepFolders, p.Folders...)
	for _, id := range []string{"ops/deploy", "malta/vhosts"} {
		job := &folder.Job{Folder: id, ID: "up"}
		job.Name = "up"
		p.StepFolders = append(p.StepFolders, &folder.Folder{ID: id, Jobs: []*folder.Job{job}})
	}
	p.Workspaces = map[string]int{"malta": 0, "ops": 1, "empty": 2}
	p.WorkspaceList = []model.Workspace{{Folder: "malta", Title: "Malta"}}
	return p
}

// Internal comes first, then the named ones holding a folder the user
// sees, by rank, each by its title or its folder.
func TestStepSpace(t *testing.T) {
	t.Parallel()
	p := spacePage(t)
	assert.Equal(t, []view.StepSpace{{ID: "internal", Title: "Internal"}, {ID: "malta", Title: "Malta"}, {ID: "ops", Title: "ops"}}, p.StepSpaces())
	assert.Equal(t, "malta", p.StepSpaceOf("malta/vhosts"))
	assert.Equal(t, "internal", p.StepSpaceOf("infra"))
	assert.Equal(t, "internal", p.StepSpaceOf(""))
	assert.True(t, p.StepNone(), "a user who sees every folder may name none")

	p.WorkspaceList = append(p.WorkspaceList, model.Workspace{Folder: "internal", Title: "Root"})
	assert.Equal(t, "Root", p.StepSpaces()[0].Title)

	// Without the default workspace row and without a folder at the root,
	// internal is still a place, then the named workspaces.
	p.WorkspaceList = p.WorkspaceList[:1]
	p.StepFolders = p.StepFolders[len(p.Folders):]
	assert.Equal(t, []view.StepSpace{{ID: "internal", Title: "Internal"}, {ID: "malta", Title: "Malta"}, {ID: "ops", Title: "ops"}}, p.StepSpaces())
	p.WorkspaceList = append(p.WorkspaceList, model.Workspace{Folder: "internal", Title: "Root"})
	assert.Equal(t, "Root", p.StepSpaces()[0].Title, "a titled internal keeps its title")

	p.User = &model.User{Folders: `["infra"]`}
	assert.False(t, p.StepNone(), "a user who sees some folders may not name none")
}

// Each step card picks a workspace first, preselected from the step's
// folder; the folder options carry their workspace for the cascade.
func TestView_RenderStepSpace(t *testing.T) {
	t.Parallel()
	v := shared(t)
	p := spacePage(t)
	p.StepWorkflows = []view.Workflow{{Workflow: model.Workflow{ID: "w2", Name: "nightly"}}}
	p.Workflow = &view.Workflow{
		Workflow: model.Workflow{ID: "w1", Name: "release"},
		Steps: []view.Step{
			{Kind: model.StepJob, Folder: "malta/vhosts", Job: "up"},
			{Kind: model.StepWorkflow, WorkflowID: "w2"},
			{Kind: model.StepPrompt, Folder: "ops/deploy", Prompt: "Why?"},
			{Kind: model.StepPrompt, Folder: "", Prompt: "Where?"},
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
	assert.Contains(t, out, `<option value="malta/vhosts" data-workspace="malta" data-rel="vhosts" selected>vhosts</option>`)
	assert.Contains(t, out, `<option value="infra" data-workspace="internal" data-rel="infra">infra</option>`)
	assert.Contains(t, out, `<option value="workflow:w2" data-workspace="workflow:" selected>nightly</option>`)
	assert.Contains(t, out, `<option value="internal" data-workspace="internal" data-rel=".">the root</option>`)
	assert.Contains(t, out, `<option value="ops/deploy" data-workspace="ops" data-rel="deploy" selected>deploy</option>`)
	assert.Contains(t, out, `<option value="" data-workspace="" data-rel="." selected>none (the workspace root)</option>`, "a step of no workspace")
	assert.Contains(t, out, `data-step-folder-label>Workflow</span>`)
	assert.NotContains(t, out, `value="empty"`, "a workspace without a folder the user sees")
}

// movePage is spacePage with infra in malta too, its job plan alone.
func movePage(t *testing.T) *view.Page {
	t.Helper()
	p := spacePage(t)
	job := &folder.Job{Folder: "malta/infra", ID: "plan"}
	job.Name = "plan"
	p.StepFolders = append(p.StepFolders, &folder.Folder{ID: "malta/infra", Jobs: []*folder.Job{job}})
	p.Workflow = &view.Workflow{Workflow: model.Workflow{ID: "w1", Name: "release"}, CanChange: true}
	return p
}

// A folder shows relative to its workspace; the root of either is ".".
func TestStepRel(t *testing.T) {
	t.Parallel()
	p := spacePage(t)
	assert.Equal(t, "vhosts", p.StepRel("malta/vhosts"))
	assert.Equal(t, "apps/blog", p.StepRel("malta/apps/blog"))
	assert.Equal(t, "apps/blog", p.StepRel("apps/blog"))
	assert.Equal(t, ".", p.StepRel("malta"))
	assert.Equal(t, ".", p.StepRel("internal"))
	assert.Equal(t, "the root", p.StepLabel("malta"))
	assert.Equal(t, "vhosts", p.StepLabel("malta/vhosts"))

	assert.Equal(t, "malta/apps/blog", view.StepFolderIn("malta", "apps/blog"))
	assert.Equal(t, "apps/blog", view.StepFolderIn("internal", "apps/blog"))
	assert.Equal(t, "malta", view.StepFolderIn("malta", "."))
	assert.Equal(t, "internal", view.StepFolderIn("internal", "."))
}

// Moving the steps keeps each folder and job the workspace has, and
// clears and names what it lacks; a workflow step stays.
func TestMoveSteps(t *testing.T) {
	t.Parallel()
	p := movePage(t)
	steps := []view.Step{
		{Kind: model.StepJob, Folder: "infra", Job: "plan", Args: `["-v"]`},
		{Kind: model.StepJob, Folder: "ops/deploy", Job: "up"},
		{Kind: model.StepWorkflow, WorkflowID: "w2"},
		{Kind: model.StepPrompt, Folder: "infra", Dir: "docs", Prompt: "Why?"},
		{Kind: model.StepSummary, Folder: "ops/deploy"},
		{Kind: model.StepSummary, Folder: "internal"},
	}
	moved, missed := p.StepMove(steps, "malta")
	assert.Equal(t, 2, missed)
	assert.Equal(t, view.Step{Kind: model.StepJob, Folder: "malta/infra", Job: "plan", Args: `["-v"]`}, moved[0])
	assert.Equal(t, view.Step{Kind: model.StepJob, Missing: "deploy is not in Malta."}, moved[1])
	assert.Equal(t, steps[2], moved[2])
	assert.Equal(t, view.Step{Kind: model.StepPrompt, Folder: "malta/infra", Dir: "docs", Prompt: "Why?"}, moved[3])
	assert.Equal(t, view.Step{Kind: model.StepSummary, Folder: "malta", Missing: "deploy is not in Malta."}, moved[4])
	assert.Equal(t, view.Step{Kind: model.StepSummary, Folder: "malta"}, moved[5])

	back, missed := p.StepMove(moved, "internal")
	assert.Equal(t, 0, missed)
	assert.Equal(t, "infra", back[0].Folder)
	assert.Equal(t, "plan", back[0].Job)
	assert.Equal(t, "", back[4].Missing, "a move clears the last one's note")

	p.StepFolders[len(p.StepFolders)-1].Jobs[0].ID = "apply"
	moved, missed = p.StepMove(steps[:1], "malta")
	assert.Equal(t, 1, missed)
	assert.Equal(t, view.Step{Kind: model.StepJob, Folder: "malta/infra", Missing: "infra has no job plan in Malta."}, moved[0])

	kept, missed := p.StepMove([]view.Step{{Kind: model.StepPrompt, Prompt: "Why?"}}, "malta")
	assert.Equal(t, 0, missed)
	assert.Equal(t, "", kept[0].Folder, "a step of no workspace keeps none on a move")
}

// A move names the workspace by its title and writes its folder: the
// moved steps hold folder IDs, never the title; the title itself names
// no workspace to move to.
func TestMoveSteps_FolderNotTitle(t *testing.T) {
	t.Parallel()
	p := movePage(t)
	steps := []view.Step{
		{Kind: model.StepJob, Folder: "infra", Job: "plan"},
		{Kind: model.StepPrompt, Folder: "infra"},
		{Kind: model.StepSummary, Folder: "internal"},
	}
	moved, missed := p.StepMove(steps, "malta")
	assert.Equal(t, 0, missed)
	for _, s := range moved {
		assert.Equal(t, "malta", p.StepSpaceOf(s.Folder), s.Folder)
		assert.False(t, strings.HasPrefix(s.Folder, "Malta"), s.Folder)
	}
	assert.Equal(t, "malta/infra", moved[0].Folder)
	assert.Equal(t, "malta", moved[2].Folder)
	assert.Equal(t, "malta/infra", view.StepFolderIn("malta", "infra"))
	assert.False(t, p.StepSpaceHas("Malta"), "a title moves nothing")
	assert.True(t, p.StepSpaceHas("malta"))
}

// Every option of the edit page's workspace and folder selects carries
// an ID: a workspace's folder, a folder's ID or a workflow's; a title
// shows as the text alone.
func TestView_RenderStepValuesAreIDs(t *testing.T) {
	t.Parallel()
	v := shared(t)
	p := movePage(t)
	p.WorkspaceList = append(p.WorkspaceList, model.Workspace{Folder: "internal", Title: "Root"}, model.Workspace{Folder: "ops", Title: "Operations"})
	p.StepWorkflows = []view.Workflow{{Workflow: model.Workflow{ID: "w2", Name: "nightly"}}}
	p.Workflow.Steps, _ = p.StepMove([]view.Step{{Kind: model.StepJob, Folder: "infra", Job: "plan"}, {Kind: model.StepPrompt, Folder: "ops/deploy", Prompt: "Why?"}}, "malta")
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "workflow_edit", p))
	out := buf.String()
	titles := []string{}
	for _, s := range p.StepSpaces() {
		titles = append(titles, s.Title)
	}
	assert.Equal(t, []string{"Root", "Malta", "Operations"}, titles)
	values := regexp.MustCompile(`<option value="([^"]*)"`).FindAllStringSubmatch(out, -1)
	assert.True(t, len(values) > 6, "options rendered")
	for _, m := range values {
		for _, title := range titles {
			assert.False(t, m[1] == title || strings.HasPrefix(m[1], title+"/"), m[1])
		}
	}
	assert.Contains(t, out, `<option value="malta" selected>Malta</option>`)
	assert.Contains(t, out, `<option value="malta/infra" data-workspace="malta" data-rel="infra" selected>infra</option>`)
}

// The edit page's menu moves the steps to each workspace, and a moved
// step says under its card what the workspace lacks.
func TestView_RenderMoveSteps(t *testing.T) {
	t.Parallel()
	v := shared(t)
	p := movePage(t)
	assert.True(t, p.StepSpaceHas("malta"))
	assert.False(t, p.StepSpaceHas("empty"))
	p.Workflow.Steps, _ = p.StepMove([]view.Step{{Kind: model.StepJob, Folder: "ops/deploy", Job: "up"}}, "malta")
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "workflow_edit", p))
	out := buf.String()
	assert.Contains(t, out, `href="/workflows/w1/edit?workspace=malta"><svg class="ic"><use href="#i-folder"/></svg><span>Move to Malta</span>`)
	assert.Contains(t, out, `href="/workflows/w1/edit?workspace=internal">`)
	assert.Contains(t, out, `<p class="wf-hint" data-step-hint role="status">deploy is not in Malta.</p>`)

	p.StepFolders, p.Workspaces = p.StepFolders[:1], nil
	assert.Equal(t, 0, len(p.StepMoveItems()), "one workspace has nowhere to move to")
}
