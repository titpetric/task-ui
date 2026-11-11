package model_test

import (
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

func TestFolderRef_String(t *testing.T) {
	t.Parallel()
	ref := model.FolderRef{Kind: "workflow_step", Folder: "web", Name: "release: build"}
	assert.Equal(t, "workflow step web · release: build", ref.String())
}

func TestUnder(t *testing.T) {
	t.Parallel()
	assert.True(t, model.Under("web", "web"))
	assert.True(t, model.Under("web/docs", "web"))
	assert.False(t, model.Under("website", "web"))
	assert.False(t, model.Under("acme/web", "web"))
}

func TestMoved(t *testing.T) {
	t.Parallel()
	got, ok := model.Moved("web/docs", "web", "acme/web")
	assert.True(t, ok)
	assert.Equal(t, "acme/web/docs", got)
	got, ok = model.Moved("web", "web", "acme/web")
	assert.True(t, ok)
	assert.Equal(t, "acme/web", got)
	got, ok = model.Moved("website", "web", "acme/web")
	assert.False(t, ok)
	assert.Equal(t, "website", got)
}
