package view_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/permission"
)

// A workflow as a page shows it: the row, its steps drawn as a chain,
// and what the user asking may do with it.
func TestWorkflow(t *testing.T) {
	t.Parallel()
	w := view.Workflow{
		Workflow: model.Workflow{ID: "w1", Name: "release"},
		Steps:    []view.Step{{Folder: "dockerfiles", Job: "build"}, {Folder: "dockerfiles", Job: "push"}},
		CanRun:   true,
	}
	assert.Equal(t, "release", w.Name)
	assert.Equal(t, "push", view.Chain(w.Steps, nil)[1].Text)
	assert.True(t, w.CanRun)
	assert.False(t, w.CanChange)
	assert.False(t, w.Prompted())
	assert.Equal(t, "_", w.PromptPlace())

	w.Steps = append(w.Steps, view.Step{Kind: model.StepSummary, Folder: "infra", Dir: "app"})
	w.Steps[1].Disabled = true
	assert.True(t, w.Prompted())
	assert.Equal(t, "infra/app", w.PromptPlace())
	assert.True(t, view.Chain(w.Steps, nil)[1].Disabled)
	assert.True(t, view.Chain(w.Steps, nil)[2].Summary)
	assert.Equal(t, "summary in infra/app", view.Chain(w.Steps, nil)[2].Text)
	w.Steps[2].Disabled = true
	assert.False(t, w.Prompted(), "a step that is off asks nothing")
	box := view.PromptOfWorkflowRun("wr1", "infra/app", "infra/app/", true)
	assert.True(t, box.Watch)
	assert.True(t, box.Open, "open while the run goes")
	assert.Equal(t, "infra/app", box.Place)
	assert.Equal(t, "wr1", box.WorkflowRun)
	assert.False(t, view.PromptOfWorkflowRun("wr1", "infra/app", "infra/app/", false).Open, "folded once it ends")
}

// A run of a workflow carries the states of its steps.
func TestWorkflowRun(t *testing.T) {
	t.Parallel()
	r := view.WorkflowRun{
		WorkflowRun: model.WorkflowRun{ID: "wr1", State: model.StateFailed},
		Steps:       []view.Step{{Folder: "infra", Job: "plan", State: "ok"}, {Folder: "infra", Job: "apply", State: "failed"}},
	}
	assert.Equal(t, "wr1", r.ID)
	assert.Equal(t, "failed", view.Chain(r.Steps, nil)[1].State)
}

func TestWorkflowRun_Took(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	end := start.Add(90 * time.Second)
	run := view.WorkflowRun{WorkflowRun: model.WorkflowRun{StartedAt: &start, FinishedAt: &end}}
	assert.Equal(t, 90*time.Second, run.Took())
}

// The editor draws each step as a card of its kind on the rail, with
// its [...] menu, and Add step lists the kinds; the prompt and summary
// kinds take prompt/edit.
func TestView_RenderWorkflowEditor(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Workflow = &view.Workflow{
		Workflow: model.Workflow{ID: "w1", Name: "Porec watering", Description: "Water the plants or not."},
		Steps: []view.Step{
			{Kind: model.StepJob, Folder: "infra", Job: "plan"},
			{Kind: model.StepPrompt, Folder: "_", Prompt: "Fetch the weather for Porec.", Disabled: true},
			{Kind: model.StepSummary, Folder: "infra", Dir: "app"},
		},
		CanChange: true,
	}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "workflow_edit", p))
	out := buf.String()
	assert.Contains(t, out, `data-kind="job"`)
	assert.Contains(t, out, `class="wf-step kind-prompt off"`)
	assert.Contains(t, out, `data-kind="summary"`)
	assert.Contains(t, out, ">infra / plan</span>")
	assert.Contains(t, out, ">Fetch the weather for Porec.</span>")
	assert.Contains(t, out, ">of the steps before it, in infra/app</span>")
	assert.Contains(t, out, note.WorkflowSummaryQuestion)
	assert.Contains(t, out, `data-step-move="up"`)
	assert.Contains(t, out, `data-step-duplicate="1"`)
	assert.Contains(t, out, `<span>Enable</span>`, "the step that is off turns on")
	assert.Contains(t, out, `class="more-item danger" role="menuitem" data-step-remove="1"`)
	assert.Contains(t, out, `data-step-add="summary"`)
	assert.Contains(t, out, `data-step-template="prompt"`)
	assert.Contains(t, out, "Water the plants or not.")
	// Runs as: an admin picks from the users, the saved user selected on
	// edit, the login on new; the help says what the user brings.
	p.Users = []model.User{{ID: "u1", Name: "admin", Role: model.RoleAdmin}, {ID: "u2", Name: "miguel", Role: model.RoleRunner}}
	p.UserKeys = map[string][]string{"miguel": {"deploy_ed25519", "forge_ed25519"}}
	p.Workflow.RunAs = "miguel"
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "workflow_edit", p))
	out = buf.String()
	assert.Contains(t, out, `<label for="wf-user-w1">Runs as</label>`)
	assert.Contains(t, out, `<select id="wf-user-w1" name="user">`)
	// Each user with the instance keys their runs get, or none.
	assert.Contains(t, out, `<option value="miguel" selected data-keys="deploy_ed25519 forge_ed25519">miguel (runner) · keys: deploy_ed25519, forge_ed25519</option>`)
	assert.Contains(t, out, `<option value="admin" data-keys="">admin (admin) · no keys</option>`)
	assert.Contains(t, out, `<option value="">Whoever starts it</option>`)
	assert.Contains(t, out, "The keys are the instance keys the user gets; a step's workspace adds its own keys.")
	p.Workflow.RunAs = ""
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "workflow_edit", p))
	assert.Contains(t, buf.String(), `<option value="" selected>Whoever starts it</option>`, "no user set: the starter")
	p.Workflow.RunAs = "miguel"
	assert.Contains(t, out, "Every step runs as this user: with their permissions, the Claude token granted to their role or user, and the keys and secrets of the workspace each step's folder is in.")
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "workflow_new", p))
	out = buf.String()
	assert.Contains(t, out, `<select id="wf-user-new" name="user">`)
	assert.Contains(t, out, `<option value="admin" selected data-keys="">admin (admin) · no keys</option>`, "the login on a new workflow")

	runner := &model.User{Name: "ops", Role: model.RoleRunner}
	p.User, p.Permissions, p.Users = runner, permission.Set{User: runner, Role: permission.Policy{"workflow/edit": permission.Allow}}, nil
	p.Workflow.Steps = nil
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "workflow_new", p))
	out = buf.String()
	assert.Contains(t, out, `data-step-add="job"`)
	assert.NotContains(t, out, `data-step-add="prompt"`, "no prompt/edit")
	assert.Contains(t, out, "No steps yet. Start with a command.")
	// A runner reads who the workflow runs as: themself on new, posted
	// hidden, and the saved user on edit, which the form leaves alone.
	assert.Contains(t, out, `<span class="label">Runs as</span>
			<span class="wf-runas">ops</span><input type="hidden" name="user" value="ops">`)
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "workflow_edit", p))
	assert.Contains(t, buf.String(), `<span class="wf-runas">miguel</span>`)
	assert.NotContains(t, buf.String(), `name="user"`)
	p.Workflow.RunAs = ""
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "workflow_edit", p))
	assert.Contains(t, buf.String(), `<span class="wf-runas">Whoever starts it</span>`)
}

// threadRun is a finished run of Porec watering: its input and Claude's
// answer.
func threadRun(answer string) view.WorkflowRun {
	start := time.Now().Add(-2 * time.Minute)
	end := start.Add(90 * time.Second)
	return view.WorkflowRun{
		WorkflowRun: model.WorkflowRun{
			ID: "wr1", WorkflowID: "w1", Name: "Porec watering", UserName: "claude", State: model.StateOK,
			StartedAt: &start, FinishedAt: &end, Input: "I didn't water yesterday", Answer: answer,
		},
		Question: "I didn't water yesterday",
		Agent:    "claude",
	}
}

// The Workflows page lists the recent runs, each a line with its state
// on the dot by the link to the run, and Claude's answer in full, in
// markdown with raw HTML shown as text, folded under a chevron; a run
// without an answer has no chevron, and the input stays out.
func TestView_RenderWorkflowsRecentRuns(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	long := strings.Repeat("Rain all week. ", 40)
	failed := threadRun("")
	failed.ID, failed.State = "wr2", model.StateFailed
	running := threadRun("")
	running.ID, running.State, running.FinishedAt = "wr3", model.StateRunning, nil
	p.WorkflowRuns = []view.WorkflowRun{threadRun("## Skip watering\n\nToday is **rainy**.\n\n<script>alert(1)</script>\n\n" + long), failed, running}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "workflows", p))
	out := buf.String()
	assert.Contains(t, out, "Recent runs")
	assert.Contains(t, out, `<a href="/workflows/w1/runs/wr1" title="ok"><i class="dot ok"></i>Porec watering</a>`)
	assert.Contains(t, out, `<i class="dot failed"></i>`)
	assert.Contains(t, out, `<i class="dot running"></i>`)
	assert.NotContains(t, out, "I didn&#39;t water yesterday", "no input")
	assert.Contains(t, out, `<tr class="wf-answer-row" id="wf-answer-wr1" data-answer-of="wr1"><td colspan="3">`, "the answer open from the start")
	assert.NotContains(t, out, `data-answer-of="wr1" hidden`)
	assert.Contains(t, out, "<strong>rainy</strong>")
	assert.Contains(t, out, strings.TrimSpace(long), "the answer in full")
	assert.NotContains(t, out, "<script>alert(1)</script>")
	assert.Contains(t, out, "&lt;script&gt;")
	assert.Equal(t, 1, strings.Count(out, "data-answer-toggle"), "a chevron for the run with an answer alone")
	assert.Contains(t, out, `data-answer-toggle aria-expanded="true" aria-controls="wf-answer-wr1" title="Hide the answer" aria-label="Hide the answer"`)
	assert.NotContains(t, out, `wf-answer-wr2`)
	assert.NotContains(t, out, "No workflow has run yet.")

	p.WorkflowRuns = nil
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "workflows", p))
	out = buf.String()
	assert.Contains(t, out, `<p class="empty-line">No workflow has run yet.</p>`)
	assert.NotContains(t, out, "wf-answer-row")
}

// A workflow's page lists its runs as Workflows does: the state on the
// dot by the run's link, how it ended, and Claude's reply folded under it.
func TestView_RenderWorkflowRuns(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Workflow = &view.Workflow{Workflow: model.Workflow{ID: "w1", Name: "Porec watering"}}
	failed := threadRun("")
	failed.ID, failed.State, failed.Message = "wr2", model.StateFailed, "step 2 failed"
	p.WorkflowRuns = []view.WorkflowRun{threadRun("Skip **watering**."), failed}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "workflow", p))
	out := buf.String()
	assert.Contains(t, out, `<i class="dot ok" title="ok"></i>`)
	assert.Contains(t, out, `<i class="dot failed" title="failed"></i>`)
	assert.Contains(t, out, `<small class="faint wf-desc">step 2 failed</small>`)
	assert.Contains(t, out, `data-answer-of="wr1" hidden><td colspan="4">`, "folded on the workflow's page")
	assert.Contains(t, out, `data-answer-toggle aria-expanded="false" aria-controls="wf-answer-wr1" title="Show the answer"`)
	assert.Contains(t, out, "Skip <strong>watering</strong>.")
	assert.Equal(t, 1, strings.Count(out, "data-answer-toggle"))
	assert.NotContains(t, out, "I didn&#39;t water yesterday", "no input")
}

// A run's page is its thread: the input under who started it and the
// answer under Claude, in markdown with raw HTML shown as text, above
// the steps.
func TestView_RenderWorkflowRunThread(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Workflow = &view.Workflow{Workflow: model.Workflow{ID: "w1", Name: "Porec watering"}}
	p.WorkflowRuns = []view.WorkflowRun{threadRun("Skip **watering**.\n\n<script>alert(1)</script>")}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "workflow_run", p))
	out := buf.String()
	assert.Contains(t, out, `data-workflow-thread="wr1"`)
	assert.NotContains(t, out, "data-follow=", "a run that ended is not followed")
	assert.Contains(t, out, `class="note instruction done"`)
	assert.Contains(t, out, "I didn&#39;t water yesterday")
	assert.Contains(t, out, "<strong>watering</strong>")
	assert.NotContains(t, out, "<script>alert(1)</script>")
	assert.Contains(t, out, "&lt;script&gt;")
	assert.Contains(t, out, "Steps</h3>")

	// While it runs the reply works, and the page follows the workflow.
	run := threadRun("")
	run.State, run.FinishedAt = model.StateRunning, nil
	p.WorkflowRuns = []view.WorkflowRun{run}
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "workflow_run", p))
	out = buf.String()
	assert.Contains(t, out, `data-follow="w1" data-page="/workflows/w1/runs/wr1"`)
	assert.Contains(t, out, `<span class="badge note-state working">working</span>`)
	assert.Contains(t, out, "Claude replies here when the run ends")
}

// A run of commands alone answers with how it ended, no markdown, and
// without input says what ran.
func TestView_RenderWorkflowRunOutcome(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Workflow = &view.Workflow{Workflow: model.Workflow{ID: "w2", Name: "Deploy vhosts"}}
	run := threadRun("")
	run.Name, run.WorkflowID, run.Question, run.Agent, run.Input = "Deploy vhosts", "w2", "", "", ""
	run.State, run.Message = model.StateFailed, "malta/dockerfiles / push failed"
	p.WorkflowRuns = []view.WorkflowRun{run}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "workflow_run", p))
	out := buf.String()
	assert.Contains(t, out, "Ran Deploy vhosts, with no input.")
	assert.Contains(t, out, "The run ended failed, malta/dockerfiles / push failed, in 1m30s.")
	assert.Contains(t, out, `class="note instruction failed"`)
	assert.NotContains(t, out, "No answer")
}

// The thread's state is a note's: working, done or failed.
func TestWorkflowRun_NoteState(t *testing.T) {
	t.Parallel()
	run := view.WorkflowRun{WorkflowRun: model.WorkflowRun{State: model.StateRunning}}
	assert.True(t, run.Running())
	assert.Equal(t, model.NoteWorking, run.NoteState())
	run.State = model.StateOK
	assert.Equal(t, model.NoteDone, run.NoteState())
	run.State = model.StateStopped
	assert.Equal(t, model.NoteFailed, run.NoteState())
}

// The Workflows table names each workflow with its schedules under it,
// its last run and who made it, and the row's menu edits it: the chain
// of its steps and a link to the page the name is stay off the row.
func TestView_RenderWorkflowsTable(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Workflows = []view.Workflow{
		{Workflow: model.Workflow{ID: "w1", Name: "release", UserName: "ops"}, Steps: []view.Step{{Folder: "infra", Job: "plan"}}, CanRun: true, CanChange: true,
			Schedules: []model.Schedule{{ID: "s1", WorkflowID: "w1", Expression: "0 7 * * 1-5", Enabled: 1, UserName: "ops"}, {ID: "s2", WorkflowID: "w1", Expression: "@weekly", UserName: "admin"}}},
		{Workflow: model.Workflow{ID: "w2", Name: "ship", UserName: "ops"}, RunAs: "miguel", Steps: []view.Step{{Folder: "infra", Job: "plan"}}, Running: "wr9"},
		{Workflow: model.Workflow{ID: "w3", Name: "stale", UserName: "ops"}, Steps: []view.Step{{Folder: "gone/infra", Job: "plan"}}, RunDenied: "Run needs gone/infra / plan"},
	}
	p.Workflows[0].RunAs = "ops"
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "workflows", p))
	out := buf.String()
	// A workflow without Run says which step keeps it.
	assert.Contains(t, out, `<span class="faint run-denied">Run needs gone/infra / plan</span>`)
	assert.NotContains(t, out, `data-workflow-run="w3"`)
	assert.Contains(t, out, `<thead><tr><th>Workflow</th><th>Last run</th><th>By</th><th></th></tr></thead>`)
	// By names the owner, and the user the runs go as when another.
	assert.Contains(t, out, `<span class="user-name">ops</span>`)
	assert.Equal(t, 1, strings.Count(out, `· runs as miguel</span>`), "ship runs as miguel; release as its owner")
	assert.NotContains(t, out, `data-label="Steps"`)
	assert.Contains(t, out, `<a href="/workflows/w1">release</a>`)
	assert.Contains(t, out, `<small class="faint wf-desc" title="Scheduled, as ops"><svg class="ic lead"><use href="#i-clock"/></svg><code>0 7 * * 1-5</code></small>`)
	assert.Contains(t, out, `<small class="faint wf-desc" title="A schedule, disabled, as admin"><svg class="ic lead"><use href="#i-clock"/></svg><code>@weekly</code> disabled</small>`)
	assert.Contains(t, out, `<a class="more-item" role="menuitem" href="/workflows/w1/edit">`)
	assert.NotContains(t, out, `<span>Open</span>`, "the name is the link")
	assert.Contains(t, out, `href="/workflows/w2/runs/wr9"`)
	assert.Equal(t, 1, strings.Count(out[strings.Index(out, `<main id="main"`):], `href="/workflows/w1/edit"`), "Edit for the owner's alone")
}

// The edit page ends with the workflow's schedules and the form that
// adds one, which posts the workflow's ID and no user: the runs go as
// the workflow's user, whom the lead and each row name.
func TestView_RenderWorkflowEditSchedule(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	p.Users = []model.User{{ID: "u1", Name: "admin", Role: model.RoleAdmin}, {ID: "u2", Name: "ops", Role: model.RoleRunner}}
	p.Workflow = &view.Workflow{Workflow: model.Workflow{ID: "w1", Name: "release", UserName: "admin"}, RunAs: "miguel", Steps: []view.Step{{Kind: model.StepJob, Folder: "infra", Job: "plan"}}, CanRun: true, CanChange: true}
	p.SideWorkflows = []view.Workflow{*p.Workflow}
	next := time.Now()
	p.Schedules = []model.Schedule{{ID: "s1", WorkflowID: "w1", Expression: "0 7 * * 1-5", Enabled: 1, UserName: "ops", NextRunAt: &next}}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "workflow_edit", p))
	out := buf.String()
	block := out[strings.Index(out, `<section class="block" id="schedule">`):]
	assert.Contains(t, block, "<h3>Schedule <small>Runs the workflow on a cron expression, as miguel, the user set on the workflow. Each run needs permission to run every step at that time.</small></h3>")
	assert.Contains(t, block, `<form class="form inline" data-api="/api/schedules" data-method="POST" data-next="reload" data-notice="Schedule added">`)
	assert.Contains(t, block, `<input type="hidden" name="workflow" value="w1">`)
	assert.Contains(t, block, `<input type="text" id="sch-expr" name="expression" placeholder="0 3 * * *  or  @daily" spellcheck="false" autocomplete="off" required>`)
	assert.NotContains(t, block, `Run as</label>`, "the workflow's user runs it")
	assert.NotContains(t, block, `name="user"`)
	assert.Contains(t, block, `<td data-label="Runs as"><span class="user-ref">`)
	assert.Contains(t, block, `<span class="user-name">miguel</span>`, "the row names the workflow's user, not the schedule's")
	assert.NotContains(t, block, `data-submit`, "no pick on the row")
	assert.NotContains(t, block, `name="triage"`, "no triage for a workflow")
	assert.NotContains(t, block, `name="folder"`)
	assert.Contains(t, block, `<td class="key" data-label="Expression"><code>0 7 * * 1-5</code></td>`)
	assert.Contains(t, block, `<span class="badge ok">enabled</span>`)
	assert.Contains(t, block, `data-confirm="Delete the schedule 0 7 * * 1-5 for release?"`)
	assert.Contains(t, block, `data-notice="Schedule deleted"`, "the reload after a delete says what it did")
	assert.Contains(t, block, `data-notice="Schedule disabled"`)
	assert.NotContains(t, block, "Not scheduled.")
	assert.True(t, strings.Index(out, `data-api="/api/workflows/w1"`) < strings.Index(out, `id="schedule"`), "the schedule block ends the page")
	editor := strings.Index(out, `data-workflow data-api="/api/workflows/w1"`)
	assert.True(t, editor > 0)
	assert.True(t, editor+strings.Index(out[editor:], "</form>") < strings.Index(out, `id="schedule"`), "the editor's form closes before the schedule block: no nested forms")

	runner := &model.User{ID: "u2", Name: "ops", Role: model.RoleRunner}
	p.User, p.Permissions, p.Users, p.Schedules = runner, permission.Set{User: runner, Role: permission.Policy{"workflow/edit": permission.Allow}}, nil, nil
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "workflow_edit", p))
	block = buf.String()[strings.Index(buf.String(), `id="schedule"`):]
	assert.Contains(t, block, "Runs the workflow on a cron expression, as miguel, the user set on the workflow. Each run needs")
	assert.NotContains(t, block, "Run as</label>", "no picker without admin")
	assert.Contains(t, block, `<p class="empty-line">Not scheduled.</p>`)
	assert.Contains(t, block, `<input type="hidden" name="workflow" value="w1">`)
}

// Scheduled lists a workflow's schedule as a workflow, the name linked to
// it, and the delete confirm names the workflow.
func TestView_RenderSchedulesWorkflow(t *testing.T) {
	t.Parallel()
	v, err := view.New()
	assert.NoError(t, err)
	p := admin(t)
	next := time.Now()
	p.WorkflowNames = map[string]string{"w1": "release"}
	p.SideWorkflows = []view.Workflow{{Workflow: model.Workflow{ID: "w1", Name: "release", UserName: "ops"}, RunAs: "miguel"}}
	p.Users = []model.User{{ID: "u1", Name: "admin", Role: model.RoleAdmin}, {ID: "u2", Name: "ops", Role: model.RoleRunner}}
	p.UserKeys = map[string][]string{"ops": {"deploy_ed25519"}}
	p.Schedules = []model.Schedule{
		{ID: "s1", WorkflowID: "w1", Expression: "0 7 * * 1-5", Enabled: 1, UserName: "ops", NextRunAt: &next},
		{ID: "s2", Folder: "infra", Job: "plan", Expression: "@daily", Enabled: 1, UserName: "ops", NextRunAt: &next},
	}
	var buf bytes.Buffer
	assert.NoError(t, v.Render(&buf, "schedules", p))
	out := buf.String()
	// The row's pick names each user's keys in the short form, the
	// form's with the role, under a line on what the keys are.
	assert.Contains(t, out, `<option value="ops" selected data-keys="deploy_ed25519">ops · keys: deploy_ed25519</option>`)
	assert.Contains(t, out, `<option value="admin" data-keys="">admin · no keys</option>`)
	assert.Contains(t, out, `<option value="ops" data-keys="deploy_ed25519">ops (runner) · keys: deploy_ed25519</option>`)
	assert.Contains(t, out, `<small class="faint">The keys are the instance keys the user gets; the job's workspace adds its own keys.</small>`)
	// The workflow's row names the workflow's user with no pick; the
	// job's row keeps the admin's pick.
	assert.Contains(t, out, `<td data-label="Runs as"><span class="user-ref">`)
	assert.Contains(t, out, `<span class="user-name">miguel</span>`)
	assert.Equal(t, 1, strings.Count(out, `data-submit aria-label="Runs as"`))
	assert.Equal(t, "miguel", p.ScheduleUser(p.Schedules[0]))
	assert.Equal(t, "ops", p.ScheduleUser(p.Schedules[1]))
	assert.Equal(t, "miguel", p.WorkflowUser(p.Schedules[0]))
	assert.Equal(t, "", p.WorkflowUser(p.Schedules[1]), "a job's schedule")
	// A workflow with no user of its own: the schedule's user, with the
	// admin's pick on the row.
	p.SideWorkflows[0].RunAs = ""
	buf.Reset()
	assert.NoError(t, v.Render(&buf, "schedules", p))
	assert.Equal(t, "", p.WorkflowUser(p.Schedules[0]))
	assert.Equal(t, "ops", p.ScheduleUser(p.Schedules[0]))
	assert.Equal(t, 2, strings.Count(buf.String(), `data-submit aria-label="Runs as"`), "both rows take a pick")
	p.SideWorkflows[0].RunAs = "miguel"
	assert.Contains(t, out, `<td data-label="Folder"><svg class="ic lead"><use href="#i-flow"/></svg>workflow</td>`)
	assert.Contains(t, out, `<td class="key" data-label="Job"><a href="/workflows/w1">release</a></td>`)
	assert.Contains(t, out, `data-confirm="Delete the schedule 0 7 * * 1-5 for release?"`)
	assert.Contains(t, out, `<td data-label="Folder"><a href="/f/_/infra">infra</a></td>`)
	assert.Contains(t, out, `data-confirm="Delete the schedule @daily for infra / plan?"`)
	assert.Contains(t, out, "Schedule a workflow from its edit page.")
}
