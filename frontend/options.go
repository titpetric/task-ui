package frontend

import (
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// Options are what the pages read from.
type Options struct {
	Guard *access.Guard

	Users     *storage.UserStorage
	Projects  *storage.ProjectStorage
	Runs      *storage.RunStorage
	Artifacts *storage.ArtifactStorage
	Schedules *storage.ScheduleStorage
	Secrets   *secret.Store

	Runner    *runner.Manager
	Catalogue folder.Catalogue
	Git       *git.Git
	Keys      *sshkey.Store
	Env       *envfile.Store
	// Root is where the folders are, for the projects page.
	Root    string
	Version string
	// Now is the clock the pages read for the hours. Nil is time.Now.
	Now func() time.Time
}
