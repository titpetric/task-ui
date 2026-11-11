package api

import (
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/retention"
)

// saveRetention sets the Data retention rules, each area's age and, where
// its items have a size, the size limit, a size such as 500MB or the
// bytes, empty or 0 for none: for admins alone, as the page's form sends
// them. An area the request leaves out keeps its rule. The next run of
// @cleanup applies them. It answers the rules as they stand.
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
		if !retention.AreaOf(kind).Limited() {
			rule.Limit = 0
		}
		if !rule.Valid() {
			fail(w, http.StatusBadRequest, kind+": keep for "+strings.Join(retention.Ages, ", "))
			return
		}
		settings[kind] = rule
	}
	if err := h.opts.Cleaner.Save(r.Context(), settings); err != nil {
		failErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, settings)
}
