// Package access is who may do what: the passwords a sign in checks and
// the sessions that follow it. The users themselves are rows in storage
// and the rules on them are on model.User. The credentials, the email
// addresses and password hashes, live apart from the user row.
//
// A request signs in with a bearer token, the development auto login or
// the session cookie. A run token signs in a Claude run task-ui starts,
// as the user who started it, for as long as the run lasts: it needs no
// MCP agent user flag, and it names the note the run answers.
//
// Migrations owns the tables behind that: the passwords, the API tokens,
// the mail addresses and the sessions. No page queries them. The Guard
// stands in front of every mount, and minting or reading a session stays
// in Go, so the data model stays with the code that authenticates.
// migrate.Migrations merges these files with every other owner's, keyed by
// base name, the key the migrations table records: see internal/migrate.
package access
