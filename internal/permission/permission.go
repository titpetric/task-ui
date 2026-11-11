// Package permission decides what a user may do, from the rules of
// `config/permissions.yml` and the policies of the user and the user's
// role.
//
// A policy entry allows or denies a path, <section>/<rule> or an
// application permission; no entry inherits. A user's own entry decides
// first; a user without one inherits the role's, and task-ui refuses
// what neither sets. A <section>/all entry stands for every rule
// of the section, and where a subject's all and its entry for a rule
// disagree, the deny decides. The admin role allows everything, and
// the policies leave it as is.
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
