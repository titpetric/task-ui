package phpscript_test

import (
	"context"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

func TestModule(t *testing.T) {
	t.Parallel()
	m, err := phpscript.NewModule(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "", m.Name(), "a module that named itself nothing has no name")
	assert.Equal(t, "", m.Href())
	assert.Error(t, m.Validate())

	assert.NoError(t, m.SetName("titpetric/refinement"))
	m.SetTitle(" Refinement ")
	m.SetDescription(" Refine a design ")
	m.SetIcon(" flow ")
	m.SetVersion(" 1.2.0 ")

	assert.Equal(t, "titpetric", m.Vendor)
	assert.Equal(t, "refinement", m.Plugin)
	assert.Equal(t, "titpetric/refinement", m.Name())
	assert.Equal(t, "/m/titpetric/refinement", m.Href())
	assert.Equal(t, "Refinement", m.Title)
	assert.Equal(t, "Refine a design", m.Description)
	assert.Equal(t, "flow", m.Icon)
	assert.Equal(t, "1.2.0", m.Version)
	assert.Equal(t, "Refinement", m.Label())

	assert.ErrorContains(t, m.Validate(), "no root directory")
	m.Root = "/data/modules/titpetric/refinement"
	assert.NoError(t, m.Validate())

	assert.ErrorIs(t, m.SetName("nope"), phpscript.ErrName)
	assert.Equal(t, "titpetric/refinement", m.Name(), "a name it refused leaves the one it had")
}

func TestModuleLabel(t *testing.T) {
	t.Parallel()
	m := &phpscript.Module{Vendor: "titpetric", Plugin: "mysql"}
	assert.Equal(t, "mysql", m.Label(), "a module with no title reads as its plugin")
	m.SetTitle("MySQL")
	assert.Equal(t, "MySQL", m.Label())
}
