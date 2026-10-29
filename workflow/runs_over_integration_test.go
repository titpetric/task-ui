//go:build integration

package workflow_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// A workflow that runs over a workspace walks its steps once per
// project: a failure ends that project's pass, a project without the
// job skips the step, and a script step runs its command in each.
func TestRunner_RunsOver(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	dir := t.TempDir()
	list := projects{}
	for _, name := range []string{"a", "b"} {
		assert.NoError(t, os.MkdirAll(filepath.Join(dir, name), 0o755))
		p := &folder.Folder{ID: "ws/" + name, Dir: filepath.Join(dir, name)}
		p.Jobs = []*folder.Job{job("build", "echo built")}
		if name == "a" {
			p.Jobs = append(p.Jobs, job("broken", "exit 3"))
		}
		list = append(list, p)
	}
	f.runner.Projects = list

	wf := &model.Workflow{Name: "maintain", RunsOver: "ws"}
	steps := []model.WorkflowStep{
		{Folder: "ws/a", Job: "build"},
		{Folder: "ws/a", Job: "broken"},
		{Folder: "ws/a", Job: folder.ScriptID, Args: model.EncodeList([]string{"echo done > marker"})},
	}
	assert.NoError(t, f.workflows.Create(context.Background(), wf, steps))

	run, err := f.runner.Start(context.Background(), wf, ops, "")
	assert.NoError(t, err)
	ended := f.wait(t, run.ID)
	assert.Equal(t, model.StateFailed, ended.State)

	got, err := f.wfRuns.Steps(context.Background(), run.ID)
	assert.NoError(t, err)
	assert.Len(t, got, 6)
	states := []string{}
	for _, step := range got {
		states = append(states, step.Folder+" "+step.State)
	}
	assert.Equal(t, []string{
		"ws/a " + model.StateOK, "ws/a " + model.StateFailed, "ws/a " + model.StepSkipped,
		"ws/b " + model.StateOK, "ws/b " + model.StepSkipped, "ws/b " + model.StateOK,
	}, states)

	_, err = os.Stat(filepath.Join(dir, "a", "marker"))
	assert.True(t, os.IsNotExist(err), "a failed before its script")
	marker, err := os.ReadFile(filepath.Join(dir, "b", "marker"))
	assert.NoError(t, err)
	assert.Equal(t, "done\n", string(marker))
}

// A workflow over a workspace without projects refuses to start.
func TestRunner_RunsOverNothing(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	wf := &model.Workflow{Name: "empty", RunsOver: "nowhere"}
	assert.NoError(t, f.workflows.Create(context.Background(), wf, []model.WorkflowStep{{Folder: "infra", Job: "build"}}))
	_, err := f.runner.Start(context.Background(), wf, ops, "")
	assert.Error(t, err)
}
