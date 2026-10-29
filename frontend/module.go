package frontend

import (
	"errors"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/phpscript"
)

// module serves a page of a plugin written in PHP.
//
// Two things come back from a module and the handler tells them apart by
// what the Manager answered. A file under the module's public tree is
// written as it stands, with its content type and no layout, which is how
// the module's own JavaScript and CSS load. Anything else is what
// index.php echoed, and it becomes the body of an ordinary task-ui page:
// the header, the sidebar and the footer are the instance's.
//
// The request is not rewritten on the way in, so the module reads the
// whole path out of $_SERVER["REQUEST_URI"], prefix and all.
func (h *Handlers) module(w http.ResponseWriter, r *http.Request) {
	if !h.can(r, "module") {
		h.refuse(w, r)
		return
	}
	out, err := h.opts.Modules.Render(r.Context(), w, r)
	switch {
	case errors.Is(err, phpscript.ErrNotFound):
		h.missing(w, r, "No module answers "+r.URL.Path+".")
		return
	case err != nil:
		// The trace of this request holds the span the module recorded;
		// the page says what went wrong and names the module.
		slog.Error("rendering a module", "path", r.URL.Path, "err", err)
		http.Error(w, "the module failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if out.File {
		for name, values := range out.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}
		// A module's asset is served from the instance, so it carries the
		// same caching as the instance's own: short, and revalidated.
		w.Header().Set("Cache-Control", "public, max-age=300")
		w.WriteHeader(out.Status)
		if _, err := w.Write(out.Body); err != nil {
			slog.Debug("writing a module's file", "path", r.URL.Path, "err", err)
		}
		return
	}

	name, _, err := phpscript.Route(r.URL.Path)
	if err != nil {
		h.missing(w, r, "No module answers "+r.URL.Path+".")
		return
	}
	found, ok := h.opts.Modules.Module(name)
	if !ok {
		h.missing(w, r, "No module answers "+r.URL.Path+".")
		return
	}

	p := h.page(r, found.Label())
	p.Crumbs = view.Crumbs(found.Label())
	p.Module = &view.Module{
		Module: *found,
		Body:   template.HTML(out.Body), //nolint:gosec // a module is code the instance runs, see view.Module.Body
		Took:   out.Took,
	}
	h.render(w, out.Status, "module", p)
}
