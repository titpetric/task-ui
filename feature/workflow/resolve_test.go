package workflow_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/workflow"
)

// A script step resolves to the folder's script job with its command,
// and refuses a step without one.
func TestRunner_ResolveScript(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	infra := &folder.Folder{ID: "infra", Dir: t.TempDir()}
	r := workflow.New(nil, nil, catalogue{"infra": infra}, nil)

	steps, err := r.Resolve(ctx, []model.WorkflowStep{{Folder: "infra", Job: folder.ScriptID, Args: model.EncodeList([]string{"go get -u ./..."})}})
	assert.NoError(t, err)
	assert.True(t, steps[0].Job.IsScript())
	assert.Equal(t, []string{"go get -u ./..."}, steps[0].Args)

	_, err = r.Resolve(ctx, []model.WorkflowStep{{Folder: "infra", Job: folder.ScriptID}})
	assert.ErrorIs(t, err, folder.ErrScriptEmpty)
	_, err = r.Resolve(ctx, []model.WorkflowStep{{Folder: "nowhere", Job: folder.ScriptID, Args: model.EncodeList([]string{"true"})}})
	assert.ErrorContains(t, err, "no job nowhere / @script")
}
