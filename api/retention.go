package api

import (
	"net/http"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/retention"
)

// saveRetention sets the Data retention rules, each kind's age and size
// limit, the limit a size such as 500MB, empty for none: for admins
// alone, as the page's form sends them. The next run of @cleanup applies
// them. It answers the rules as they stand.
func (h *Handlers) saveRetention(w http.ResponseWriter, r *http.Request) {
	if !access.UserFrom(r.Context()).CanAdmin() {
		fail(w, http.StatusForbidden, "only an admin sets the data retention rules")
		return
	}
	if h.opts.Cleaner == nil {
		fail(w, http.StatusNotFound, "data retention isn't set up on this instance")
		return
	}
	var req map[string]map[string]string
	if err := readJSON(r, &req); err != nil {
		fail(w, http.StatusBadRequest, err.Error())
		return
	}
	settings, err := h.opts.Cleaner.Load(r.Context())
	if err != nil {
		failErr(w, err)
		return
	}
	rules := map[string]retention.Rule{}
	for _, kind := range retention.Kinds {
		rule := settings.Rule(kind)
		if in, ok := req[kind]; ok {
			if in["age"] != "" {
				rule.Age = in["age"]
			}
			limit, err := retention.ParseSize(in["limit"])
			if err != nil {
				fail(w, http.StatusBadRequest, kind+": "+err.Error())
				return
			}
			rule.Limit = limit
		}
		if !rule.Valid() {
			fail(w, http.StatusBadRequest, kind+": keep for 24h, 7d, 31d or never")
			return
		}
		rules[kind] = rule
	}
	settings = retention.Settings{Recordings: rules[retention.Recordings], Artifacts: rules[retention.Artifacts], SecretReads: rules[retention.SecretReads]}
	if err := h.opts.Cleaner.Save(r.Context(), settings); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}
