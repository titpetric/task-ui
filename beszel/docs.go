// Package beszel reads the systems of a Beszel hub, which watches the hosts, for
// task-ui's host inventory. The hub is PocketBase, and the client lists its
// systems collection with a token.
//
// Hosts writes the inventory for a run, in the order of the systems. Its
// logins hold user@host for each host task-ui signs in to, as sshprobe
// finds them; a host without one has an empty user.
package beszel
