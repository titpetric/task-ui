package view

import (
	"github.com/titpetric/task-ui/sshkey"
)

// KeyFolder is a project's .ssh/ folder and the keys in it: the ones its
// runs get through an ssh agent, listed under the instance keys.
type KeyFolder struct {
	// Project is the folder ID of the project, a checkout at the top of
	// the root.
	Project string
	Dir     string
	Keys    []sshkey.Key
}
