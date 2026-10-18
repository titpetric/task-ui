package view

// Crumb is one step of the breadcrumb. The last one has no link.
type Crumb struct {
	Text string
	Href string
}

// Crumbs builds a breadcrumb from text and href pairs; an odd last
// argument is a text without a link.
func Crumbs(pairs ...string) []Crumb {
	var crumbs []Crumb
	for i := 0; i < len(pairs); i += 2 {
		crumb := Crumb{Text: pairs[i]}
		if i+1 < len(pairs) {
			crumb.Href = pairs[i+1]
		}
		crumbs = append(crumbs, crumb)
	}
	return crumbs
}
