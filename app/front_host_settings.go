package app

import (
	"context"
	"path/filepath"
	"slices"
	"strings"

	"github.com/titpetric/oida"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/integration/ssh/diskusage"
	"github.com/titpetric/task-ui/integration/ssh/keystore"
	"github.com/titpetric/task-ui/integration/ssh/sshkey"
	"github.com/titpetric/task-ui/internal/access"
	"github.com/titpetric/task-ui/internal/folder"
	"github.com/titpetric/task-ui/internal/retention"
	"github.com/titpetric/task-ui/phpfront"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/storage"
)

// Keys is the key folder: where it is, the name of the instance's encryption
// key and every key in it.
//
// The private half of a key never leaves the keystore and the public half is a
// file beside it, which the key_info row does not hold, so a page that prints
// an authorized_keys line reads it here and nowhere else.
func (h *FrontHost) Keys(ctx context.Context) phpfront.Keystore {
	if h.Keystore == nil {
		return phpfront.Keystore{}
	}
	out := phpfront.Keystore{Dir: h.Keystore.Dir}
	if key := h.Keystore.EncryptKey(ctx, storage.ScopeInstance); key != nil {
		out.Instance = key.Name
	}
	keys, err := h.Keystore.List(ctx)
	if err != nil {
		return out
	}
	for _, key := range keys {
		out.Keys = append(out.Keys, frontKey(key))
	}
	return out
}

// KeyFolders is the .ssh folder of every checkout at the top of the root that
// holds a private key of its own, which frontend/key.go:143 walks: runs are
// given none of them, and the keys page warns about each.
func (h *FrontHost) KeyFolders(context.Context) []phpfront.KeyFolder {
	var out []phpfront.KeyFolder
	for _, f := range h.Catalogue.List() {
		if f.Depth() != 0 || h.Git == nil || !h.Git.IsRepo(f.Dir) {
			continue
		}
		dir := filepath.Join(f.Dir, ".ssh")
		store := &sshkey.Store{Dir: dir}
		keys := store.List()
		if len(keys) == 0 {
			continue
		}
		names := make([]string, 0, len(keys))
		for _, key := range keys {
			names = append(names, key.Name)
		}
		out = append(out, phpfront.KeyFolder{Project: f.ID, Dir: dir, Keys: names})
	}
	return out
}

// frontKey is one key of the folder in the depth a keys table draws, which is
// the row and the public half together.
func frontKey(key keystore.Key) phpfront.Key {
	out := phpfront.Key{
		Name: key.Name, Purpose: key.Purpose, Scope: key.Scope,
		Algo: key.Algo, Bits: key.Bits, Size: key.Size(),
		Fingerprint: key.Fingerprint, Comment: key.Comment, PublicKey: key.PublicKey,
		CreatedAt: frontTime(key.CreatedAt),
		Roles:     key.Roles, Users: key.Users,
		Git: key.Git, Encrypts: key.Encrypts(), Workspace: key.Workspace(),
		Backup: key.Backup, BackupFingerprint: key.BackupFingerprint,
	}
	if key.RotatedAt != nil {
		out.RotatedAt = frontTime(*key.RotatedAt)
	}
	return out
}

// Spaces is the space each directory takes, by the directory, through the
// meter both front ends read.
//
// The measure is of the directory and not of a checkout: a project under src
// is a plain folder as often as it is a clone, and the Projects settings draw
// a size for both.
func (h *FrontHost) Spaces(_ context.Context, dirs []string) map[string]phpfront.Space {
	out := make(map[string]phpfront.Space, len(dirs))
	if h.Disk == nil {
		return out
	}
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		usage := h.Disk.Measure(dir)
		out[dir] = phpfront.Space{
			Measured: true, Files: usage.Files, History: usage.History, Ignored: usage.Ignored,
		}
	}
	return out
}

// Instance is what the server runs with, as the Instance step of the settings
// draws it: the configuration, the runner's line for a slot, the parent
// environment in effect and whether the inventory has a hub.
//
// admin is the reader's flag, which PHP decides. The names of the server's own
// variables and the Storage table are an admin's: a name reads as a credential
// and a measure walks the disk, so neither is answered to anybody else.
func (h *FrontHost) Instance(ctx context.Context, admin bool) phpfront.Instance {
	var out phpfront.Instance
	if h.Config != nil {
		// The configuration half is read once at start and copied here; what
		// follows is the half that moves between one request and the next.
		out = *h.Config
	}
	out.EnvBase, out.EnvSensitive = runner.BaseEnv, runner.SensitiveWords
	if h.Integrations != nil {
		values, _ := h.Integrations.Values(ctx, integration.Beszel)
		out.Beszel = values["url"] != "" && values["token"] != ""
	}
	if h.Runner != nil {
		backlog := h.Runner.Backlog(ctx)
		out.Running = true
		out.Backlog = phpfront.Backlog{
			Parallel: backlog.Parallel, Going: backlog.Going, Waiting: backlog.Waiting,
			Auto: backlog.Auto, Configured: backlog.Configured, Length: backlog.Length,
		}
		out.ParentEnv = h.Runner.ParentEnv(ctx)
		if admin {
			for _, name := range h.Runner.EnvNames() {
				out.Env = append(out.Env, phpfront.EnvName{Name: name.Name, Sensitive: name.Sensitive, None: name.None, Auto: name.Auto})
			}
		}
	}
	if h.Storage != nil && admin {
		out.Measured = true
		out.Storage = frontStorage(h.Storage.Table())
	}
	return out
}

// frontStorage is the Storage table of the Instance page, through the meter
// both front ends read, so opening the page twice walks the disk once.
func frontStorage(table diskusage.Table) phpfront.Storage {
	out := phpfront.Storage{Total: frontStorageRow(table.Total), Pending: table.Pending}
	for _, row := range table.Rows {
		out.Rows = append(out.Rows, frontStorageRow(row))
	}
	return out
}

// frontStorageRow is one item of the Storage table.
func frontStorageRow(row diskusage.Row) phpfront.StorageRow {
	return phpfront.StorageRow{Name: row.Name, Where: row.Where, Size: row.Size, Files: row.Files, Pending: row.Pending}
}

// Services is every integration as the Integrations settings draw it, in
// registry order: each service's description beside its current settings.
//
// No secret value leaves this call: a secret answers as set or not.
func (h *FrontHost) Services(ctx context.Context) []phpfront.Integration {
	kinds := integration.Kinds()
	out := make([]phpfront.Integration, 0, len(kinds))
	for _, kind := range kinds {
		var values map[string]string
		var failed string
		if h.Integrations != nil {
			read, err := h.Integrations.Values(ctx, kind.ID)
			if err != nil {
				failed = err.Error()
			}
			values = read
		}
		shown := phpfront.Integration{
			ID: kind.ID, Title: kind.Title, Description: kind.Description,
			Credential: kind.Credential(), Steps: kind.Steps,
			Fields:     frontFields(kind.Fields),
			Configured: integration.Configured(kind, values),
			Values:     map[string]string{}, Set: map[string]bool{}, Enabled: map[string]bool{},
			Checked: frontChecked(kind, values), Error: failed,
		}
		for _, field := range kind.Fields {
			if field.Toggle {
				shown.Enabled[field.Name] = integration.Enabled(field, values)
			}
			if field.Secret {
				shown.Set[field.Name] = values[field.Name] != ""
				continue
			}
			shown.Values[field.Name] = values[field.Name]
		}
		out = append(out, shown)
	}
	return out
}

// Variables is what the Secrets page prints of the vault: the value of every
// plain env variable, by row ID, with the result of the walk that read them.
//
// It is the list the Go page draws, opened once. A secret is left out, as
// neither page prints one. A variable the walk stopped at keeps the armour its
// row holds, and the Go page prints that armour.
func (h *FrontHost) Variables(ctx context.Context) phpfront.Variables {
	out := phpfront.Variables{Values: map[string]string{}}
	if h.Secrets == nil {
		return out
	}
	secrets, err := h.Secrets.List(ctx)
	if err != nil {
		out.Error = err.Error()
	}
	for _, found := range secrets {
		if found.IsSecret == 0 {
			out.Values[found.ID] = found.Content
		}
	}
	return out
}

// Variable is one variable opened on its own, for the form that edits it.
//
// The form opens the row it draws, as frontend/secret.go does: a variable the
// list could not reach past opens here all the same, and one the vault will
// not open is the shared 404.
func (h *FrontHost) Variable(ctx context.Context, id string) phpfront.Variable {
	if h.Secrets == nil {
		return phpfront.Variable{}
	}
	found, err := h.Secrets.ByID(ctx, id)
	if err != nil {
		return phpfront.Variable{Error: err.Error()}
	}
	if found.IsSecret != 0 {
		return phpfront.Variable{}
	}
	return phpfront.Variable{Content: found.Content}
}

// IntegrationEnv is what the integrations add to a run of the reader, name and
// value in name order, a credential's value withheld.
//
// The reader is the one the Guard resolved for this request, as the Go page
// reads the environment of whoever opened it: the services a Who setting names
// reach some readers and not others.
func (h *FrontHost) IntegrationEnv(ctx context.Context) []phpfront.EnvVar {
	user := access.UserFrom(ctx)
	if h.Integrations == nil || user == nil {
		return nil
	}
	vars, err := h.Integrations.Env(ctx, user.Name, user.EffectiveRole())
	if err != nil {
		return nil
	}
	out := make([]phpfront.EnvVar, 0, len(vars))
	for _, name := range model.SortedKeys(vars) {
		item := phpfront.EnvVar{Name: name, Value: vars[name]}
		if runner.Sensitive(name) {
			item.Value, item.Redacted = "", true
		}
		out = append(out, item)
	}
	return out
}

// frontFields is a service's settings in the shape the boundary passes.
func frontFields(found []integration.Field) []phpfront.IntegrationField {
	out := make([]phpfront.IntegrationField, 0, len(found))
	for _, field := range found {
		out = append(out, phpfront.IntegrationField{
			Name: field.Name, Label: field.Label, Help: field.Help, Placeholder: field.Placeholder,
			Secret: field.Secret, Optional: field.Optional, Who: field.Who, Options: field.Options,
			Generated: field.Generated, Env: field.Env, Default: field.Default,
			Toggle: field.Toggle, DefaultOn: field.DefaultOn, Empty: field.Empty, Source: field.Source,
			Fixed: field.Fixed(),
		})
	}
	return out
}

// frontChecked is the roles and users a service's Who setting names, in the
// order it names them: the view holds them as a map, and what crosses the
// boundary is a list, so the boxes a form checks are in one order.
func frontChecked(kind integration.Kind, values map[string]string) []string {
	var out []string
	for _, field := range kind.Fields {
		if !field.Who {
			continue
		}
		for _, name := range strings.Split(values[field.Name], ",") {
			if name = strings.TrimSpace(name); name != "" && !slices.Contains(out, name) {
				out = append(out, name)
			}
		}
	}
	return out
}

// Usage is what each area of the Data retention table holds now, in the order
// retention.Areas lists them, which is the order the sweep applies them in.
//
// The rules are rows and PHP reads them; what the rules apply to is a walk of
// the recordings, the artifact history and the tool databases, which is this.
func (h *FrontHost) Usage(ctx context.Context) phpfront.Retention {
	if h.Cleaner == nil {
		return phpfront.Retention{Off: true}
	}
	found, err := h.Cleaner.Usage(ctx)
	if err != nil {
		return phpfront.Retention{Error: err.Error()}
	}
	out := phpfront.Retention{}
	for _, area := range retention.Areas {
		usage := found[area.Kind]
		row := phpfront.AreaUsage{Kind: area.Kind, Count: usage.Count, Bytes: usage.Bytes}
		if usage.Oldest != nil {
			row.Oldest = frontTime(*usage.Oldest)
		}
		out.Areas = append(out.Areas, row)
	}
	return out
}

// Telemetry is the options the tracer runs on, the stored ones over the
// defaults, and where the dashboard is mounted.
//
// An instance with no telemetry answers the defaults, the fields it would
// record, and the page draws them read only. Wired names that state; the
// reader's own flag is PHP's.
func (h *FrontHost) Telemetry(ctx context.Context) phpfront.Telemetry {
	opts, path := oida.NewOptions("task-ui"), oida.DefaultPath
	out := phpfront.Telemetry{}
	if h.Tracing != nil {
		stored, err := h.Tracing.Load(ctx)
		if err != nil {
			return phpfront.Telemetry{Error: err.Error()}
		}
		opts, path, out.Wired = stored, h.Tracing.Path(), true
	}
	out.Path = path
	out.Enabled, out.ServiceName = opts.Enabled, opts.ServiceName
	out.SampleRate, out.RingBufferSize = opts.SampleRate, opts.RingBufferSize
	out.TopRequests, out.MaxSpansPerTrace = opts.TopRequests, opts.MaxSpansPerTrace
	out.IgnorePaths, out.AllowedNetworks = opts.IgnorePaths, opts.AllowedNetworks
	out.LiveStream, out.RefreshInterval = opts.LiveStream, opts.RefreshInterval
	out.OptionPath, out.CaptureLogs = opts.Path, opts.CaptureLogs
	out.TrackMemoryUse, out.TrustRequestID = opts.TrackMemoryUse, opts.TrustRequestID
	out.DebugSpanMinDuration = opts.DebugSpanMinDuration.String()
	return out
}

// Modules are the plugins the instance serves under /m, which the Sidebar
// settings list beside the pages: a module is a tree on disk that registered
// itself at startup and is in no table.
func (h *FrontHost) Modules(_ context.Context) []phpfront.Plugin {
	if h.Plugins == nil {
		return nil
	}
	found := h.Plugins.Modules()
	out := make([]phpfront.Plugin, 0, len(found))
	for _, module := range found {
		out = append(out, phpfront.Plugin{Name: module.Name(), Title: module.Label(), Description: module.Description, Icon: module.Icon})
	}
	return out
}

// Entries are the project directories under the src of each workspace, as the
// Projects settings list them.
//
// workspaces is the folder of each workspace row, in the order the settings
// list them: the walk reads nothing else of a row, so the rows stay PHP's.
func (h *FrontHost) Entries(_ context.Context, workspaces []string) []phpfront.Project {
	rows := make([]model.Workspace, 0, len(workspaces))
	for _, name := range workspaces {
		rows = append(rows, model.Workspace{Folder: name})
	}
	found := folder.Entries(h.Root, h.Git, rows)
	out := make([]phpfront.Project, 0, len(found))
	for _, entry := range found {
		item := phpfront.Project{ID: entry.ID, Dir: entry.Dir, Sources: entry.Sources, Jobs: entry.Jobs, Error: entry.Error}
		if info := entry.Checkout; info != nil {
			item.Checkout = true
			item.Remote, item.Branch = info.Remote, info.Branch
			item.Commit, item.Subject, item.When = info.Commit, info.Subject, frontTime(info.When)
		}
		out = append(out, item)
	}
	return out
}
