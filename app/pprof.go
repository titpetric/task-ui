package app

import (
	"net/http"
	"net/http/pprof"

	chi "github.com/go-chi/chi/v5"
)

// PprofPath is where the Go runtime's own profiles answer, beside the
// telemetry dashboard at /debug/oida.
const PprofPath = "/debug/pprof"

// pprofRoutes returns the runtime profiles, behind the same admin check
// the telemetry dashboard has.
//
// A trace says which request was slow and what it queried; it does not
// say which Go function spent the time. That is what a CPU profile
// answers, and the only place to take one is the running instance:
//
//	go tool pprof https://host/debug/pprof/profile?seconds=30
//	go tool pprof https://host/debug/pprof/heap
//
// pprof.Index serves the named profiles itself by cutting /debug/pprof/
// off the request path, and chi's Mount leaves the path whole, so the
// catch-all needs no rewriting as long as PprofPath is that prefix.
func pprofRoutes() http.Handler {
	r := chi.NewRouter()
	r.Use(requireAdmin)
	r.HandleFunc("/cmdline", pprof.Cmdline)
	r.HandleFunc("/profile", pprof.Profile)
	r.HandleFunc("/symbol", pprof.Symbol)
	r.HandleFunc("/trace", pprof.Trace)
	r.HandleFunc("/*", pprof.Index)
	r.HandleFunc("/", pprof.Index)
	return r
}

// requireAdmin answers anyone but an admin of the instance with a 404,
// as authorizeTelemetry does and for the same reason: a profile holds
// whatever the process has in memory, and the endpoint does not announce
// itself to whoever asks.
func requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !authorizeTelemetry(r) {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
