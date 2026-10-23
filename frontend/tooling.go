package frontend

import (
	"time"

	"github.com/titpetric/task-ui/frontend/view"
)

// tooling returns what atkins has recorded for the tree, for the panel
// on the runs page.
//
// It reports commands and not runs, so it says nothing about who ran
// what and needs no permission of its own beyond the page it sits on: a
// reader of the runs page may already see that the tree runs go test,
// because the runs are there.
//
// Nil without a prober, and nil when atkins has recorded nothing, which
// is the usual answer on a tree nobody has run a job in yet. The charts
// count their slots in now's time zone, the activity charts' own.
func (h *Handlers) tooling(now time.Time) *view.Tooling {
	if h.opts.Tooling == nil || h.opts.Root == "" {
		return nil
	}
	return view.NewTooling(h.opts.Tooling.Probe(h.opts.Root), now)
}
