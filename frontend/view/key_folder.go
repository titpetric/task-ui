package view

import (
	"github.com/titpetric/task-ui/sshkey"
)

// KeyFolder is a project's .ssh/ folder and the private keys in it. Runs
// do not load them; the keys page warns about them, to upload as instance
// keys and give by policy instead.
type KeyFolder struct {
	// Project is the folder ID of the project, a checkout at the top of
	// the root.
	Project string
	Dir     string
	Keys    []sshkey.Key
}
