package app

import (
	"context"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/hostinfo"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/keystore"
	"github.com/titpetric/task-ui/phpscript"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/retention"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshprobe"
	"github.com/titpetric/task-ui/tooling"
	"github.com/titpetric/task-ui/workflow"
)

// FrontBase is the root, where the ported front end answers: it is the front
// end the instance serves. Every link a page writes opens with it, and an
// empty FrontBase turns {base}/settings into /settings.
const FrontBase = ""

// GoBase mounts the Go pages, the front end being replaced.
// scripts/compare.sh fetches both to diff a ported page against its Go
// original; the cutover commit drops them along with this line.
//
// A Go page draws its links without a prefix, because the Go view layer has no
// notion of one: a link followed from one leaves for the ported page at the
// root. These pages serve a page-by-page diff, not browsing the old front end.
const GoBase = "/legacy"

// FrontHost answers phpfront.Host off the instance: the catalogue, the live
// runner, the checkouts, the permission file and the event bus. Every row the
// front end reads it reads itself.
//
// Rules returns the titles in file order and the path lists the file decides.
// Only the decision in code/Task/Access.php is ported out of permission/.
//
// The methods are grouped the way the interface is: the catalogue's in
// front_host_catalogue.go, the live runner's in front_host_run.go, a folder's
// pages in front_host_folder.go, the Inventory's in front_host_inventory.go,
// the settings' in front_host_settings.go, and the artifact tree, the
// permission file and the tooling report each in their own.
type FrontHost struct {
	Catalogue    folder.Catalogue
	Runner       *runner.Manager
	Git          *git.Git
	Events       *event.Bus
	Recordings   *recording.Store
	Integrations *integration.Store

	// Secrets is the store over the instance vault, asked what a variable
	// holds: the rows are PHP's and the key is not, so the value of one is
	// read here and nowhere else.
	Secrets *secret.Store

	// Workflows is the workflow runner, asked which workflows are running:
	// the rows say running of a run whose process is gone, and the two front
	// ends read the one memory so they mark the same workflows.
	Workflows *workflow.Runner

	// Artifacts is the tree the runs wrote, which the front end lists a
	// directory at a time, and Disk the meter both front ends measure it
	// with, so neither walks the tree a second time.
	Artifacts *artifactstore.Store
	Disk      *diskusage.Meter

	// Hosts is the Beszel hub, SSH who task-ui signs in to each host as,
	// and Facts what a host last told over that login.
	Hosts *beszel.Inventory
	SSH   *sshprobe.Prober
	Facts *hostinfo.Reader

	// Tools is what atkins has recorded for the tree, the prober both front
	// ends ask, so running atkins is not a cost the second one pays again.
	Tools *tooling.Prober

	// Keystore is the key folder, which answers the public half of a key no
	// row holds, and Config what the server runs with, the one the Instance
	// page of either front end draws.
	Keystore *keystore.Local
	Config   *view.Instance

	// Tracing is the tracer's options and Plugins the modules registered at
	// startup: the Telemetry and the Sidebar settings.
	Tracing *Telemetry
	Plugins *phpscript.Manager

	// Cleaner measures what the retention rules apply to, and Storage what
	// the instance keeps on disk, both the meters the Go pages read, so a
	// page drawn twice walks the disk once.
	Cleaner *retention.Cleaner
	Storage *diskusage.Inventory

	// Root is the workspace root, which a checkout's status is relative to.
	Root string

	// Now is the instance's clock, the one TASK_UI_NOW pins. An hours rule
	// is read at that moment, because view.Page.InHours reads it at the
	// page's own, and a binding on the process clock would answer open where
	// the Go page beside it answers closed. Nil is time.Now.
	Now func() time.Time

	// Dispatcher starts the runs a page requests and Runs answers which run
	// one turned into, the pair the plugins already use: a dispatched run is
	// an ordinary prompt either way, and the instance holds one of each.
	//
	// FrontRoot is where phpfront.Extract put the front end, the root a
	// dispatch's action file resolves against. It is the same path as
	// phpfront.Options.Root, kept here because the dispatcher takes the root
	// it reads in and the host is the one place that names this tree.
	Dispatcher *Dispatcher
	Runs       *ModuleRuns
	FrontRoot  string

	// Compose is the compose project task-ui itself runs in: compose up and
	// down of that project confirm first, since they stop the instance that
	// drew the page.
	Compose string

	// subs keeps what Folder.Sub read of the directories a Jobs table
	// probes, so one page view reads each directory once.
	subs frontSubs
}

// Whoami is the name the Guard resolved for this request.
func (h *FrontHost) Whoami(ctx context.Context) string {
	user := access.UserFrom(ctx)
	if user == nil {
		return ""
	}
	return user.Name
}
