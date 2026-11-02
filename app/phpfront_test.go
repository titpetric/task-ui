package app

import (
	"context"
	"slices"
	"testing"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/phpfront"
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

// Folders is the catalogue with each folder's jobs by name.
func TestFrontHost_Folders(t *testing.T) {
	t.Parallel()

	assert.Empty(t, (&FrontHost{}).Folders(context.Background()))

	host := &FrontHost{Catalogue: catalogue{{
		ID: "infra", Title: "Infra", Dir: "/src/infra",
		Jobs: []*folder.Job{{Name: "build"}, {Name: "test"}},
	}}}
	found := host.Folders(context.Background())
	assert.Len(t, found, 1)
	assert.Equal(t, "infra", found[0].ID)
	assert.Equal(t, "Infra", found[0].Title)
	assert.Equal(t, "/src/infra", found[0].Dir)
	assert.Equal(t, []string{"build", "test"}, found[0].Jobs)
}

// Running is empty without a runner, which is a test instance.
func TestFrontHost_Running(t *testing.T) {
	t.Parallel()

	assert.Empty(t, (&FrontHost{}).Running(context.Background()))
}

// Dirty is 0 without git or a root, so a page draws no count.
func TestFrontHost_Dirty(t *testing.T) {
	t.Parallel()

	host := &FrontHost{}
	assert.Equal(t, 0, host.Dirty(context.Background(), "/src/infra"))
	assert.Equal(t, 0, host.Dirty(context.Background(), ""))
}

// Rules carries the file in order, with the path lists it decides, so the
// decision is the only part of permission/ the front end ports.
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

	// The titles are the file's, so a page draws what config/permissions.yml says.
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

// The host satisfies the interface it is handed to, which keeps the binding
// list and the host in step.
var _ phpfront.Host = &FrontHost{}
