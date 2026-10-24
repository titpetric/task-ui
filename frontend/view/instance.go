package view

import (
	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/runner"
)

// Instance is what the server runs with, as the Instance step of the
// settings shows it: the configuration from the environment and the
// command line, less anything secret. A path shows; a token never does.
type Instance struct {
	// URL is TASK_UI_URL, empty when the links derive from each request.
	URL      string
	Listen   string
	Root     string
	DataDir  string
	SSHDir   string
	Htpasswd string
	// HostData is the data directory on the host, empty when the same.
	HostData string
	Discover bool
	// AutoLogin is the user every request signs in as, if any.
	AutoLogin string
	// TimeZone is the zone of the server's clock, as the pages and the
	// hours read it.
	TimeZone string
	// SessionTTL, Keep and Grace are TASK_UI_SESSION_TTL, TASK_UI_KEEP
	// and TASK_UI_GRACE, as durations read.
	SessionTTL string
	Keep       string
	Grace      string
	// ArtifactHistory is TASK_UI_ARTIFACT_HISTORY, 0 for every version.
	ArtifactHistory int
	// ComposeProject is the compose project task-ui runs in, if any.
	ComposeProject string
	// Beszel says the Beszel integration has a hub and a token, which
	// the Inventory reads.
	Beszel bool
	// Backlog is how the slots and the line for them stand, nil without
	// a runner.
	Backlog *runner.BacklogState
	// ParentEnv is the parent environment mode in effect, empty without
	// a runner; Env the names of the server's variables, never their
	// values, for an admin alone.
	ParentEnv string
	Env       []runner.EnvName
	// Storage is what the instance keeps on disk, by item, for an admin
	// alone; nil leaves the Storage table out.
	Storage *diskusage.Table
}

// EnvBase returns runner.BaseEnv, the variables None passes.
func (i *Instance) EnvBase() []string {
	return runner.BaseEnv
}

// EnvSensitive returns runner.SensitiveWords.
func (i *Instance) EnvSensitive() []string {
	return runner.SensitiveWords
}
