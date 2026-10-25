package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/retention"
	"github.com/titpetric/task-ui/storage"
)

// dataRetention shows the Data retention table, one row per area with
// what it holds and the rule that keeps it, and the last run of
// @cleanup: for a user who may see any step of the settings, and an
// admin edits the rules.
func (h *Handlers) dataRetention(w http.ResponseWriter, r *http.Request) {
	p := h.page(r, "Data retention")
	if !p.CanSettings() {
		h.refuse(w, r)
		return
	}
	p.Crumbs = view.Crumbs("Settings", "/settings", "Data retention")
	p.Step = view.StepRetention
	if c := h.opts.Cleaner; c != nil {
		settings, _ := c.Load(r.Context())
		usage, err := c.Usage(r.Context())
		if err != nil {
			http.Error(w, "measuring the data failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		for _, area := range retention.Areas {
			p.Retention = append(p.Retention, view.RetentionRow{Area: area, Usage: usage[area.Kind], Rule: settings.Rule(area.Kind)})
		}
	}
	if last, err := h.opts.Runs.Last(r.Context(), storage.RunFilter{Folder: model.Default, Job: folder.CleanupID}); err == nil {
		p.Cleanup = &view.Run{Run: *last}
	}
	h.render(w, http.StatusOK, "retention", p)
}
