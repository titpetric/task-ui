package hostrun_test

import (
	"context"
	"testing"
	"time"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/inventory/hostrun"
	"github.com/titpetric/task-ui/integration/beszel"
	"github.com/titpetric/task-ui/integration/ssh/sshprobe"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/internal/folder"
)

// root is a catalogue of the root folder alone.
type root struct{ f *folder.Folder }

func (r root) Find(id string) *folder.Folder {
	if id == model.Default {
		return r.f
	}
	return nil
}

// off is a hub the integration is not set up for, which answers ErrOff.
func off() *beszel.Inventory {
	return &beszel.Inventory{Settings: func(context.Context) (string, string, error) { return "", "", nil }}
}

// signsIn is an ssh probe that signs in to every address, or to none.
func signsIn(answers bool) *sshprobe.Prober {
	command := "false"
	if answers {
		command = "true"
	}
	return &sshprobe.Prober{Users: []string{"debian"}, Command: command, Timeout: time.Second, TTL: time.Minute}
}

// A starter needs a hub and a prober to resolve anything at all.
func TestStarter(t *testing.T) {
	t.Parallel()
	_, _, err := (&hostrun.Starter{}).Resolve(context.Background(), nil)
	assert.ErrorIs(t, err, beszel.ErrOff)
	_, _, err = (&hostrun.Starter{Hosts: off()}).Resolve(context.Background(), nil)
	assert.ErrorIs(t, err, beszel.ErrOff)
}

// A host of the inventory is a target while the hub answers nothing, and an
// address the inventory does not hold is no host.
func TestStarter_Resolve(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	held := func(context.Context) map[string]string {
		return map[string]string{"10.0.0.2": "box", "10.0.0.3": "web"}
	}
	starter := &hostrun.Starter{Hosts: off(), SSH: signsIn(true), Held: held, Wait: 2 * time.Second}

	hosts, logins, err := starter.Resolve(ctx, []string{"10.0.0.2"})
	assert.NoError(t, err)
	assert.Len(t, hosts, 1)
	assert.Equal(t, "debian@10.0.0.2", logins["10.0.0.2"])

	// Every host, which is every held address an ssh login reaches.
	hosts, _, err = starter.Resolve(ctx, []string{"*"})
	assert.NoError(t, err)
	assert.Len(t, hosts, 2)

	_, _, err = starter.Resolve(ctx, []string{"10.0.0.9"})
	assert.ErrorIs(t, err, hostrun.ErrUnknownHost)

	// A hub that answers nothing and an inventory with no row leave the
	// hub's error, because there is no list to run over.
	_, _, err = (&hostrun.Starter{Hosts: off(), SSH: signsIn(true)}).Resolve(ctx, nil)
	assert.ErrorIs(t, err, beszel.ErrOff)

	// A host nothing signs in to is no host to run a command on.
	closed := &hostrun.Starter{Hosts: off(), SSH: signsIn(false), Held: held, Wait: 2 * time.Second}
	_, _, err = closed.Resolve(ctx, nil)
	assert.ErrorIs(t, err, hostrun.ErrNoHosts)
}

// A start needs a command and the root's host jobs.
func TestStarter_Start(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	starter := &hostrun.Starter{
		Catalogue: root{f: &folder.Folder{ID: model.Default}},
		Hosts:     off(),
		SSH:       signsIn(true),
		Held:      func(context.Context) map[string]string { return map[string]string{"10.0.0.2": "box"} },
		Wait:      2 * time.Second,
	}
	user := &model.User{ID: "1", Name: "root", Admin: true, Role: model.RoleAdmin}

	_, _, err := starter.Start(ctx, user, hostrun.Request{Command: "  "})
	assert.ErrorIs(t, err, hostrun.ErrNoCommand)

	// A target the inventory does not hold is refused before a run starts.
	_, _, err = starter.Start(ctx, user, hostrun.Request{Command: "uptime", Targets: []string{"10.0.0.9"}})
	assert.ErrorIs(t, err, hostrun.ErrUnknownHost)
}
