package phpfront

import (
	"reflect"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
)

// The keys the settings pages index by. The helpers are convert_test.go's.
//
// Three of these converters are ordered rather than keyed by position, and the
// order is the fields' and not a Go map's: a settings page that draws the same
// service twice draws it the same way, which a map iteration would not.

// A key crosses with its fingerprint and what it is for, and never with the
// private half: the keystore holds that and nothing reads it over the
// boundary.
func TestKeystore(t *testing.T) {
	t.Parallel()

	out := keystore(Keystore{
		Dir:      "/data/keys",
		Instance: "task-ui",
		Keys: []Key{{
			Name: "deploy", Purpose: "git", Scope: "workspace", Algo: "ed25519",
			Bits: 256, Size: 464, Fingerprint: "SHA256:abc", Comment: "deploy@task-ui",
			PublicKey: "ssh-ed25519 AAAA", CreatedAt: "2026-11-02 10:32:12.414 +0000 UTC",
			Roles: []string{"ops"}, Users: []string{"ann"},
			Git: true, Workspace: "malta", Backup: "deploy.bak", BackupFingerprint: "SHA256:def",
		}},
	})

	assert.Equal(t, "/data/keys", text(t, out, "dir"))
	assert.Equal(t, "task-ui", text(t, out, "instance"))

	key := at(t, list(t, out, "keys"), 0)
	assert.Equal(t, "deploy", text(t, key, "name"))
	assert.Equal(t, "ed25519", text(t, key, "algo"))
	assert.Equal(t, "SHA256:abc", text(t, key, "fingerprint"))
	assert.Equal(t, "ssh-ed25519 AAAA", text(t, key, "public_key"))
	assert.True(t, flag(t, key, "git"))
	assert.False(t, flag(t, key, "encrypts"))
	assert.Equal(t, "SHA256:def", text(t, key, "backup_fingerprint"))
	assert.Equal(t, 1, list(t, key, "roles").Len())
	assert.Equal(t, 1, list(t, key, "users").Len())

	_, ok := key.Get("private_key")
	assert.False(t, ok, "the private half is not a key of the answer at all")

	assert.Equal(t, 0, list(t, keystore(Keystore{}), "keys").Len())
}

// The .ssh folder of every project that holds a key, by project, with the key
// names alone: the page lists what is there and opens one on a click.
func TestKeyFolders(t *testing.T) {
	t.Parallel()

	out := keyFolders([]KeyFolder{{
		Project: "malta/infra", Dir: "/src/infra/.ssh", Keys: []string{"id_ed25519", "deploy"},
	}})
	assert.Equal(t, 1, out.Len())

	row := at(t, out, 0)
	assert.Equal(t, "malta/infra", text(t, row, "project"))
	assert.Equal(t, "/src/infra/.ssh", text(t, row, "dir"))
	assert.Equal(t, 2, list(t, row, "keys").Len())

	assert.Equal(t, 0, keyFolders(nil).Len())
}

// Every field of the struct is a key, checked by reflection rather than by
// name, the same guarantee TestTelemetry gives: the Instance page is the one
// that prints what the server runs with, and a field it cannot reach is a row
// drawing an empty box over a value the instance holds.
func TestInstance(t *testing.T) {
	t.Parallel()

	out := instance(Instance{
		URL: "https://task.cdn.si", Listen: ":3000", Root: "/src",
		DataDir: "/data", SSHDir: "/data/keys", TimeZone: "Europe/Ljubljana",
		Running: true,
		Backlog: Backlog{Parallel: 4, Going: 2, Waiting: 1, Auto: 8, Configured: 4, Length: 3},
		Env:     []EnvName{{Name: "CLAUDE_CODE_OAUTH_TOKEN", Sensitive: true}, {Name: "PATH"}},
		Storage: Storage{
			Rows:  []StorageRow{{Name: "artifacts", Where: "/data/artifacts", Size: 1 << 20, Files: 128}},
			Total: StorageRow{Name: "total", Size: 1 << 21, Files: 256},
		},
		Measured: true,
	})

	for i := range reflect.TypeFor[Instance]().NumField() {
		key := reflect.TypeFor[Instance]().Field(i).Tag.Get("json")
		_, ok := out.Get(key)
		assert.True(t, ok, key)
	}

	assert.Equal(t, "https://task.cdn.si", text(t, out, "url"))
	assert.Equal(t, "Europe/Ljubljana", text(t, out, "time_zone"))
	assert.True(t, flag(t, out, "running"))

	// The backlog is one nested array, because the page draws it as one
	// line: six fields of arithmetic the template does not do. auto is what
	// the machine would allow and configured what the instance was told,
	// and the line says which of them is in force.
	backlog := list(t, out, "backlog")
	assert.Equal(t, 4, value[int](t, backlog, "parallel"))
	assert.Equal(t, 1, value[int](t, backlog, "waiting"))
	assert.Equal(t, 8, value[int](t, backlog, "auto"))
	assert.Equal(t, 4, value[int](t, backlog, "configured"))

	// A sensitive variable crosses by name with no value: the page lists
	// what is set and never what it is set to.
	env := at(t, list(t, out, "env"), 0)
	assert.Equal(t, "CLAUDE_CODE_OAUTH_TOKEN", text(t, env, "name"))
	assert.True(t, flag(t, env, "sensitive"))
	_, ok := env.Get("value")
	assert.False(t, ok)

	storage := list(t, out, "storage")
	assert.Equal(t, int64(1<<20), value[int64](t, at(t, list(t, storage, "rows"), 0), "size"))
	assert.Equal(t, int64(256), value[int64](t, list(t, storage, "total"), "files"))
}

// Every field of the struct is a key, checked by reflection rather than by
// name: a field added to host_settings.go and forgotten here is a form drawing
// an empty box over a value the instance holds, which is what happened to
// debug_span_min_duration.
func TestTelemetry(t *testing.T) {
	t.Parallel()

	out := telemetry(Telemetry{Wired: true, ServiceName: "task-ui", DebugSpanMinDuration: "1ms"})
	for i := range reflect.TypeFor[Telemetry]().NumField() {
		key := reflect.TypeFor[Telemetry]().Field(i).Tag.Get("json")
		_, ok := out.Get(key)
		assert.True(t, ok, key)
	}

	assert.True(t, flag(t, out, "wired"))
	assert.Equal(t, "task-ui", text(t, out, "service_name"))
	assert.Equal(t, "1ms", text(t, out, "debug_span_min_duration"))
}

// The retention areas cross keyed by kind, because a row of the table reads
// its own area and does not search the list for it.
func TestUsage(t *testing.T) {
	t.Parallel()

	out := usage(Retention{Areas: []AreaUsage{
		{Kind: "artifacts", Count: 128, Bytes: 1 << 20, Oldest: "2026-10-04 09:00:00 +0000 UTC"},
		{Kind: "recordings", Count: 12, Bytes: 1 << 16},
	}})

	assert.False(t, flag(t, out, "off"))
	areas := list(t, out, "areas")
	assert.Equal(t, 2, areas.Len())

	artifacts := list(t, areas, "artifacts")
	assert.Equal(t, 128, value[int](t, artifacts, "count"))
	assert.Equal(t, int64(1<<20), value[int64](t, artifacts, "bytes"))
	assert.Equal(t, "2026-10-04 09:00:00 +0000 UTC", text(t, artifacts, "oldest"))

	_, ok := areas.Get("notes")
	assert.False(t, ok, "an area nothing holds is no row")

	off := usage(Retention{Off: true})
	assert.True(t, flag(t, off, "off"))
	assert.Equal(t, 0, list(t, off, "areas").Len())
}

// A service crosses with its registry entry, the state of its settings, and no
// secret in it: set says the token has a value and nothing says what it is.
//
// The three maps are written in the order the fields are declared, which is
// what makes the answer the same twice over for the same service.
func TestIntegrations(t *testing.T) {
	t.Parallel()

	service := Integration{
		ID: "claude", Title: "Claude Code", Description: "The token Claude Code signs in with.",
		Credential: "OAuth token", Steps: []string{"Run claude setup-token."},
		Fields: []IntegrationField{
			{Name: "token", Label: "OAuth token", Secret: true, Env: "CLAUDE_CODE_OAUTH_TOKEN"},
			{Name: "model", Label: "Model", Options: []string{"opus", "sonnet"}, Default: "opus"},
			{Name: "live", Label: "Live", Toggle: true, DefaultOn: true},
			{Name: "who", Label: "Who", Optional: true, Who: true},
		},
		Configured: true,
		Values:     map[string]string{"who": "admin, ops", "model": "opus"},
		Set:        map[string]bool{"token": true},
		Checked:    []string{"admin", "ops"},
		Enabled:    map[string]bool{"live": true},
	}

	out := at(t, integrations([]Integration{service}), 0)
	assert.Equal(t, "claude", text(t, out, "id"))
	assert.Equal(t, "OAuth token", text(t, out, "credential"))
	assert.Equal(t, 1, list(t, out, "steps").Len())
	assert.True(t, flag(t, out, "configured"))
	assert.Equal(t, 2, list(t, out, "checked").Len())

	field := at(t, list(t, out, "fields"), 0)
	assert.True(t, flag(t, field, "secret"))
	assert.Equal(t, "CLAUDE_CODE_OAUTH_TOKEN", text(t, field, "env"))
	assert.Equal(t, 2, list(t, at(t, list(t, out, "fields"), 1), "options").Len())
	assert.True(t, flag(t, at(t, list(t, out, "fields"), 2), "default_on"))

	// set carries a flag and never the value, so a secret the instance holds
	// crosses as the fact that it is held.
	assert.True(t, flag(t, list(t, out, "set"), "token"))
	_, ok := list(t, out, "values").Get("token")
	assert.False(t, ok)

	assert.Equal(t, "opus", text(t, list(t, out, "values"), "model"))
	assert.True(t, flag(t, list(t, out, "enabled"), "live"))

	assert.Equal(t, 0, integrations(nil).Len())
}

// The two orderings, stated directly: what a service's maps cross as is the
// declaration order of its fields, and a key the map leaves out is left out
// rather than written empty.
func TestSettingsOrder(t *testing.T) {
	t.Parallel()

	fields := []IntegrationField{{Name: "token"}, {Name: "model"}, {Name: "who"}}

	var keys []string
	settings(fields, map[string]string{"who": "ops", "token": "t", "model": "opus"}).
		Range(func(key, _ any) bool {
			keys = append(keys, key.(string))
			return true
		})
	assert.Equal(t, []string{"token", "model", "who"}, keys, "the fields' order, not the map's")

	// A field the map says nothing about is no key, which is how a form
	// tells an unset setting from one set to the empty string.
	out := settings(fields, map[string]string{"model": "opus"})
	assert.Equal(t, 1, out.Len())
	_, ok := out.Get("token")
	assert.False(t, ok)

	var marks []string
	flags(fields, map[string]bool{"who": false, "token": true}).
		Range(func(key, _ any) bool {
			marks = append(marks, key.(string))
			return true
		})
	assert.Equal(t, []string{"token", "who"}, marks)
	assert.Equal(t, 0, flags(fields, nil).Len())
}

// The vault crosses keyed by row ID with what the walk of it said, which is
// the shape the page draws: the table and the error above it.
func TestVariables(t *testing.T) {
	t.Parallel()

	out := variables(Variables{Values: map[string]string{"TOKEN": "", "REGION": "eu"}})
	assert.Equal(t, "", text(t, out, "error"))

	values := list(t, out, "values")
	assert.Equal(t, 2, values.Len())
	assert.Equal(t, "eu", text(t, values, "REGION"))
	assert.Equal(t, "", text(t, values, "TOKEN"), "a secret is a row with no value, not a row left out")

	failed := variables(Variables{Error: "no encryption key"})
	assert.Equal(t, "no encryption key", text(t, failed, "error"))
	assert.Equal(t, 0, list(t, failed, "values").Len())

	// One opened variable is its content and what refusing to open it said,
	// and the form draws whichever is set.
	assert.Equal(t, "eu", text(t, variable(Variable{Content: "eu"}), "content"))
	assert.Equal(t, "sealed", text(t, variable(Variable{Error: "sealed"}), "error"))
}

// What the integrations add to a run, in the order the page lists them: a
// credential crosses by name, marked redacted, with no value.
func TestEnvVars(t *testing.T) {
	t.Parallel()

	out := envVars([]EnvVar{
		{Name: "CLAUDE_CODE_OAUTH_TOKEN", Redacted: true},
		{Name: "CLAUDE_MODEL", Value: "opus"},
	})
	assert.Equal(t, 2, out.Len())

	assert.True(t, flag(t, at(t, out, 0), "redacted"))
	assert.Equal(t, "", text(t, at(t, out, 0), "value"))
	assert.Equal(t, "opus", text(t, at(t, out, 1), "value"))
	assert.False(t, flag(t, at(t, out, 1), "redacted"))

	assert.Equal(t, 0, envVars(nil).Len())
}

// The modules the instance serves, and the project directories of the
// workspaces: both are lists the sidebar and the Projects table walk.
func TestPluginsAndProjects(t *testing.T) {
	t.Parallel()

	module := at(t, plugins([]Plugin{{
		Name: "refinement", Title: "Refinement", Description: "The threads a run opened.", Icon: "note",
	}}), 0)
	assert.Equal(t, "refinement", text(t, module, "name"))
	assert.Equal(t, "note", text(t, module, "icon"))

	project := at(t, projects([]Project{{
		ID: "malta/infra", Dir: "/src/infra", Sources: []string{"Taskfile.yml"}, Jobs: 12,
		Checkout: true, Remote: "git@github.com:malta/infra", Branch: "main",
		Commit: "9bbee11", Subject: "feat: the first page", When: "2026-11-02 10:32:12.414 +0000 UTC",
	}}), 0)
	assert.Equal(t, "malta/infra", text(t, project, "id"))
	assert.Equal(t, 1, list(t, project, "sources").Len())
	assert.Equal(t, 12, value[int](t, project, "jobs"))
	assert.True(t, flag(t, project, "checkout"))
	assert.Equal(t, "9bbee11", text(t, project, "commit"))

	// A project that detected nothing says why, in the one key the row
	// draws in place of the job count.
	assert.Equal(t, "not a checkout", text(t, at(t, projects([]Project{{Error: "not a checkout"}}), 0), "error"))

	assert.Equal(t, 0, plugins(nil).Len())
	assert.Equal(t, 0, projects(nil).Len())
}
