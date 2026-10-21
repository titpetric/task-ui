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
	"sync/atomic"
	"time"

	chi "github.com/go-chi/chi/v5"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/artifactstore"
	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/diskusage"
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/event"
	"github.com/titpetric/task-ui/frontend"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/mcp"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/scheduler"
	"github.com/titpetric/task-ui/secret"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/sshprobe"
	"github.com/titpetric/task-ui/storage"
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

// prepare reads the users, adds the default workspace, and settles what
// the previous process left.
func (a *App) prepare(ctx context.Context) error {
	if err := a.seed(ctx); err != nil {
		return err
	}
	if err := a.autoLogin(ctx); err != nil {
		return err
	}
	if err := storage.NewWorkspaceStorage(a.DB).EnsureDefault(ctx); err != nil {
		return err
	}
	if err := a.ensureKeys(ctx); err != nil {
		return err
	}
	n, err := storage.NewRunStorage(a.DB).KillActive(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		slog.Warn("runs left running by the previous process marked killed", "count", n)
	}
	return storage.NewWorkflowRunStorage(a.DB).Interrupt(ctx, "task-ui restarted while it ran")
}

// ensureKeys makes the key directories of every workspace, the default
// one among them, and its encryption key when it has none.
func (a *App) ensureKeys(ctx context.Context) error {
	workspaces, err := storage.NewWorkspaceStorage(a.DB).List(ctx)
	if err != nil {
		return err
	}
	keyring := a.Config.Keyring()
	for _, ws := range workspaces {
		made, err := keyring.Ensure(ws.ID, "task-ui "+ws.Folder)
		if err != nil {
			return fmt.Errorf("the keys of workspace %s: %w", ws.Folder, err)
		}
		if made {
			slog.Info("workspace encryption key generated", "workspace", ws.Folder, "dir", keyring.Dir(ws.ID))
		}
	}
	return nil
}

// seed writes the password file to storage, when there is one. Without
// one the users in storage stay as they are.
func (a *App) seed(ctx context.Context) error {
	if a.Config.Htpasswd == "" {
		slog.Warn("no password file: nobody new can sign in", "hint", "set TASK_UI_HTPASSWD or put .htpasswd in the root")
		return nil
	}
	a.UsersFile = NewUsersFile(a.Config.Htpasswd, a.Config.Admins, a.Config.Viewers)
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
	slog.Info("users read from the password file", "file", a.Config.Htpasswd, "created", created, "removed", removed, "admins", a.Config.Admins)
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
	return a.Users.Create(ctx, &model.User{Name: name, Role: model.RoleAdmin})
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
	g := &git.Git{SSHDir: cfg.SSHDir}
	keys := &sshkey.Store{Dir: cfg.SSHDir}
	var homeKeys *sshkey.Store
	if dir := filepath.Join(home(), ".ssh"); filepath.Clean(dir) != filepath.Clean(cfg.SSHDir) {
		homeKeys = &sshkey.Store{Dir: dir}
	}
	keyPolicies := storage.NewSSHKeyPolicyStorage(db)
	keyring, grants := cfg.Keyring(), storage.NewWorkspaceKeyGrantStorage(db)
	secrets := secret.New(storage.NewSecretStorage(db), secret.NewVault(keys))
	env := envfile.New(storage.NewEnvFileStorage(db), secrets)
	integrations := &integration.Store{Secrets: secrets}
	prober := &sshprobe.Prober{
		Users: []string{"debian", "root"}, KeyDirs: []string{cfg.SSHDir, filepath.Join(home(), ".ssh")},
		KnownHosts: filepath.Join(cfg.SSHDir, "known_hosts"), Timeout: 4 * time.Second, TTL: 10 * time.Minute,
		Terminal: true,
	}
	hosts := &beszel.Inventory{TTL: time.Minute, Settings: func(ctx context.Context) (string, string, error) {
		values, err := integrations.Values(ctx, integration.Beszel)
		return values["url"], values["token"], err
	}}
	recordings := &recording.Store{Root: cfg.RecordingsDir()}

	a.Catalogue = NewCatalogue(cfg.Root, cfg.Discover, projects)
	a.Catalogue.Agent = cfg.Agent
	store := &artifactstore.Store{Dir: cfg.ArtifactTree(), Keep: cfg.ArtifactHistory}
	a.Runner = runner.New(runner.Options{
		KeyDir: keys.Dir, KeyPolicies: keyPolicies,
		Runs:          runs,
		Artifacts:     artifacts,
		Secrets:       secrets,
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
		WorkspaceKeys: keyring,
		KeyGrants:     grants,
	})
	a.Scheduler = scheduler.NewLoop(schedules, a.Users, a.Catalogue, a.Runner)
	a.Workflows = workflow.New(workflows, workflowRuns, a.Catalogue, a.Runner)
	events := event.NewBus()
	a.Workflows.Events = events
	a.Runner.Events = events
	a.Guard = access.NewGuard(a.Users, access.NewSessions(storage.NewUserSessionStorage(a.DB), a.Users, cfg.SessionTTL))
	a.Guard.AutoLogin = cfg.AutoLogin
	a.Guard.Tokens = storage.NewUserAuthTokenStorage(a.DB)
	a.Board = &note.Board{
		Notes: storage.NewNoteStorage(a.DB), Users: a.Users, Catalogue: a.Catalogue, Runner: a.Runner,
		Recordings: recordings, Events: events, Agent: cfg.Agent, AgentJob: cfg.AgentJob,
	}
	if err := a.Board.Recover(context.Background()); err != nil {
		return err
	}

	var addUser func(ctx context.Context, name, password string) error
	if a.UsersFile != nil {
		addUser = func(ctx context.Context, name, password string) error {
			return a.UsersFile.Add(ctx, a.Users, a.Passwords, name, password)
		}
	}
	apiHandlers := api.New(api.Options{
		Users: a.Users, Projects: projects, Workspaces: workspaces, Runs: runs, Artifacts: artifacts, ArtifactStore: store, Schedules: schedules, Secrets: secrets,
		WorkspaceKeys: keyring, KeyGrants: grants,
		Workflows: workflows, WorkflowRuns: workflowRuns, Flows: a.Workflows, Events: events,
		Runner: a.Runner, Catalogue: a.Catalogue, Git: g, Keys: keys, KeyPolicies: keyPolicies, Env: env, Recordings: recordings, Root: cfg.Root, Board: a.Board,
		Integrations: integrations, Hosts: hosts, SSH: prober, Summarizer: &claude.Summarizer{Timeout: 3 * time.Minute},
		AddUser: addUser, Summaries: storage.NewRunSummaryStorage(a.DB),
	})
	pages, err := frontend.New(frontend.Options{
		Guard: a.Guard, Auth: access.NewAuthenticator(a.Users, a.Emails, a.Passwords), Users: a.Users, Projects: projects, Workspaces: workspaces, Runs: runs, Artifacts: artifacts, Schedules: schedules, Secrets: secrets,
		WorkspaceKeys: keyring, KeyGrants: grants,
		Workflows: workflows, WorkflowRuns: workflowRuns, Flows: a.Workflows,
		Runner: a.Runner, Catalogue: a.Catalogue, Git: g, Keys: keys, HomeKeys: homeKeys, KeyPolicies: keyPolicies, Env: env, Recordings: recordings, Root: cfg.Root, Version: version, Now: cfg.Clock(),
		ArtifactStore: store, Disk: &diskusage.Meter{TTL: 10 * time.Minute}, Board: a.Board,
		Integrations: integrations, Hosts: hosts,
		SSH: prober, AddUsers: a.UsersFile != nil,
	})
	if err != nil {
		return err
	}

	routes := apiHandlers.Routes()
	r := chi.NewRouter()
	// MCP clients sign with a bearer token; its tools call the API as
	// the token's user.
	r.Mount("/api/mcp", a.Guard.Require(access.SameOrigin(mcp.New(routes, version))))
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

// home is the user's home directory, or / when there is none.
func home() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "/"
	}
	return dir
}
