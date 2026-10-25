package config

// Extra is a rule one section adds to the rules every section gets,
// which the section's All leaves out.
type Extra struct {
	Section string `json:"section"`
	Entry
}

// Path returns the path a check names: <section>/<rule>.
func (e Extra) Path() string {
	return e.Section + "/" + e.Key
}
