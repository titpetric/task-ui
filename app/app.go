package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	chi "github.com/go-chi/chi/v5"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/envfile"
	"github.com/titpetric/task-ui/frontend"
	"github.com/titpetric/task-ui/git"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/scheduler"
	"github.com/titpetric/task-ui/sshkey"
	"github.com/titpetric/task-ui/storage"
)

// App is a running instance: the database, the runner, the scheduler and
// the handler that serves the API and the pages.
type App struct {
	Config    *Config
	DB        storage.DB
	Users     *storage.UserStorage
	Runner    *runner.Manager
	Scheduler *scheduler.Loop
	Catalogue *Catalogue
	Guard     *access.Guard

	handler http.Handler
	addr    atomic.Pointer[string]
}

// New opens the database, seeds the users, and builds everything. Runs
// left running by a previous process are marked killed.
func New(ctx context.Context, cfg *Config, version string) (*App, error) {
	for _, dir := range []string{cfg.DataDir, cfg.RecordingsDir(), cfg.ArtifactsDir()} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	db, err := storage.OpenAndMigrate(ctx, cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", cfg.DB, err)
	}
	a := &App{Config: cfg, DB: db, Users: storage.NewUserStorage(db)}
	if err := a.seed(ctx); err != nil {
		db.Close()
		return nil, err
	}
	runs := storage.NewRunStorage(db)
	if n, err := runs.KillActive(ctx); err != nil {
		db.Close()
		return nil, err
	} else if n > 0 {
		slog.Warn("runs left running by the previous process marked killed", "count", n)
	}

	projects := storage.NewProjectStorage(db)
	artifacts := storage.NewArtifactStorage(db)
	schedules := storage.NewScheduleStorage(db)
	secrets := storage.NewSecretStorage(db)
	env := envfile.New(storage.NewEnvFileStorage(db), secrets)
	g := &git.Git{SSHDir: cfg.SSHDir}
	keys := &sshkey.Store{Dir: cfg.SSHDir}
	recordings := &recording.Store{Root: cfg.RecordingsDir()}

	a.Catalogue = NewCatalogue(cfg.Root, cfg.Discover, projects)
	a.Runner = runner.New(runner.Options{
		Runs:         runs,
		Artifacts:    artifacts,
		Secrets:      secrets,
		Env:          env,
		Git:          g,
		Recordings:   recordings,
		ArtifactRoot: cfg.ArtifactsDir(),
		Grace:        cfg.Grace,
		KillTimeout:  cfg.Grace + 5*time.Second,
		Keep:         cfg.Keep,
	})
	a.Scheduler = scheduler.NewLoop(schedules, a.Users, a.Catalogue, a.Runner)

	secret := cfg.Secret
	if secret == "" {
		secret = randomSecret()
	}
	a.Guard = access.NewGuard(a.Users, access.NewSessions(secret, cfg.SessionTTL))
	a.Guard.AutoLogin = cfg.AutoLogin

	apiHandlers := api.New(api.Options{
		Users: a.Users, Projects: projects, Runs: runs, Artifacts: artifacts, Schedules: schedules, Secrets: secrets,
		Runner: a.Runner, Catalogue: a.Catalogue, Git: g, Keys: keys, Env: env, Recordings: recordings, Root: cfg.Root,
	})
	pages, err := frontend.New(frontend.Options{
		Guard: a.Guard, Users: a.Users, Projects: projects, Runs: runs, Artifacts: artifacts, Schedules: schedules, Secrets: secrets,
		Runner: a.Runner, Catalogue: a.Catalogue, Git: g, Keys: keys, Env: env, Root: cfg.Root, Version: version,
	})
	if err != nil {
		db.Close()
		return nil, err
	}

	r := chi.NewRouter()
	r.Mount("/api", a.Guard.Require(access.SameOrigin(apiHandlers.Routes())))
	r.Mount("/", pages.Routes())
	a.handler = r
	return a, nil
}

// seed writes the users file to storage, when there is one.
func (a *App) seed(ctx context.Context) error {
	if a.Config.UsersFile == "" {
		return nil
	}
	file, err := LoadUsersFile(a.Config.UsersFile)
	if err != nil {
		return err
	}
	created, err := file.Seed(ctx, a.Users)
	if err != nil {
		return fmt.Errorf("seeding users from %s: %w", a.Config.UsersFile, err)
	}
	if created > 0 {
		slog.Info("users created from the users file", "count", created, "file", a.Config.UsersFile)
	}
	return nil
}

// randomSecret returns a session secret for one process.
func randomSecret() string {
	buf := make([]byte, 32)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

// Handler returns what serves the API and the pages.
func (a *App) Handler() http.Handler {
	return a.handler
}

// Addr returns the address being served, once Serve has started.
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

// Serve serves on a listener, with the scheduler running, until the
// context ends. Running jobs are killed on the way out.
func (a *App) Serve(ctx context.Context, ln net.Listener) error {
	addr := ln.Addr().String()
	a.addr.Store(&addr)
	slog.Info("task-ui listening", "addr", addr, "root", a.Config.Root, "login", a.Guard.LoginRequired())
	if a.Config.AutoLogin != "" {
		slog.Warn("auto login is on: every request is signed in", "user", a.Config.AutoLogin)
	}

	server := &http.Server{Handler: a.handler, ReadHeaderTimeout: 10 * time.Second}
	go a.Scheduler.Run(ctx)
	errs := make(chan error, 1)
	go func() { errs <- server.Serve(ln) }()

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
	}
	shutdown, cancel := context.WithTimeout(context.Background(), a.Config.Grace+5*time.Second)
	defer cancel()
	a.Runner.Shutdown(shutdown)
	if err := server.Shutdown(shutdown); err != nil && !errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return nil
}

// Close releases the database.
func (a *App) Close() error {
	return a.DB.Close()
}
