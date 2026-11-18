package phpfront_test

import (
	"context"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpfront"
)

// A report inside a folder links its files through that folder.
func TestArtifactHref(t *testing.T) {
	t.Parallel()
	assert.Equal(t, phpfront.ArtifactPrefix, phpfront.ArtifactHref(""))
	assert.Equal(t, phpfront.ArtifactPrefix+"thread-1/", phpfront.ArtifactHref("/thread-1/"))
}

func TestModule(t *testing.T) {
	t.Parallel()
	m, err := phpfront.NewModule(context.Background())
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

	assert.ErrorIs(t, m.SetName("nope"), phpfront.ErrName)
	assert.Equal(t, "titpetric/refinement", m.Name(), "a name it refused leaves the one it had")
}

func TestModuleSidebar(t *testing.T) {
	t.Parallel()
	m := &phpfront.Module{Vendor: "titpetric", Plugin: "runs", Root: "/x"}
	m.SetTitle(" Runs ")
	m.SetDescription(" Every run ")
	m.SetIcon(" list ")

	assert.False(t, m.GetSidebar())
	assert.Equal(t, "/m/titpetric/runs", m.GetSideBarHref(), "a row with no link of its own opens the module")
	m.EnableSideBar()
	m.SetSideBarHref(" /runs ")
	m.SetSideBarRule(" run ")
	assert.True(t, m.GetSidebar())
	assert.Equal(t, "/runs", m.GetSideBarHref())
	assert.Equal(t, "run", m.GetSideBarRule())
	assert.Equal(t, "Runs", m.GetTitle())
	assert.Equal(t, "Every run", m.GetDescription())
	assert.Equal(t, "list", m.GetIcon())

	fresh, err := phpfront.NewModule(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, phpfront.ModuleRule, fresh.GetSideBarRule(), "a module that set no rule takes the module rule")
}

func TestModule_SetTables(t *testing.T) {
	t.Parallel()
	m := &phpfront.Module{}
	m.SetTables([]any{"run_summary", "run", " ", "run_queued"})
	assert.Equal(t, []string{"run", "run_queued", "run_summary"}, m.Tables,
		"the tables read in the order a schema is listed in")
}

func TestModuleLabel(t *testing.T) {
	t.Parallel()
	m := &phpfront.Module{Vendor: "titpetric", Plugin: "mysql"}
	assert.Equal(t, "mysql", m.Label(), "a module with no title reads as its plugin")
	m.SetTitle("MySQL")
	assert.Equal(t, "MySQL", m.Label())
}

func TestModule_Project(t *testing.T) {
	t.Parallel()
	m := &phpfront.Module{Vendor: "titpetric", Plugin: "refinement"}
	assert.Equal(t, 16, len(m.Project()))
	assert.True(t, strings.HasPrefix(m.Project(), "refinem_"), "the project keeps the plugin prefix")
	assert.Equal(t, "", (&phpfront.Module{}).Project(), "a module that named itself nothing has no project")
}
