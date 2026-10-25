package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	chi "github.com/go-chi/chi/v5"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/frontend"
	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/hostinfo"
	"github.com/titpetric/task-ui/hostrun"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/mcp"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/scheduler"
	"github.com/titpetric/task-ui/skills"
	"github.com/titpetric/task-ui/sshprobe"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/tooling"
	"github.com/titpetric/task-ui/workflow"
)

// App is a running instance: the database, the runner, the scheduler and
// the handler that serves the API and the pages.
type App struct {
	Config    *Config
	DB        storage.DB
	Users     *storage.UserStorage
	Emails    *storage.UserEmailStorage
	Passwords *storage.UserAuthPasswordStorage
	Runner    *runner.Manager
	Scheduler *scheduler.Loop
	Workflows *workflow.Runner
	Catalogue *Catalogue
	Guard     *access.Guard
	// Board is the notes board and the chat.
	Board *note.Board
	// UsersFile is the password file, nil when there is none.
	UsersFile *UsersFile

	handler http.Handler
	addr    atomic.Pointer[string]
}

// New opens the database, reads the users, and builds everything. It
// marks runs a previous process left running as killed.
func New(ctx context.Context, cfg *Config, version string) (*App, error) {
	db, err := openStorage(ctx, cfg)
	if err != nil {
		return nil, err
	}
	a := &App{
		Config: cfg, DB: db,
		Users:     storage.NewUserStorage(db),
		Emails:    storage.NewUserEmailStorage(db),
		Passwords: storage.NewUserAuthPasswordStorage(db),
	}
	if err := a.prepare(ctx); err != nil {
		db.Close()
		return nil, err
	}
	if err := a.wire(version); err != nil {
		db.Close()
		return nil, err
	}
	return a, nil
}

// openStorage makes the data directories and opens the database.
func openStorage(ctx context.Context, cfg *Config) (storage.DB, error) {
	for _, dir := range []string{cfg.DataDir, cfg.RecordingsDir(), cfg.ArtifactsDir()} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	db, err := storage.OpenAndMigrate(ctx, cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", cfg.DB, err)
	}
	return db, nil
}

// prepare reads the users, adds the default workspace, puts the
// workspaces' secrets on their keys, and settles what the previous
// process left.
func (a *App) prepare(ctx context.Context) error {
	if err := a.seed(ctx); err != nil {
		return err
	}
	if err := a.autoLogin(ctx); err != nil {
		return err
	}
	if err := a.importRoles(ctx); err != nil {
		return err
	}
	if err := a.seedCleanup(ctx); err != nil {
		return err
	}
	if err := a.ensureWorkspaces(ctx); err != nil {
		return err
	}
	if err := a.ensureKeys(ctx); err != nil {
		return err
	}
	if err := a.migrateSecrets(ctx); err != nil {
		return err
	}
	n, err := storage.NewRunStorage(a.DB).KillActive(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		slog.Warn("runs left running by the previous process marked killed", "count", n)
	}
	n, err = storage.NewRunStorage(a.DB).FailQueued(ctx, model.QueuedLost)
	if err != nil {
		return err
	}
	if n > 0 {
		slog.Warn("runs left queued by the previous process marked failed", "count", n)
	}
	return storage.NewWorkflowRunStorage(a.DB).Interrupt(ctx, "task-ui restarted while it ran")
}

// seed writes the password file to storage, when there is one. Without
// one the users in storage stay as they are.
func (a *App) seed(ctx context.Context) error {
	if a.Config.Htpasswd == "" {
		slog.Warn("no password file: nobody new can sign in", "hint", "task-ui passwd <user> makes <data>/.htpasswd")
		return nil
	}
	a.UsersFile = NewUsersFile(a.Config.Htpasswd)
	created, removed, err := a.UsersFile.Sync(ctx, a.Users, a.Passwords)
	if errors.Is(err, os.ErrNotExist) {
		// Not made yet: start anyway, the watcher reads it once it exists,
		// so `task-ui passwd` works against a running instance.
		slog.Warn("no password file yet: nobody new can sign in until it is made", "file", a.Config.Htpasswd, "hint", "task-ui passwd <user>")
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading users: %w", err)
	}
	slog.Info("users read from the password file", "file", a.Config.Htpasswd, "created", created, "removed", removed)
	return nil
}

// autoLogin makes the auto login user when there is none, as an admin:
// the knob is for development, where there may be no password file at
// all, and a user that cannot sign in would make it useless.
func (a *App) autoLogin(ctx context.Context) error {
	name := a.Config.AutoLogin
	if name == "" {
		return nil
	}
	if _, err := a.Users.GetByName(ctx, name); err == nil {
		return nil
	}
	slog.Warn("auto login names a user that is not there: created as an admin", "user", name)
	return a.Users.Create(ctx, &model.User{Name: name, Role: model.RoleRunner, Admin: true, Mcp: true})
}

// wire builds the services and the handler over the open database.
func (a *App) wire(version string) error {
	cfg, db := a.Config, a.DB
	runs := storage.NewRunStorage(db)
	projects := storage.NewProjectStorage(db)
	workspaces := storage.NewWorkspaceStorage(db)
	artifacts := storage.NewArtifactStorage(db)
	schedules := storage.NewScheduleStorage(db)
	workflows := storage.NewWorkflowStorage(db)
	workflowRuns := storage.NewWorkflowRunStorage(db)
	refs := storage.NewFolderRefStorage(db)
	keys := a.keys()
	g := &git.Git{Keys: keys}
	secrets := a.secrets()
	env := envfile.New(storage.NewEnvFileStorage(db), secrets)
	integrations := &integration.Store{Secrets: secrets}
	prober := &sshprobe.Prober{
		Users: []string{"debian", "root"}, Keys: keys,
		KnownHosts: keys.KnownHosts(), Timeout: 4 * time.Second, TTL: 10 * time.Minute,
		Terminal: true,
	}
	hosts := &beszel.Inventory{TTL: time.Minute, Settings: func(ctx context.Context) (string, string, error) {
		values, err := integrations.Values(ctx, integration.Beszel)
		return values["url"], values["token"], err
	}}
	recordings := &recording.Store{Root: cfg.RecordingsDir()}

	a.Catalogue = NewCatalogue(cfg.Root, cfg.Discover, projects)
	a.Catalogue.Workspaces = workspaces
	a.Catalogue.Skills = &skills.Prober{Timeout: 3 * time.Second, TTL: 30 * time.Second, Home: a.Catalogue.Home}
	// The detector reads a folder's pipeline from the same listing the
	// skill chips come from, so one atkins answers both.
	detector.UseProber(a.Catalogue.Skills)
	self := cfg.ComposeProject
	if self == "" {
		self = ComposeProject("docker")
	}
	store := &artifactstore.Store{Dir: cfg.ArtifactTree(), Keep: cfg.ArtifactHistory}
	cleaner := NewCleaner(db, cfg, os.Getenv("HOME"))
	cleaner.Store = store
	summaries := storage.NewRunSummaryStorage(a.DB)
	secretReads := storage.NewSecretReadStorage(a.DB)
	mounts, hostData := hostPaths(cfg)
	a.Runner = runner.New(runner.Options{
		HostPaths: mounts, HostData: hostData,
		Keys:          keys,
		Runs:          runs,
		Artifacts:     artifacts,
		Summaries:     summaries,
		Secrets:       secrets,
		SecretReads:   secretReads,
		Env:           env,
		Git:           g,
		Recordings:    recordings,
		ArtifactRoot:  cfg.ArtifactsDir(),
		ArtifactStore: store,
		Integrations:  integrations,
		Grace:         cfg.Grace,
		KillTimeout:   cfg.Grace + 5*time.Second,
		Keep:          cfg.Keep,
		Now:           cfg.Clock(),
		Root:          cfg.Root,
		Workspaces:    workspaces,
		Users:         a.Users,
		Backlog:       &runner.Backlog{Settings: storage.NewSettingStorage(a.DB)},
		Environment:   &runner.Environment{Settings: storage.NewSettingStorage(a.DB)},
		APIURL:        func() string { return strings.TrimSuffix(MCPURL(a.Addr()), "/mcp") },
	})
	hostRuns := &hostrun.Starter{Catalogue: a.Catalogue, Hosts: hosts, SSH: prober, Runner: a.Runner}
	a.Scheduler = scheduler.NewLoop(schedules, a.Users, a.Catalogue.Steps(), a.Runner)
	a.Scheduler.Hosts = hostRuns
	a.Workflows = workflow.New(workflows, workflowRuns, a.Catalogue.Steps(), a.Runner)
	a.Workflows.Hosts = hostRuns
	a.Scheduler.Workflows, a.Scheduler.Flows = workflows, a.Workflows
	events := event.NewBus()
	a.Workflows.Events = events
	a.Runner.Events = events
	a.Guard = access.NewGuard(a.Users, access.NewSessions(storage.NewUserSessionStorage(a.DB), a.Users, cfg.SessionTTL))
	a.Guard.AutoLogin = cfg.AutoLogin
	a.Guard.Tokens = storage.NewUserAuthTokenStorage(a.DB)
	a.Guard.Now = cfg.Clock()
	a.Guard.RunTokens = access.NewRunTokens()
	a.Board = &note.Board{
		Notes: storage.NewNoteStorage(a.DB), Users: a.Users, Runs: runs, WorkflowRuns: workflowRuns, RunSummaries: summaries, Catalogue: a.Catalogue, Runner: a.Runner,
		Recordings: recordings, Events: events, Root: cfg.Root, Workspaces: workspaces,
		Hosts: promptHosts(hosts, prober), BaseURL: cfg.URL,
		RunTokens: a.Guard.RunTokens, MCPURL: func() string { return MCPURL(a.Addr()) },
	}
	a.Workflows.Prompter = a.Board
	a.Scheduler.Board = a.Board
	if err := a.Board.Recover(context.Background()); err != nil {
		return err
	}

	var addUser func(ctx context.Context, name, password string) error
	var removeUser func(ctx context.Context, name string) error
	if a.UsersFile != nil {
		addUser = func(ctx context.Context, name, password string) error {
			return a.UsersFile.Add(ctx, a.Users, a.Passwords, name, password)
		}
		removeUser = func(ctx context.Context, name string) error {
			return a.UsersFile.Remove(ctx, a.Users, a.Passwords, name)
		}
	}
	pins := storage.NewPinStorage(db)
	apiHandlers := api.New(api.Options{
		Users: a.Users, Projects: projects, Workspaces: workspaces, Runs: runs, Artifacts: artifacts, ArtifactStore: store, Schedules: schedules, Secrets: secrets, SecretReads: secretReads,
		FolderRefs: refs,
		Workflows:  workflows, WorkflowRuns: workflowRuns, Flows: a.Workflows, Events: events,
		Runner: a.Runner, Catalogue: a.Catalogue, Steps: a.Catalogue.Steps(), Git: g, Keys: keys, Env: env, Recordings: recordings, Root: cfg.Root, Board: a.Board,
		Integrations: integrations, Hosts: hosts, HostRuns: hostRuns, Summarizer: &claude.Summarizer{Timeout: 3 * time.Minute},
		AddUser: addUser, RemoveUser: removeUser, Tokens: a.Guard.Tokens, Now: cfg.Clock(), Summaries: summaries, Version: version,
		Cleaner: cleaner, Pins: pins,
	})
	pages, err := frontend.New(frontend.Options{
		Guard: a.Guard, Auth: access.NewAuthenticator(a.Users, a.Emails, a.Passwords), Users: a.Users, Projects: projects, Workspaces: workspaces, Runs: runs, Artifacts: artifacts, Schedules: schedules, Secrets: secrets, SecretReads: secretReads,
		FolderRefs: refs,
		Workflows:  workflows, WorkflowRuns: workflowRuns, Flows: a.Workflows,
		Runner: a.Runner, Catalogue: a.Catalogue, Steps: a.Catalogue.Steps(), Git: g, Keys: keys, Env: env, Recordings: recordings, Root: cfg.Root, Version: version, Now: cfg.Clock(),
		ArtifactStore: store, Disk: &diskusage.Meter{TTL: 10 * time.Minute}, Board: a.Board, ComposeProject: self, Cleaner: cleaner,
		Storage:      &diskusage.Inventory{TTL: 30 * time.Second, Wait: 2 * time.Second, Items: storageItems(cfg, workspaces, os.Getenv("HOME"))},
		Integrations: integrations, Hosts: hosts, Pins: pins,
		Tooling: &tooling.Prober{Timeout: 5 * time.Second, TTL: time.Minute},
		SSH:     prober, AddUsers: a.UsersFile != nil, BaseURL: cfg.URL,
		Facts: &hostinfo.Reader{Run: prober.Output, Timeout: 20 * time.Second, TTL: 5 * time.Minute},
		Instance: &view.Instance{
			URL: cfg.URL, Listen: cfg.Listen, Root: cfg.Root, DataDir: cfg.DataDir, SSHDir: cfg.SSHDir, Htpasswd: cfg.Htpasswd, HostData: hostDataShown(cfg.DataDir, hostData),
			Discover: cfg.Discover, AutoLogin: cfg.AutoLogin, TimeZone: timeZone(cfg.Clock()()),
			SessionTTL: cfg.SessionTTL.String(), Keep: cfg.Keep.String(), Grace: cfg.Grace.String(), ArtifactHistory: cfg.ArtifactHistory, ComposeProject: self,
		},
	})
	if err != nil {
		return err
	}

	routes := apiHandlers.Routes()
	r := chi.NewRouter()
	// MCP clients sign with a bearer token; its tools call the API as
	// the token's user.
	server := mcp.New(routes, version)
	server.BaseURL = cfg.URL
	// A run's task-ui get-secret signs in with the run's own token,
	// which no other route takes.
	r.Get("/api/run/secrets/{name}", (&api.RunSecret{Source: a.Runner}).ServeHTTP)
	// Forgejo posts a push with no session: the hook's signature under
	// the integration's secret signs it in, and it only ever pulls.
	r.Post("/api/webhooks/forgejo", (&api.Webhook{
		Settings: func(ctx context.Context) (map[string]string, error) {
			return integrations.Values(ctx, integration.Forgejo)
		},
		Projects:   projects.List,
		Workspaces: workspaces.List,
		Root:       cfg.Root,
		Pull:       a.Runner.PullCheckout,
	}).ServeHTTP)
	r.Mount("/api/mcp", a.Guard.Require(access.RequireMCP(access.SameOrigin(server))))
	r.Mount("/api", a.Guard.Require(access.SameOrigin(routes)))
	r.Mount("/", pages.Routes())
	a.handler = r
	return nil
}

// Handler returns what serves the API and the pages.
func (a *App) Handler() http.Handler {
	return a.handler
}

// Addr returns the address Serve listens on, once Serve has started.
func (a *App) Addr() string {
	if addr := a.addr.Load(); addr != nil {
		return *addr
	}
	return ""
}

// Run listens on the configured address and serves until the context
// ends.
func (a *App) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", a.Config.Listen)
	if err != nil {
		return err
	}
	return a.Serve(ctx, ln)
}

// Serve serves on a listener, with the scheduler and the password file
// watch running, until the context ends. It kills running jobs on the
// way out.
func (a *App) Serve(ctx context.Context, ln net.Listener) error {
	addr := ln.Addr().String()
	a.addr.Store(&addr)
	slog.Info("task-ui listening", "addr", addr, "root", a.Config.Root, "login", a.Guard.LoginRequired())
	if a.Config.AutoLogin != "" {
		slog.Warn("auto login is on: every request is signed in", "user", a.Config.AutoLogin)
	}

	go a.Scheduler.Run(ctx)
	if a.UsersFile != nil {
		go a.UsersFile.Watch(ctx, a.Users, a.Passwords, func(err error) { slog.Error("re-reading the password file", "err", err) })
	}
	server := &http.Server{Handler: a.handler, ReadHeaderTimeout: 10 * time.Second}
	errs := make(chan error, 1)
	go func() { errs <- server.Serve(ln) }()

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
		return a.shutdown(server)
	}
}

// shutdown ends the runs and the server, within the grace.
func (a *App) shutdown(server *http.Server) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.Config.Grace+5*time.Second)
	defer cancel()
	a.Runner.Shutdown(ctx)
	if err := server.Shutdown(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return nil
}

// Close releases the database.
func (a *App) Close() error {
	return a.DB.Close()
}

// ensureWorkspaces adds the default workspace's row to an instance with
// no workspace at all, its first start, and makes the home and src of
// every workspace there is, so a run has a HOME and a clone has a place
// before the first page; a default workspace deleted stays gone, and the
// root gets no home or src back.
func (a *App) ensureWorkspaces(ctx context.Context) error {
	workspaces := storage.NewWorkspaceStorage(a.DB)
	if err := workspaces.EnsureDefault(ctx); err != nil {
		return err
	}
	rows, err := workspaces.List(ctx)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if err := layout.Ensure(a.Config.Root, row.Folder); err != nil {
			return err
		}
	}
	return nil
}

// storageItems names what the Storage table of the Instance page
// measures: the database with its journal, the recordings, the
// artifacts, the tool history of each workspace's home and the
// instance's own copy under the server's HOME, and the checkouts under
// each workspace's src. Every measure reads the workspaces again.
func storageItems(cfg *Config, workspaces *storage.WorkspaceStorage, serverHome string) func() []diskusage.Item {
	return func() []diskusage.Item {
		db, _, _ := strings.Cut(cfg.DB, "?")
		items := []diskusage.Item{
			{Name: "task-ui database", Where: db, Paths: []string{db, db + "-wal", db + "-shm"}},
			{Name: "Recordings", Where: cfg.RecordingsDir(), Paths: []string{cfg.RecordingsDir()}},
			{Name: "Artifacts", Where: cfg.ArtifactsDir(), Paths: []string{cfg.ArtifactsDir()}},
		}
		rows, _ := workspaces.List(context.Background())
		for _, row := range rows {
			history := filepath.Join(layout.HomeDir(cfg.Root, row.Folder), ".atkins", "atkins.db")
			items = append(items, diskusage.Item{Name: "Tool history · " + row.Folder, Where: history, Paths: []string{history, history + "-wal", history + "-shm"}})
		}
		if serverHome != "" {
			history := filepath.Join(serverHome, ".atkins", "atkins.db")
			items = append(items, diskusage.Item{Name: "Tool history · instance", Where: history, Paths: []string{history, history + "-wal", history + "-shm"}})
		}
		for _, row := range rows {
			src := layout.SrcDir(cfg.Root, row.Folder)
			items = append(items, diskusage.Item{Name: "Checkouts · " + row.Folder, Where: src, Paths: []string{src}})
		}
		return items
	}
}
