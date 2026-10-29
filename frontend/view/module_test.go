package view_test

import (
	"html/template"
	"testing"
	"time"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

func TestModule(t *testing.T) {
	t.Parallel()
	m := &view.Module{
		Module: phpscript.Module{Vendor: "titpetric", Plugin: "refinement", Title: "Refinement", Version: "0.1.0"},
		Body:   template.HTML(`<section class="refinement">a thread</section>`),
		Took:   3 * time.Millisecond,
	}
	assert.Equal(t, "titpetric/refinement", m.Name())
	assert.Equal(t, "/m/titpetric/refinement", m.Href())
	assert.Equal(t, "Refinement", m.Label())
	assert.Contains(t, string(m.Body), "a thread")
}

// The footer reads in milliseconds under a second and in seconds past
// it: a render that took two minutes reads worse as a six-digit number
// of milliseconds.
func TestModuleRender(t *testing.T) {
	t.Parallel()
	for took, want := range map[time.Duration]string{
		0:                       "0s",
		1500 * time.Microsecond: "1.5ms",
		3 * time.Millisecond:    "3ms",
		999 * time.Millisecond:  "999ms",
		2 * time.Second:         "2s",
		90 * time.Second:        "1m30s",
	} {
		m := &view.Module{Took: took}
		assert.Equal(t, want, m.Render(), took.String())
	}
}

// The module's page renders what its PHP echoed, inside the layout, with
// the render time in the footer and no iframe anywhere.
func TestModulePage(t *testing.T) {
	t.Parallel()
	p := admin(t)
	p.Path = "/m/titpetric/refinement"
	p.Module = &view.Module{
		Module: phpscript.Module{Vendor: "titpetric", Plugin: "refinement", Title: "Refinement", Description: "Refine a design"},
		Body:   template.HTML(`<section class="refinement"><h3>A thread</h3></section>`),
		Took:   4 * time.Millisecond,
	}
	body := render(t, "module", p)

	assert.Contains(t, body, `<section class="refinement"><h3>A thread</h3></section>`,
		"the module's markup is drawn, not printed")
	assert.Contains(t, body, "Refine a design")
	assert.Contains(t, body, "titpetric/refinement")
	assert.Contains(t, body, "Rendered in <span class=\"mono\">4ms</span> by phpscript.")
	assert.NotContains(t, body, "<iframe")
	assert.Contains(t, body, `<nav class="menu">`, "the layout is the instance's")
}
