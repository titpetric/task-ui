package frontend_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestArtifacts(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Artifacts.Create(ctx, &model.Artifact{RunID: model.NewID(), Folder: "infra", Job: "plan", Name: "plan.txt", Path: "a/plan.txt", ByteSize: 10}))
	assert.NoError(t, f.opts.Artifacts.Create(ctx, &model.Artifact{RunID: model.NewID(), Folder: "private", Job: "deploy", Name: "deploy.log", Path: "b/deploy.log", ByteSize: 20}))

	body := f.page(t, asAdmin, "/artifacts")
	assert.Contains(t, body, "<h2>Artifacts</h2>")
	assert.Contains(t, body, "plan.txt")
	assert.Contains(t, body, "deploy.log")
	assert.Contains(t, body, `data-delete="/api/artifacts/`)

	limited := f.page(t, asLimited, "/artifacts")
	assert.Contains(t, limited, "plan.txt")
	assert.NotContains(t, limited, "deploy.log")
	assert.NotContains(t, limited, "data-delete")

	filtered := f.page(t, asAdmin, "/artifacts?folder=private&job=deploy")
	assert.NotContains(t, filtered, "plan.txt")
	assert.Contains(t, filtered, `<option value="private" selected>`)

	empty := f.page(t, asAdmin, "/artifacts?folder=nowhere")
	assert.Contains(t, empty, "No artifacts yet")
}
