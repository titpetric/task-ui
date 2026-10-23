package view

import (
	"fmt"
)

// MenuDetail is a line of the details at the top of a [...] menu, what a
// row holds that its columns leave out: a key's fingerprint, when a
// run started, who made a token.
type MenuDetail struct {
	Label string
	Value string
	// Title is the value in full, on a hover, such as the time behind
	// "3 days ago".
	Title string
	// Mono sets the value in the data font: fingerprints, paths, IDs.
	Mono bool
}

// NewMenuDetail returns a detail from its label, its value and pairs
// of an option and its value: title, and mono with any value.
func NewMenuDetail(label, value string, opts ...string) (MenuDetail, error) {
	d := MenuDetail{Label: label, Value: value}
	if len(opts)%2 != 0 {
		return d, fmt.Errorf("menu detail %q: option %q without a value", label, opts[len(opts)-1])
	}
	for i := 0; i < len(opts); i += 2 {
		switch opts[i] {
		case "title":
			d.Title = opts[i+1]
		case "mono":
			d.Mono = true
		default:
			return d, fmt.Errorf("menu detail %q: no option %q", label, opts[i])
		}
	}
	return d, nil
}
