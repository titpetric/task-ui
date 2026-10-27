//go:build integration

package frontend_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestEnvFile(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	ctx := context.Background()
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "infra", Name: "TOKEN", Content: "t"}))
	assert.NoError(t, f.opts.Secrets.Set(ctx, &model.Secret{Scope: "all", Name: "CDN", Content: "c"}))

	missing := f.page(t, asAdmin, "/f/internal/infra/env/.env")
	assert.Contains(t, missing, `data-api="/api/env/infra/.env"`)
	assert.Contains(t, missing, "></textarea>")
	assert.Contains(t, missing, "<td class=\"key\">CDN</td>")
	assert.Contains(t, missing, "<td class=\"key\">TOKEN</td>")

	assert.NoError(t, f.opts.Env.Save(ctx, f.infra, envfile.File{Path: ".env", Content: "A=1\n"}))
	managed := f.page(t, asAdmin, "/f/internal/infra/env/.env")
	assert.Contains(t, managed, ">A=1\n</textarea>")

	// With no path, the form that adds one names its path.
	add := f.page(t, asAdmin, "/f/internal/infra/env")
	assert.Contains(t, add, `data-api="/api/env/infra/{path}"`)
	assert.Contains(t, add, `name="path"`)
	assert.Equal(t, http.StatusForbidden, f.get(t, asViewer, "/f/internal/infra/env").Status)

	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/f/internal/nowhere/env/.env").Status)
	assert.Equal(t, http.StatusNotFound, f.get(t, asAdmin, "/f/internal/infra/env/../x").Status)
}
