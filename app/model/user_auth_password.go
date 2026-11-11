package model

// Usable reports whether the password may still sign in: a password
// always, a backup password until it has signed in once.
func (p *UserAuthPassword) Usable() bool {
	if p == nil || p.Hash == "" {
		return false
	}
	return p.Kind == AuthPassword || p.Kind == AuthBackup && p.UsedAt == nil
}
