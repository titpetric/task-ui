package model

// AuthKind is what a stored password is for. The values mirror the CHECK
// constraint in schema/user_auth_password.up.sql.
type AuthKind = string

// The kinds of password a user signs in with.
const (
	// AuthPassword is the user's password, from the password file.
	AuthPassword AuthKind = "password"
	// AuthBackup is a backup password: it signs in once, then counts as used.
	AuthBackup AuthKind = "backup"
)
