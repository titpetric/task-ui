package api

import (
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// Options are what the handlers work with.
type Options struct {
	Users     *storage.UserStorage
	Projects  *storage.ProjectStorage
	Runs      *storage.RunStorage
	Artifacts *storage.ArtifactStorage
	Schedules *storage.ScheduleStorage
	Secrets   *storage.SecretStorage

	Runner     *runner.Manager
	Catalogue  folder.Catalogue
	Git        *git.Git
	Keys       *sshkey.Store
	Env        *envfile.Store
	Recordings *recording.Store
	// Root is where projects are cloned to.
	Root string
}
