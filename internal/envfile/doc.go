// Package envfile keeps the dotenv files of older instances: a workspace
// bundle carries them, and an import writes them into the folder. New
// variables belong to the secret store, a secret or an env entry, and
// runs read them from there. It discovers nothing: what a folder's
// sources read, atkins and the job tools read themselves.
package envfile
