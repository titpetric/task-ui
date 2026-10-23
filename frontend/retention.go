package frontend

import (
	"net/http"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/retention"
)

// retentionKinds describe the kinds of data on the Data retention page,
// in its order.
var retentionKinds = map[string]view.RetentionBlock{
	retention.Recordings: {Title: "Run recordings", One: "run", Many: "runs",
		Desc: "The terminal recording of each finished run, the .cast file Replay and the output read. Only the file goes: the run, who started it, how it ended and its artifacts stay, and its page says the recording expired."},
	retention.Artifacts: {Title: "Artifact history", One: "version", Many: "versions",
		Desc: "The earlier versions of the artifact tree's files, which the history keeps when a run or a user writes a file again. The latest version of every file in the tree stays."},
	retention.SecretReads: {Title: "Secret read log", One: "read", Many: "reads",
		Desc: "Which run got which secret in its environment. The log keeps the last 1000 reads of each secret whatever the rule says."},
}

// dataRetention shows what each kind of data holds and the rule that
// keeps it: for a user who may see any step of the settings, and an
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
		for _, kind := range retention.Kinds {
			block := retentionKinds[kind]
			block.Kind, block.Usage, block.Rule = kind, usage[kind], settings.Rule(kind)
			p.Retention = append(p.Retention, block)
		}
	}
	h.render(w, http.StatusOK, "retention", p)
}
