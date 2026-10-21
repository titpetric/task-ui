package frontend

import (
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/storage"
)

// pageSizes are the rows a page of runs may hold, the first by default.
var pageSizes = []int{15, 50, 100}

// pageSize returns the rows a page holds: ?n= when one of the
// page sizes. It sets n in the query, for the links of the pager.
func pageSize(q url.Values) int {
	n, err := strconv.Atoi(q.Get("n"))
	if err != nil || !slices.Contains(pageSizes, n) {
		n = pageSizes[0]
	}
	q.Set("n", strconv.Itoa(n))
	return n
}

// paged trims a page read with one row more than n, the row that tells
// whether the list goes on, and returns the cursors of the pages either
// side: prev, the newest row's ID when newer rows exist, and next, the
// oldest row's when older ones do. A page read after a cursor has the
// extra row on top, any other at the bottom.
func paged[T any](rows []T, n int, before, after string, id func(T) string) ([]T, string, string) {
	more := len(rows) > n
	newer, older := before != "", more
	if after != "" {
		if more {
			rows = rows[1:]
		}
		newer, older = more, true
	} else if more {
		rows = rows[:n]
	}
	if len(rows) == 0 {
		return rows, "", ""
	}
	var prev, next string
	if newer {
		prev = id(rows[0])
	}
	if older {
		next = id(rows[len(rows)-1])
	}
	return rows, prev, next
}

// runs lists every run the user may see, filtered by the query, a page
// at a time.
func (h *Handlers) runs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	user := access.UserFrom(r.Context())
	p := h.page(r, "Runs")
	p.Crumbs = view.Crumbs("Runs")
	n := pageSize(q)
	p.Form = map[string]string{"folder": q.Get("folder"), "job": q.Get("job"), "state": q.Get("state"), "n": q.Get("n")}
	filter := storage.RunFilter{
		Folder: q.Get("folder"), Folders: user.FolderList(), Job: q.Get("job"), Dir: q.Get("dir"), State: q.Get("state"),
		Before: q.Get("before"), After: q.Get("after"), Limit: n + 1,
	}
	runs := h.visibleRuns(r.Context(), user, filter)
	var prev, next string
	p.Runs, prev, next = paged(runs, n, filter.Before, filter.After, func(run view.Run) string { return run.ID })
	p.Pager = view.NewPager("/runs", q, prev, next)
	filter.Before, filter.After, filter.Limit = "", "", activityLimit
	p.Activity = h.activity(r.Context(), user, filter, p.Now, false)
	p.Summarize = h.claudeFor(r, p.User)
	h.render(w, http.StatusOK, "runs", p)
}

// run shows one run: its terminal while live, its details and its
// artifacts.
func (h *Handlers) run(w http.ResponseWriter, r *http.Request) {
	user := access.UserFrom(r.Context())
	row, err := h.opts.Runs.Get(r.Context(), param(r, "id"))
	if err != nil || !user.CanSee(row.Folder) {
		h.missing(w, r, "No run with that ID.")
		return
	}
	run := h.framed(h.runView(*row))
	where := run.Folder
	if run.Dir != "" {
		where += "/" + run.Dir
	}
	p := h.page(r, where+" / "+run.Job)
	p.Run = &run
	p.Crumbs = view.Crumbs("Runs", "/runs", where+" / "+run.Job)
	// A run of a directory's job finds the job there; the folder's rules
	// stand for it.
	if f := h.opts.Catalogue.Find(run.Folder); f != nil {
		p.Folder = f
		if sub := f.Sub(run.Dir); sub != nil {
			p.Job = sub.Job(run.Job)
		}
	}
	artifacts, _ := h.opts.Artifacts.List(r.Context(), storage.ArtifactFilter{RunID: run.ID, Limit: 200})
	p.Artifacts = artifacts
	h.render(w, http.StatusOK, "run", p)
}
