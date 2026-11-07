package app

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/phpfront"
	"github.com/titpetric/task-ui/workflow"
)

// catalogue is the folders a test serves.
type catalogue []*folder.Folder

func (c catalogue) List() []*folder.Folder { return c }

func (c catalogue) Find(id string) *folder.Folder {
	for _, f := range c {
		if f.ID == id {
			return f
		}
	}
	return nil
}

// Whoami is the name the Guard resolved; nobody signed in is "".
func TestFrontHost_Whoami(t *testing.T) {
	t.Parallel()

	host := &FrontHost{}
	assert.Equal(t, "", host.Whoami(context.Background()))

	ctx := access.WithUser(context.Background(), &model.User{Name: "ops"})
	assert.Equal(t, "ops", host.Whoami(ctx))
}

// Folders is the catalogue with each folder's jobs, and a job holds what a
// gate reads: its kind and its own allow list.
func TestFrontHost_Folders(t *testing.T) {
	t.Parallel()

	assert.Empty(t, (&FrontHost{}).Folders(context.Background()))

	host := &FrontHost{Catalogue: catalogue{{
		ID: "infra", Title: "Infra", Dir: "/src/infra",
		Jobs: []*folder.Job{{Name: "build"}, {Name: "test", Allow: []string{"ann"}}},
	}}}
	found := host.Folders(context.Background())
	assert.Len(t, found, 1)
	assert.Equal(t, "infra", found[0].ID)
	assert.Equal(t, "Infra", found[0].Title)
	assert.Equal(t, "/src/infra", found[0].Dir)
	assert.Len(t, found[0].Jobs, 2)
	assert.Equal(t, "build", found[0].Jobs[0].Name)
	assert.Equal(t, "test", found[0].Jobs[1].Name)
	assert.Equal(t, []string{"ann"}, found[0].Jobs[1].Allow)
}

// The directories of a tree are read in one call, keyed by ID: the
// catalogue's list is indexed once, so a tree of fifty rows costs one scan.
// An ID the layout places nowhere is left out of the map; no key is set to the
// empty string.
func TestFrontHost_FolderDirs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	workspaces := []string{"malta"}

	assert.Empty(t, (&FrontHost{}).FolderDirs(ctx, []string{"infra"}, workspaces))

	host := &FrontHost{Catalogue: catalogue{{ID: "infra", Dir: "/src/infra"}, {ID: "bare"}}}
	found := host.FolderDirs(ctx, []string{"infra", "bare", "nowhere", ""}, workspaces)
	assert.Len(t, found, 1)
	assert.Equal(t, "/src/infra", found["infra"])

	// A folder the catalogue holds with no directory measured is no key
	// either, and neither is the empty ID the chrome's list may hold.
	_, ok := found["bare"]
	assert.False(t, ok, "a folder with no directory is left out")
	_, ok = found[""]
	assert.False(t, ok, "an empty ID is asked of nobody")

	// An ID no folder holds is the layout's, and the batch answers what
	// FolderDir answers for it: one rule, read once for the whole tree.
	rooted := &FrontHost{Root: "/root", Catalogue: host.Catalogue}
	batched := rooted.FolderDirs(ctx, []string{"infra", "malta/deploy"}, workspaces)
	assert.Equal(t, "/src/infra", batched["infra"])
	assert.Equal(t, rooted.FolderDir(ctx, "malta/deploy", workspaces), batched["malta/deploy"])
	assert.NotEmpty(t, batched["malta/deploy"])
}

// Running is empty without a runner, which is a test instance.
func TestFrontHost_Running(t *testing.T) {
	t.Parallel()

	assert.Empty(t, (&FrontHost{}).Running(context.Background()))
}

// Flows is empty without a workflow runner, and empty from a runner that
// holds nothing: no workflow is marked running off the rows either way.
func TestFrontHost_Flows(t *testing.T) {
	t.Parallel()

	assert.Empty(t, (&FrontHost{}).Flows(context.Background()))
	assert.Empty(t, (&FrontHost{Workflows: &workflow.Runner{}}).Flows(context.Background()))
}

// Dirty is 0 without git or a root, so a page draws no count.
func TestFrontHost_Dirty(t *testing.T) {
	t.Parallel()

	host := &FrontHost{}
	assert.Equal(t, 0, host.Dirty(context.Background(), "/src/infra"))
	assert.Equal(t, 0, host.Dirty(context.Background(), ""))
}

// Rules returns the file in order, with the path lists it decides. Only that
// decision is ported out of permission/.
func TestFrontHost_Rules(t *testing.T) {
	t.Parallel()

	rules := (&FrontHost{}).Rules(context.Background())
	assert.NotEmpty(t, rules.Sections)
	assert.NotEmpty(t, rules.Rules)
	assert.NotEmpty(t, rules.Permissions)

	assert.True(t, slices.Contains(rules.Paths, "user/all"), "paths hold the all rules")
	assert.True(t, slices.Contains(rules.Paths, "user/edit"))
	assert.True(t, slices.Contains(rules.Checks, "user/edit"))
	assert.False(t, slices.Contains(rules.Checks, "user/all"), "a check never names an all rule")
	assert.True(t, slices.Contains(rules.Checks, "run"), "an application permission is a check")

	// The titles are the file's, so a page draws what `config/permissions.yml` says.
	assert.True(t, slices.ContainsFunc(rules.Sections, func(r phpfront.Rule) bool {
		return r.Key == "note" && r.Title == "Notes"
	}))
}

// Publish names the event off the payload, so a label follows the same stream
// the Go pages publish on.
func TestFrontHost_Publish(t *testing.T) {
	t.Parallel()

	assert.NoError(t, (&FrontHost{}).Publish(context.Background(), "notes", nil), "no bus is not a failure")

	bus := event.NewBus()
	events, done := bus.Subscribe("notes")
	defer done()

	host := &FrontHost{Events: bus}
	assert.NoError(t, host.Publish(context.Background(), "notes", map[string]any{"event": "note", "id": "01hz"}))

	got := <-events
	assert.Equal(t, "note", got.Name)
}

// An instance with no Beszel integration answers the Inventory page's unset
// state, which the page draws in place of an empty list. A host nobody
// monitors answers not found.
func TestFrontHost_Inventory(t *testing.T) {
	t.Parallel()

	host := &FrontHost{}
	assert.True(t, host.Inventory(context.Background(), true).Off)

	facts := host.HostFacts(context.Background(), "web1", true)
	assert.False(t, facts.Found)
	assert.Equal(t, beszel.ErrOff.Error(), facts.Error)
}

// frontUptime words how long a host has run: days past two, else hours past
// two, else minutes.
func TestFrontUptime(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "3 days", frontUptime(72*time.Hour))
	assert.Equal(t, "2 days", frontUptime(48*time.Hour))
	assert.Equal(t, "47 hours", frontUptime(47*time.Hour))
	assert.Equal(t, "2 hours", frontUptime(2*time.Hour))
	assert.Equal(t, "119 minutes", frontUptime(119*time.Minute))
	assert.Equal(t, "0 minutes", frontUptime(0))
}

// The host satisfies the interface it is handed to, which keeps the binding
// list and the host in step.
var _ phpfront.Host = &FrontHost{}
