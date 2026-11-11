// Package folder is the catalogue: a folder is a working directory and the
// targets found in it, and a job is one of those targets with an identity
// inside the folder.
//
// A folder with a pipeline file lists that pipeline's targets only: the
// Makefile, package scripts and scripts folder beside it are the
// pipeline's business and list no jobs of their own. A configuration widens or
// narrows a folder by the detector kinds it names. Load and Skills order a
// folder's jobs with DefaultFirst, so the default job comes first everywhere.
//
// Folder IDs nest, so Resolve takes the longest folder that starts a path:
// "services/api/deploy" is the job deploy of services/api when that is a
// folder, and the folder services/api/deploy when that is one.
//
// Every folder has built-in jobs beside the ones its sources declare:
//
//   - @shell, `bash -l`, and @claude, Claude Code in a terminal, are for
//     admins only and need someone at the keyboard.
//   - @claude-note runs as the root's, _/@claude-note: the board starts it
//     on a note, as the agent's user, and no run, schedule or workflow step
//     takes it by ID.
//   - @claude-prompt starts from the prompt routes with its thread, and no
//     run, schedule or workflow step takes it by ID.
//
// The root, _, adds two more. Its cleanup job removes the recordings, the
// artifact versions and the secret reads the Data retention settings let
// go. Its @host:<name> job runs its args on a host over ssh, and @host:*
// on every host with an ssh login, a run a host. The host job does not
// check that the inventory lists the host; the run's starter does, and
// fills in the ssh client as the command, so the job alone has none. The
// run names the host as its target and the command as its args, and the
// hosts permission says who runs it. Unlike @shell it needs no one at the
// keyboard, so a schedule or a workflow step takes it.
package folder
