// Package keystore keeps every key of the instance in one folder, flat:
// <name>.key, the private key, and <name>.pub, beside known_hosts. The
// name says the algorithm and the size, as deploy_rsa4096 or
// task_encryption_ed25519, and the database says what a key is: an
// encryption key or an SSH key, of the instance or a workspace, and
// whose runs get an SSH key. Nothing else reads a key file: git, the
// inventory's logins, the secrets and the runs go through the store,
// and a run gets its SSH keys in an ssh agent of its own, in memory.
//
// A workspace export bundles the SSH keys the workspace's runs get. No
// instance key and no encryption key leaves: the instance's encryption
// key encrypts every secret, and an export carries the secrets inside
// its passphrase.
//
// Migrate names the keys of the older layout so: the instance's
// encryption key as task_encryption_<algo>, a workspace's as
// workspace_<folder>_encryption_<algo> and its SSH keys as
// workspace_<folder>_<label>_<algo>, the folder's id_<type> as
// task_ssh_<algo>, which git pulls with, and id_<type>_<label> as
// <label>_<algo>. The backups come along, to decrypt what they
// encrypted.
package keystore
