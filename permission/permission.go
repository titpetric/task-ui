// Package permission decides what a user may do, from the rules of
// `config/permissions.yml` and the policies of the user and the user's
// role.
//
// A policy entry allows or denies a path, <section>/<rule> or an
// application permission; no entry inherits. A user may do what the
// user's entry or the role's entry allows, unless either denies it, and
// nothing that neither sets. A <section>/all entry stands for every
// rule of the section, and a subject's own entry for a rule overrides
// its all. The admin role allows everything, whatever the policies say.
//
// Checks name their path as a string literal, so internal/cmd/rulelint
// finds every one and matches it against the configuration.
package permission

// The effects of a policy entry.
const (
	Allow = "allow"
	Deny  = "deny"
)

// The subjects a policy belongs to.
const (
	KindRole = "role"
	KindUser = "user"
)
