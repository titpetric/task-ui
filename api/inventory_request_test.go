package api_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
)

// withHosts sets up a hub of two hosts, dev1 and dev2, which the fake
// ssh signs in to as debian.
func (f *fixture) withHosts(t *testing.T) {
	t.Helper()
	f.withKey(t)
	hub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[{"id":"a","name":"dev1","host":"dev1","status":"up","info":{"c":8}},{"id":"b","name":"dev2","host":"dev2","status":"up","info":{"c":4}}]}`))
	}))
	t.Cleanup(hub.Close)
	res := f.call(t, asAdmin, http.MethodPut, "/integrations/beszel", api.IntegrationRequest{Values: map[string]string{"url": hub.URL, "token": "t1"}})
	assert.Equal(t, http.StatusOK, res.Status, string(res.Raw))
}

// A Run on two hosts is two runs of @host under one batch, each with
// its host, its command and a recording; Run and summarize asks Claude
// once about the batch, with every host's run, in the hosts' summary
// shape. The batch lists in /api/runs?batch=.
func TestInventoryRequest_Batch(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withHosts(t)
	assert.NoError(t, os.MkdirAll(promptOut(f, ""), 0o755))
	f.opts.Board.AgentCommand = []string{"sh", "-c", `printf '%s' "$TASK_UI_PROMPT_INSTRUCTIONS" > "` + promptOut(f, "instructions.out") + `"; cp "$TASK_UI_PROMPT_RUN" "` + promptOut(f, "run.out") + `"; printf 'Summary of: %s' "$TASK_UI_PROMPT" > "$TASK_UI_PROMPT_REPLY"`}

	started := f.call(t, asAdmin, http.MethodPost, "/inventory/run", api.InventoryRequest{Command: "echo up", Hosts: []string{"dev1", "dev2"}, Summarize: true})
	assert.Equal(t, http.StatusCreated, started.Status, string(started.Raw))
	batch := started.str("batch")
	runs := started.list("runs")
	assert.Len(t, runs, 2)
	first := runs[0].(map[string]any)["id"].(string)
	for i, host := range []string{"dev1", "dev2"} {
		run := f.finished(t, runs[i].(map[string]any)["id"].(string))
		assert.Equal(t, host, run.Target)
		assert.Equal(t, batch, run.BatchID)
		assert.Equal(t, model.HostJob, run.Job)
		assert.Equal(t, model.StateOK, run.State)
		assert.NotEqual(t, "", run.CastPath)
	}
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/runs?batch="+batch, nil).list("runs"), 2)

	var thread map[string]any
	assert.Eventually(t, func() bool {
		notes := f.call(t, asAdmin, http.MethodGet, "/prompts?run="+first, nil).list("notes")
		if len(notes) != 1 {
			return false
		}
		thread = notes[0].(map[string]any)
		replies, _ := thread["replies"].([]any)
		return len(replies) == 1 && thread["state"] == model.NoteDone
	}, 10*time.Second, 20*time.Millisecond)
	assert.Equal(t, note.BatchSummaryQuestion, thread["body"])
	instructions, err := os.ReadFile(promptOut(f, "instructions.out"))
	assert.NoError(t, err)
	assert.Equal(t, claude.HostsSummary, string(instructions))
	about, err := os.ReadFile(promptOut(f, "run.out"))
	assert.NoError(t, err)
	assert.Contains(t, string(about), "A batch of 2 runs")
	assert.Contains(t, string(about), "## dev1")
	assert.Contains(t, string(about), "## dev2")
	assert.Equal(t, 2, strings.Count(string(about), "\nup\n"))
}

// POST /api/runs runs @host again on its target, its command the args;
// it takes the hosts permission and a target.
func TestInventoryRequest_RunAgain(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withHosts(t)

	again := f.call(t, asAdmin, http.MethodPost, "/runs", api.RunRequest{Folder: model.Default, Job: model.HostJob, Target: "dev2", Args: api.RunArgs{"echo again"}})
	assert.Equal(t, http.StatusCreated, again.Status, string(again.Raw))
	run := f.finished(t, again.str("id"))
	assert.Equal(t, "dev2", run.Target)
	assert.Equal(t, "", run.BatchID)
	assert.Contains(t, string(f.call(t, asAdmin, http.MethodGet, "/runs/"+run.ID+"/frame.txt", nil).Raw), "again")

	every := f.call(t, asAdmin, http.MethodPost, "/runs", api.RunRequest{Folder: model.Default, Job: model.HostJob, Target: "*", Args: api.RunArgs{"true"}})
	assert.Equal(t, http.StatusCreated, every.Status, string(every.Raw))
	assert.NotEqual(t, "", every.str("batch_id"))

	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPost, "/runs", api.RunRequest{Folder: model.Default, Job: model.HostJob, Target: "dev2", Args: api.RunArgs{"uptime"}}).Status)
	assert.Equal(t, http.StatusBadRequest, f.call(t, asAdmin, http.MethodPost, "/runs", api.RunRequest{Folder: model.Default, Job: model.HostJob, Args: api.RunArgs{"uptime"}}).Status)
	assert.Equal(t, http.StatusBadRequest, f.call(t, asAdmin, http.MethodPost, "/runs", api.RunRequest{Folder: model.Default, Job: model.HostJob, Target: "elsewhere", Args: api.RunArgs{"uptime"}}).Status)
}

// A schedule and a workflow step take @host with a target and a
// command, which no other job takes; @shell stays refused.
func TestInventoryRequest_SchedulesAndWorkflows(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withHosts(t)

	made := f.call(t, asAdmin, http.MethodPost, "/schedules", api.ScheduleRequest{Folder: model.Default, Job: model.HostJob, Expression: "0 * * * *", Target: "*", Args: api.RunArgs{"uptime"}})
	assert.Equal(t, http.StatusCreated, made.Status, string(made.Raw))
	assert.Equal(t, "*", made.str("target"))
	assert.Equal(t, `["uptime"]`, made.str("args"))
	for _, req := range []api.ScheduleRequest{
		{Folder: model.Default, Job: model.HostJob, Expression: "0 * * * *", Args: api.RunArgs{"uptime"}},
		{Folder: model.Default, Job: model.HostJob, Expression: "0 * * * *", Target: "dev1"},
		{Folder: "infra", Job: "plan", Expression: "0 * * * *", Target: "dev1"},
		{Folder: model.Default, Job: "@shell", Expression: "0 * * * *"},
	} {
		res := f.call(t, asAdmin, http.MethodPost, "/schedules", req)
		assert.Equal(t, http.StatusBadRequest, res.Status, req.Job+" "+req.Target)
	}
	assert.Equal(t, http.StatusForbidden, f.call(t, asOps, http.MethodPost, "/schedules", api.ScheduleRequest{Folder: model.Default, Job: model.HostJob, Expression: "0 * * * *", Target: "dev1", Args: api.RunArgs{"uptime"}}).Status)

	id := f.createWorkflow(t, asAdmin, "hosts", []api.StepRequest{{Folder: model.Default, Job: model.HostJob, Target: "*", Args: api.RunArgs{"echo step"}}})
	started := f.call(t, asAdmin, http.MethodPost, "/workflows/"+id+"/runs", nil)
	assert.Equal(t, http.StatusCreated, started.Status, string(started.Raw))
	run := f.workflowRun(t, id, started.str("id"))
	assert.Equal(t, model.StateOK, run["state"])
	step := run["steps"].([]any)[0].(map[string]any)
	stepRun := f.finished(t, step["run_id"].(string))
	assert.NotEqual(t, "", stepRun.BatchID)
	assert.Len(t, f.call(t, asAdmin, http.MethodGet, "/runs?batch="+stepRun.BatchID, nil).list("runs"), 2)

	bad := f.call(t, asAdmin, http.MethodPost, "/workflows", api.WorkflowRequest{Name: "no host", Steps: []api.StepRequest{{Folder: model.Default, Job: model.HostJob, Args: api.RunArgs{"uptime"}}}})
	assert.Equal(t, http.StatusBadRequest, bad.Status)
	ops := f.call(t, asOps, http.MethodPost, "/workflows/"+id+"/runs", nil)
	assert.Equal(t, http.StatusForbidden, ops.Status)
}
