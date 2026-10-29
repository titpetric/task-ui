package view

import (
	"html/template"
	"time"

	"github.com/titpetric/task-ui/phpscript"
)

// Module is a module's page: what its PHP echoed, and how long the
// render took.
type Module struct {
	phpscript.Module

	// Body is what index.php wrote. It reaches the page unescaped,
	// because a module's job is to render HTML: the module is code the
	// instance runs, as trusted as a Go handler, and escaping its output
	// would print the markup instead of drawing it.
	//
	// The one thing this is not is a place for anything a user typed. A
	// module escapes what it renders, which is what MiniTPL does by
	// default for a text node, an attribute and an HTML comment.
	Body template.HTML

	// Took is how long the render ran, which the footer prints.
	Took time.Duration
}

// Render is the render time as the footer shows it: milliseconds under a
// second, and seconds past it, because a page that took two minutes
// reads worse as a six-digit number of milliseconds.
func (m *Module) Render() string {
	if m.Took >= time.Second {
		return m.Took.Round(time.Millisecond).String()
	}
	return m.Took.Round(time.Microsecond).String()
}
