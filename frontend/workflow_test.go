package frontend_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func (f *fixture) workflow(t *testing.T, owner, name string, pairs ...string) *model.Workflow {
	t.Helper()
	steps := []model.WorkflowStep{}
	for i := 0; i+1 < len(pairs); i += 2 {
		steps = append(steps, model.WorkflowStep{Folder: pairs[i], Job: pairs[i+1]})
	}
	user := f.users[owner]
	wf := &model.Workflow{Name: name, UserID: user.ID, UserName: user.Name}
	assert.NoError(t, f.opts.Workflows.Create(context.Background(), wf, steps))
	return wf
}

// The sidebar has Workflows between the dashboard and the runs.
func TestWorkflows_Sidebar(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	body := f.page(t, asViewer, "/")
	dashboard := strings.Index(body, "<span>Dashboard</span>")
	workflows := strings.Index(body, `<a href="/workflows"`)
	runs := strings.Index(body, `<a href="/runs"`)
	assert.True(t, dashboard >= 0 && workflows > dashboard && runs > workflows, "Dashboard, Workflows, Runs")
}

// Under Workflows the sidebar lists the workflows the user sees, each
// with a run button when the user may run every step.
func TestWorkflows_SidebarList(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	release := f.workflow(t, asOps, "release", "infra", "plan", "infra/edge", "deploy")
	private := f.workflow(t, asAdmin, "private", "infra", "plan", "private", "deploy")

	ops := f.page(t, asOps, "/")
	workflows := strings.Index(ops, `<a href="/workflows"`)
	item := strings.Index(ops, `<a class="name" href="/workflows/`+release.ID+`"`)
	runs := strings.Index(ops, `<a href="/runs"`)
	assert.True(t, workflows >= 0 && item > workflows && runs > item, "release listed under Workflows")
	assert.Contains(t, ops, `data-workflow-run="`+release.ID+`"`)
	// ops sees private, and can not run its step in the admin's folder.
	assert.Contains(t, ops, `<a class="name" href="/workflows/`+private.ID+`"`)
	assert.NotContains(t, ops, `data-workflow-run="`+private.ID+`"`)

	// limited does not see infra/edge, so not release either.
	limited := f.page(t, asLimited, "/")
	assert.NotContains(t, limited, `href="/workflows/`+release.ID+`"`)

	// A viewer sees the workflow and no run button.
	viewer := f.page(t, asViewer, "/")
	assert.Contains(t, viewer, `<a class="name" href="/workflows/`+release.ID+`"`)
	assert.NotContains(t, viewer, `data-workflow-run="`)

	// The open workflow is the active entry.
	page := f.page(t, asOps, "/workflows/"+release.ID)
	assert.Contains(t, page, `<div class="flow active">`)
}

func TestWorkflows_List(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.workflow(t, asOps, "release", "infra", "plan", "infra/edge", "deploy")
	f.workflow(t, asAdmin, "private", "infra", "plan", "private", "deploy")
	f.workflow(t, asOps, "plan only", "infra", "plan")

	ops := f.page(t, asOps, "/workflows")
	assert.Contains(t, ops, "<h2>Workflows</h2>")
	assert.Contains(t, ops, ">release</a>")
	assert.Contains(t, ops, "infra / plan")
	assert.Contains(t, ops, "infra/edge / deploy")
	// ops may run release, not private, which needs the admin's folder.
	// The sidebar lists them too: count the page's own buttons.
	content := ops[strings.Index(ops, "<main>"):]
	assert.Contains(t, content, `data-workflow-run="`)
	assert.Equal(t, 2, strings.Count(content, `data-workflow-run="`))
	// The step editor offers every job ops can see.
	assert.Contains(t, ops, `data-api="/api/workflows"`)
	assert.Contains(t, ops, `<option value="plan" data-folder="infra">`)

	// Any user makes workflows, a viewer too, and runs none.
	viewer := f.page(t, asViewer, "/workflows")
	assert.Contains(t, viewer, `data-api="/api/workflows"`)
	assert.NotContains(t, viewer, `data-workflow-run="`)

	// A workflow with a step in a folder the user can not see is not
	// listed: limited sees infra, not infra/edge or private.
	limited := f.page(t, asLimited, "/workflows")
	assert.Contains(t, limited, ">plan only</a>")
	assert.NotContains(t, limited, ">release</a>")
	assert.NotContains(t, limited, ">private</a>")
}

func TestWorkflow_Page(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	wf := f.workflow(t, asOps, "release", "infra", "plan", "infra/edge", "deploy")

	run, err := f.opts.Flows.Start(context.Background(), wf, f.users[asOps])
	assert.NoError(t, err)
	select {
	case <-f.opts.Flows.Done(run.ID):
	case <-time.After(20 * time.Second):
		t.Fatal("the workflow did not finish")
	}

	ops := f.page(t, asOps, "/workflows/"+wf.ID)
	assert.Contains(t, ops, "<h2>release</h2>")
	assert.Contains(t, ops, `data-workflow-run="`)
	assert.Contains(t, ops, `data-method="PUT"`) // the owner edits it
	assert.Contains(t, ops, `<i class="dot ok"></i>ok`)
	steps, err := f.opts.WorkflowRuns.Steps(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.Contains(t, ops, `href="/runs/`+steps[0].RunID+`"`)

	mine := f.workflow(t, asOps, "plan only", "infra", "plan")
	limited := f.page(t, asLimited, "/workflows/"+mine.ID)
	assert.NotContains(t, limited, `data-method="PUT"`) // not theirs
	hidden := f.get(t, asLimited, "/workflows/"+wf.ID)
	assert.Equal(t, 404, hidden.Status) // a step in infra/edge

	missing := f.get(t, asOps, "/workflows/nope")
	assert.Equal(t, 404, missing.Status)
}
