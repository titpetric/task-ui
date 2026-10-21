package secret

// Migration counts what Migrate did: the secrets it wrote for their
// workspace's key, those already on it, and those it left untouched
// because no key opens them or their scope's workspace has no key.
type Migration struct {
	Migrated int `json:"migrated"`
	Current  int `json:"current"`
	Failed   int `json:"failed"`
}
