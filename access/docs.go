// Package access is who may do what: the passwords users sign in with
// and the sessions they carry. The users themselves are rows in storage,
// the rules on them are on model.User, and what they sign in with, the
// email addresses and password hashes, lives apart from the user row.
//
// A request signs in with a bearer token, the development auto login or
// the session cookie. A run token signs in a Claude run task-ui starts,
// as the user who started it, for as long as the run lasts: it needs no
// MCP agent user flag, and it names the note the run answers.
package access
