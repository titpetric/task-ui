package frontend_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/integration"
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
	// Each carries a shebang for a marker.
	assert.Contains(t, ops, `<i class="shebang" aria-hidden="true">#!</i>`)
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

	// The sidebar leaves out a workflow set not to show there, and the
	// Workflows page still lists it.
	user := f.users[asOps]
	part := &model.Workflow{Name: "part", UserID: user.ID, UserName: user.Name, SidebarHidden: true}
	assert.NoError(t, f.opts.Workflows.Create(context.Background(), part, []model.WorkflowStep{{Folder: "infra", Job: "plan"}}))
	ops = f.page(t, asOps, "/")
	assert.NotContains(t, ops, `href="/workflows/`+part.ID+`"`)
	assert.Contains(t, f.page(t, asOps, "/workflows"), `href="/workflows/`+part.ID+`"`)

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
	// The list links to the form for a new one, which lists every job
	// ops can see; the list itself has no form.
	assert.Contains(t, ops, `href="/workflows/new"`)
	assert.NotContains(t, ops, `data-api="/api/workflows"`)
	form := f.page(t, asOps, "/workflows/new")
	assert.Contains(t, form, `data-api="/api/workflows"`)
	assert.Contains(t, form, `<option value="plan" data-folder="infra">`)

	// Any user makes workflows, a viewer too, and runs none.
	viewer := f.page(t, asViewer, "/workflows/new")
	assert.Contains(t, viewer, `data-api="/api/workflows"`)
	assert.NotContains(t, f.page(t, asViewer, "/workflows"), `data-workflow-run="`)

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

	run, err := f.opts.Flows.Start(context.Background(), wf, f.users[asOps], "")
	assert.NoError(t, err)
	select {
	case <-f.opts.Flows.Done(run.ID):
	case <-time.After(20 * time.Second):
		t.Fatal("the workflow did not finish")
	}

	ops := f.page(t, asOps, "/workflows/"+wf.ID)
	assert.Contains(t, ops, "<h2>release</h2>")
	assert.Contains(t, ops, `data-workflow-run="`)
	assert.Contains(t, ops, `href="/workflows/`+wf.ID+`/edit"`) // the owner edits it, on its own page
	assert.NotContains(t, ops, `data-method="PUT"`)
	assert.Contains(t, ops, `<i class="dot ok"></i>ok`)
	steps, err := f.opts.WorkflowRuns.Steps(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.Contains(t, ops, `href="/runs/`+steps[0].RunID+`"`)
	// The pipeline runs top to bottom as the steps, the finished run left
	// out of it, and the page follows the workflow's topic and never
	// reloads.
	assert.Contains(t, ops, `<ol class="pipeline" data-pipeline>`)
	assert.Contains(t, ops, `<li class="node" data-step="0">`)
	assert.NotContains(t, ops, `<a class="frame"`)
	assert.Contains(t, ops, `data-sse="workflow:`+wf.ID+`"`)
	assert.Contains(t, ops, `data-latest>Latest run`)
	assert.Contains(t, ops, `<tbody data-runs>`)
	assert.Contains(t, ops, `<tr data-run-id="`+run.ID+`">`)
	assert.NotContains(t, ops, `data-refresh`)
	assert.NotContains(t, ops, `data-live`)
	// Each run has View, which opens its pipeline: each step's terminal's
	// last screen, from the session still in memory; a finished one is
	// not polled.
	assert.Contains(t, ops, `<a class="btn btn-sm" href="/workflows/`+wf.ID+`/runs/`+run.ID+`" data-workflow-view>View</a>`)
	viewed := f.page(t, asOps, "/workflows/"+wf.ID+"/runs/"+run.ID)
	assert.Contains(t, viewed, `<section class="block" data-pipeline-view>`)
	assert.Contains(t, viewed, `<li class="node ok" data-step="0">`)
	assert.Contains(t, viewed, `<a class="frame" href="/runs/`+steps[0].RunID+`" data-attach="`+steps[0].RunID+`" role="img"`)
	assert.Contains(t, viewed, `aria-label="Terminal of infra / plan"><svg`)
	assert.Contains(t, viewed, `<div class="step-head">`)
	assert.NotContains(t, viewed, `data-pipeline>`)
	// plan prints a line: the screen shows it all, so there is no Expand.
	assert.NotContains(t, viewed, `data-expand="`+steps[0].RunID+`"`)
	assert.Equal(t, http.StatusNotFound, f.get(t, asOps, "/workflows/"+wf.ID+"/runs/"+model.NewID()).Status)

	// With a Claude token for the user, a finished run has Summary too.
	assert.NotContains(t, ops, `data-summary=`)
	f.withKey(t)
	assert.NoError(t, f.opts.Integrations.Save(context.Background(), integration.Claude, map[string]string{"token": "sk-1"}))
	summed := f.page(t, asOps, "/workflows/"+wf.ID)
	assert.Contains(t, summed, `<tbody data-runs data-summarize>`)
	assert.Contains(t, summed, `data-summary="/api/workflow-runs/`+run.ID+`/summary" data-summary-page="/workflows/`+wf.ID+`/runs/`+run.ID+`" data-summary-title="release"`)

	// A run that prints more than a screen gets Expand, with its lines.
	long := f.workflow(t, asOps, "long", "infra", "many")
	longRun, err := f.opts.Flows.Start(context.Background(), long, f.users[asOps], "")
	assert.NoError(t, err)
	select {
	case <-f.opts.Flows.Done(longRun.ID):
	case <-time.After(20 * time.Second):
		t.Fatal("the workflow did not finish")
	}
	longSteps, _ := f.opts.WorkflowRuns.Steps(context.Background(), longRun.ID)
	longPage := f.page(t, asOps, "/workflows/"+long.ID+"/runs/"+longRun.ID)
	assert.Contains(t, longPage, `data-expand="`+longSteps[0].RunID+`" aria-expanded="false"><svg class="ic"><use href="#i-expand"/></svg><span>Expand `)
	assert.Contains(t, longPage, `<span>Expand 201 lines</span></button>`) // the ssh-agent line, seq 1 200
	assert.NotContains(t, viewed, `data-frame=`)

	mine := f.workflow(t, asOps, "plan only", "infra", "plan")
	limited := f.page(t, asLimited, "/workflows/"+mine.ID)
	assert.NotContains(t, limited, `/edit"`) // not theirs
	hidden := f.get(t, asLimited, "/workflows/"+wf.ID)
	assert.Equal(t, 404, hidden.Status) // a step in infra/edge

	missing := f.get(t, asOps, "/workflows/nope")
	assert.Equal(t, 404, missing.Status)
}

// The sidebar's recents hold the runs of workflows among those of jobs,
// newest first, linking to the workflow.
func TestWorkflow_Recents(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	wf := f.workflow(t, asOps, "release", "infra", "plan")
	private := f.workflow(t, asAdmin, "private", "private", "deploy")
	for _, w := range []*model.Workflow{wf, private} {
		run, err := f.opts.Flows.Start(context.Background(), w, f.users[asAdmin], "")
		assert.NoError(t, err)
		select {
		case <-f.opts.Flows.Done(run.ID):
		case <-time.After(20 * time.Second):
			t.Fatal("the workflow did not finish")
		}
	}

	ops := f.page(t, asOps, "/")
	recents := ops[strings.Index(ops, `id="recents"`):]
	entry := `<a href="/workflows/` + wf.ID + `" class="recent flow" title="Workflow release, admin, ok">`
	assert.Contains(t, recents, entry)
	assert.Contains(t, recents, `<div class="recent-group">`)
	assert.Contains(t, recents, `<b>release</b><small><svg class="ic"><use href="#i-flow"/></svg>workflow</small>`)
	// The list shows its step's run too; and ops sees the private folder.
	assert.Contains(t, recents, `<b>plan</b>`)
	assert.Contains(t, recents, `href="/workflows/`+private.ID+`"`)
	// limited does not see the private folder, nor the workflow that ran in it.
	limited := f.page(t, asLimited, "/")
	assert.NotContains(t, limited[strings.Index(limited, `id="recents"`):], `href="/workflows/`+private.ID+`"`)
}

// The edit page carries the form, for the owner or an admin.
func TestWorkflow_EditPage(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	wf := f.workflow(t, asOps, "release", "infra", "plan")

	edit := f.page(t, asOps, "/workflows/"+wf.ID+"/edit")
	assert.Contains(t, edit, "<h2>Edit release</h2>")
	assert.Contains(t, edit, `data-api="/api/workflows/`+wf.ID+`" data-method="PUT"`)
	assert.Contains(t, edit, `value="release"`)
	assert.Contains(t, f.page(t, asAdmin, "/workflows/"+wf.ID+"/edit"), `data-method="PUT"`)

	assert.Equal(t, 403, f.get(t, asLimited, "/workflows/"+wf.ID+"/edit").Status)
	assert.Equal(t, 404, f.get(t, asOps, "/workflows/nope/edit").Status)
}

// A chain names a folder once for the steps that follow in it; a step
// that runs a workflow links to it. The step editor lists the
// workflows, except the one edited and those that run it.
func TestWorkflows_ChainAndNested(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	same := f.workflow(t, asOps, "same", "infra", "plan", "infra", "deploy")
	user := f.users[asOps]
	outer := &model.Workflow{Name: "outer", UserID: user.ID, UserName: user.Name}
	assert.NoError(t, f.opts.Workflows.Create(context.Background(), outer, []model.WorkflowStep{{Folder: "infra", Job: "plan"}, {WorkflowRef: same.ID}}))

	list := f.page(t, asOps, "/workflows")
	assert.Contains(t, list, "infra / plan")
	assert.Contains(t, list, `>deploy</a>`)
	assert.NotContains(t, list, "infra / deploy")
	assert.Contains(t, list, `href="/workflows/`+same.ID+`"><svg class="ic"><use href="#i-flow"/></svg>same</a>`)
	form := f.page(t, asOps, "/workflows/new")
	assert.Contains(t, form, `<option value="workflow:`+same.ID+`">same</option>`)
	assert.Contains(t, form, `<option value="workflow:`+outer.ID+`">outer</option>`)

	edit := f.page(t, asOps, "/workflows/"+same.ID+"/edit")
	assert.NotContains(t, edit, `<option value="workflow:`+same.ID+`"`)
	assert.NotContains(t, edit, `<option value="workflow:`+outer.ID+`"`) // outer runs same
	edit = f.page(t, asOps, "/workflows/"+outer.ID+"/edit")
	assert.Contains(t, edit, `<option value="workflow:`+same.ID+`"`)
}

// A step that ran a workflow shows that run's chain under it, each step
// with its run and terminal.
func TestWorkflow_NestedPipeline(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	inner := f.workflow(t, asOps, "inner", "infra", "plan")
	user := f.users[asOps]
	outer := &model.Workflow{Name: "outer", UserID: user.ID, UserName: user.Name}
	assert.NoError(t, f.opts.Workflows.Create(context.Background(), outer, []model.WorkflowStep{{WorkflowRef: inner.ID}}))

	run, err := f.opts.Flows.Start(context.Background(), outer, user, "")
	assert.NoError(t, err)
	select {
	case <-f.opts.Flows.Done(run.ID):
	case <-time.After(20 * time.Second):
		t.Fatal("the workflow did not finish")
	}
	steps, _ := f.opts.WorkflowRuns.Steps(context.Background(), run.ID)
	innerSteps, _ := f.opts.WorkflowRuns.Steps(context.Background(), steps[0].WorkflowRunRef)

	page := f.page(t, asOps, "/workflows/"+outer.ID+"/runs/"+run.ID)
	assert.Contains(t, page, `<ol class="pipeline sub" data-sub="`+steps[0].WorkflowRunRef+`" data-workflow="`+inner.ID+`">`)
	// The step's name links to its run's page; only the screen and the
	// buttons open the window.
	assert.Contains(t, page, `href="/runs/`+innerSteps[0].RunID+`" data-run-link="`+innerSteps[0].RunID+`">infra / plan</a>`)
	assert.Contains(t, page, `<a class="frame" href="/runs/`+innerSteps[0].RunID+`"`)
}

// A workflow's runs come a page at a time, as on Runs; the latest run
// above the pipeline stays the newest on every page.
func TestWorkflow_Pages(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	wf := f.workflow(t, asOps, "release", "infra", "plan")
	var made []string // newest first
	for i := range 18 {
		run := &model.WorkflowRun{WorkflowID: wf.ID, Name: wf.Name, UserName: "ops"}
		assert.NoError(t, f.opts.WorkflowRuns.Create(context.Background(), run, []model.WorkflowStep{{Folder: "infra", Job: "plan"}}))
		run.State = model.StateOK
		started := time.Now().Add(time.Duration(i-60) * time.Second)
		run.StartedAt = &started
		assert.NoError(t, f.opts.WorkflowRuns.Update(context.Background(), run))
		made = append([]string{run.ID}, made...)
	}
	rows := func(body string) int { return strings.Count(body, `<tr data-run-id="`) }
	path := "/workflows/" + wf.ID

	first := f.page(t, asOps, path)
	assert.Equal(t, 15, rows(first))
	assert.Contains(t, first, `<option value="15" selected>15</option>`)
	assert.Contains(t, first, `href="`+path+`?before=`+made[14]+`&amp;n=15" rel="next">Next</a>`)
	assert.NotContains(t, first, `rel="prev"`)
	assert.Equal(t, 18, rows(f.page(t, asOps, path+"?n=50")))

	next := f.page(t, asOps, path+"?before="+made[14])
	assert.Equal(t, 3, rows(next))
	assert.Contains(t, next, `href="`+path+`?after=`+made[15]+`&amp;n=15" rel="prev">Previous</a>`)
	assert.Contains(t, next, `data-latest>Latest run`)
	back := f.page(t, asOps, path+"?after="+made[15])
	assert.Equal(t, 15, rows(back))
	assert.Contains(t, back, `<tr data-run-id="`+made[0]+`">`)
}

// The editor has a toggle on every step, a disabled one dimmed, and
// for a user with prompt/edit the Prompt, its Summarize preset and Run
// in; the workflow page shows the prompt under the steps, and its Run
// opens the input dialog.
func TestWorkflow_PromptEditor(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	admin := f.users[asAdmin]
	wf := &model.Workflow{Name: "summed", UserID: admin.ID, UserName: admin.Name, Prompt: "summarize, change nothing", PromptFolder: "infra"}
	assert.NoError(t, f.opts.Workflows.Create(ctx, wf, []model.WorkflowStep{{Folder: "infra", Job: "plan"}, {Folder: "infra", Job: "many", Disabled: true}}))

	edit := f.page(t, asAdmin, "/workflows/"+wf.ID+"/edit")
	assert.Equal(t, 2, strings.Count(edit, `<input type="checkbox" data-step-enabled aria-label="Run this step"`)-1, "two rows and the template")
	assert.Contains(t, edit, `<li class="step disabled" data-step>`)
	assert.Contains(t, edit, `data-step-enabled aria-label="Run this step" checked>`)
	assert.Contains(t, edit, `name="prompt" rows="4"`)
	assert.Contains(t, edit, `>summarize, change nothing</textarea>`)
	assert.Contains(t, edit, `data-workflow-preset="Summarise the run in one or two short paragraphs`)
	assert.Contains(t, edit, `<option value="infra" selected>infra</option>`)
	assert.Contains(t, edit, `<option value="">the root</option>`)

	// Without prompt/edit the form has no prompt fields.
	ops := f.page(t, asOps, "/workflows/new")
	assert.NotContains(t, ops, `name="prompt" rows="4"`)
	assert.Contains(t, ops, `data-step-enabled`)

	page := f.page(t, asAdmin, "/workflows/"+wf.ID)
	assert.Contains(t, page, `data-workflow-run="`+wf.ID+`" data-workflow-prompt="summed"`)
	assert.Contains(t, page, `<p class="workflow-prompt" data-workflow-prompt-shown>summarize, change nothing</p>`)
	assert.Contains(t, page, "Claude gets it in infra/ once the steps end")
	assert.Contains(t, page, `<li class="node disabled" data-step="1">`)
	assert.Contains(t, page, "step 2, disabled: a run skips it")
	assert.Contains(t, page, `id="workflow-input-dialog"`)

	plain := f.workflow(t, asOps, "plain", "infra", "plan")
	assert.Contains(t, f.page(t, asOps, "/workflows/"+plain.ID), `data-workflow-run="`+plain.ID+`">`)
}

// A workflow run's page shows its skipped steps and the thread of its
// prompt: the prompt with the input, and Claude's answer.
func TestWorkflow_RunPromptThread(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	admin := f.users[asAdmin]
	wf := &model.Workflow{Name: "summed", UserID: admin.ID, UserName: admin.Name, Prompt: "summarize, change nothing"}
	assert.NoError(t, f.opts.Workflows.Create(ctx, wf, []model.WorkflowStep{{Folder: "infra", Job: "plan", Disabled: true}}))
	run := &model.WorkflowRun{WorkflowID: wf.ID, Name: wf.Name, UserID: admin.ID, UserName: admin.Name}
	steps, _ := f.opts.Workflows.Steps(ctx, wf.ID)
	assert.NoError(t, f.opts.WorkflowRuns.Create(ctx, run, steps))
	rows, _ := f.opts.WorkflowRuns.Steps(ctx, run.ID)
	rows[0].State = model.StepSkipped
	assert.NoError(t, f.opts.WorkflowRuns.UpdateStep(ctx, &rows[0]))
	run.State = model.StateOK
	assert.NoError(t, f.opts.WorkflowRuns.Update(ctx, run))
	asked := &model.Note{UserID: admin.ID, UserName: admin.Name, Body: "summarize, change nothing\n\nInput:\nfor 1.2", State: model.NoteDone, WorkspaceID: model.Default, Dir: model.Default, AboutWorkflowRunID: run.ID}
	assert.NoError(t, f.opts.Board.Notes.Create(ctx, asked))
	assert.NoError(t, f.opts.Board.Notes.Create(ctx, &model.Note{ParentID: asked.ID, UserName: "claude", Body: "All **quiet**.", WorkspaceID: model.Default, Dir: model.Default, AboutWorkflowRunID: run.ID}))

	page := f.page(t, asAdmin, "/workflows/"+wf.ID+"/runs/"+run.ID)
	assert.Contains(t, page, "<small>skipped</small>")
	assert.Contains(t, page, "the workflow's prompt, asked in the root once the steps ended")
	assert.Contains(t, page, `data-notes data-dir="_"`)
	assert.Contains(t, page, "for 1.2")
	assert.Contains(t, page, "<strong>quiet</strong>")
	// The root's prompts leave it out.
	assert.NotContains(t, f.page(t, asAdmin, "/f/_"), "for 1.2")
}
