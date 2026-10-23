package secret

// Migration counts what Migrate did: the secrets it wrote for the
// instance's encryption key, those already on it, and those it left
// untouched because no key opens them.
type Migration struct {
	Migrated int `json:"migrated"`
	Current  int `json:"current"`
	Failed   int `json:"failed"`
}
