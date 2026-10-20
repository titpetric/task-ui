package view

import (
	"net/url"
)

// Pager links a page of a list to the pages either side of it: Prev to
// the newer rows, Next to the older ones. An empty link has no page.
type Pager struct {
	Prev string
	Next string
}

// NewPager builds the links of a page at path with the query, less its
// cursors: Prev reads after the run prev names, Next before the run next
// names. It returns nil without either.
func NewPager(path string, query url.Values, prev, next string) *Pager {
	if prev == "" && next == "" {
		return nil
	}
	link := func(key, id string) string {
		if id == "" {
			return ""
		}
		q := url.Values{}
		for k, v := range query {
			if k != "before" && k != "after" {
				q[k] = v
			}
		}
		q.Set(key, id)
		return path + "?" + q.Encode()
	}
	return &Pager{Prev: link("after", prev), Next: link("before", next)}
}
