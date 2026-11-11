package phpfront

import (
	"net/http"

	"github.com/titpetric/task-ui/internal/markdown"
)

// Policy sets the content security policy of every page and asset this front
// end answers with, which is the one the Go pages carry.
//
// The Go router sets it for everything under it with one r.Use, so a page, a
// stylesheet and a refusal all carry it. Nothing in PHP can: a header set
// inside the request is the page's, and a response the Guard writes before the
// page runs would carry none. So it is a middleware at the mount, wrapped
// around the Guard rather than inside it, and the cutover keeps the line.
func Policy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", markdown.CSP)
		next.ServeHTTP(w, r)
	})
}
