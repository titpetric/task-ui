package secret

// Migration counts what Migrate did: the secrets it wrote for the
// instance's encryption key, those already on it, and those it left
// untouched because no key opens them.
//
// Migrate and MigrateInstance move rows written under older schemes onto the instance's encryption key.
// A row that key does not open decrypts with the legacy keys: the SSH keys of old and, for Migrate, the
// workspaces' encryption keys. Migrate removes a row only after it wrote the replacement and read it
// back. MigrateInstance checks that the new value decrypts to the old before it writes it, and reads the
// row back after. A second run of either finds every row current and does nothing.
type Migration struct {
	Migrated int `json:"migrated"`
	Current  int `json:"current"`
	Failed   int `json:"failed"`
}
