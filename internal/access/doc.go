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
// The tables behind that are the user module's, under
// phpfront/app/modules/user/schema: every user_ table migrates with the
// pages that draw a user. This package writes them and owns none.
package access
