package app

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/titpetric/task-ui/access"
	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/claude"
	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/frontend"
	"github.com/titpetric/task-ui/hostrun"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assets"
	"github.com/titpetric/task-ui/mcp"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/phpfront"
	"github.com/titpetric/task-ui/phpscript"
	"github.com/titpetric/task-ui/phpscripts"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/scheduler"
	"github.com/titpetric/task-ui/skills"
	"github.com/titpetric/task-ui/spotify"
	"github.com/titpetric/task-ui/workflow"
)

// modelled reports whether a model is set up for task-ui's own asks, Claude Code's or OpenRouter's.
func modelled(ctx context.Context, store *integration.Store) bool {
	if store == nil {
		return false
	}
	for _, id := range []string{integration.Claude, integration.OpenRouter} {
		values, err := store.Values(ctx, id)
		if err != nil {
			continue
		}
		if kind := integration.KindOf(id); kind != nil && integration.Configured(*kind, values) {
			return true
		}
	}
	return false
}

// wire builds the instance's own parts and the handler over its storage and services.
func (a *App) wire(version string) error {
	cfg, store, svc := a.Config, a.Storage, a.Services
	// The sidebar's top block as it stood before the table held it, on the
	// first start of an instance. A table with a row in it is left alone,
	// so the order a user saved survives a restart.
	if err := store.Sidebar.EnsureDefaults(context.Background(), model.SidebarDefaults()); err != nil {
		return err
	}

	// The model names which case a failed pull is, out of git's closed list, and task-ui runs that case's mend itself.
	if svc.Git != nil && modelled(context.Background(), svc.Integrations) {
		triage := &claude.Triager{Timeout: 90 * time.Second, Env: func(ctx context.Context) (map[string]string, error) {
			return svc.Integrations.Env(ctx, "", "")
		}}
		svc.Git.Triage = triage.Triage
	}

	a.Catalogue = NewCatalogue(cfg.Root, cfg.Discover, store.Projects)
	a.Catalogue.Workspaces = store.Workspaces
	a.Catalogue.Skills = &skills.Prober{Timeout: 3 * time.Second, TTL: 30 * time.Second, Home: a.Catalogue.Home}
	// The detector reads a folder's pipeline from the same listing the
	// skill chips come from, so one atkins answers both.
	detector.UseProber(a.Catalogue.Skills)
	a.Runner = runner.New(runner.Options{
		HostPaths: svc.mounts, HostData: svc.hostData,
		Keys:          svc.Keystore,
		Runs:          store.Runs,
		Artifacts:     store.Artifacts,
		Summaries:     store.RunSummaries,
		Secrets:       svc.Secrets,
		SecretReads:   store.SecretReads,
		Git:           svc.Git,
		Recordings:    svc.Recordings,
		ArtifactRoot:  cfg.ArtifactsDir(),
		ArtifactStore: svc.Artifacts,
		Integrations:  svc.Integrations,
		Grace:         cfg.Grace,
		KillTimeout:   cfg.Grace + 5*time.Second,
		Keep:          cfg.Keep,
		Now:           cfg.Clock(),
		Root:          cfg.Root,
		Workspaces:    store.Workspaces,
		Users:         store.Users,
		Backlog:       &runner.Backlog{Settings: store.Settings},
		Environment:   &runner.Environment{Settings: store.Settings},
		APIURL:        func() string { return strings.TrimSuffix(MCPURL(a.Addr()), "/mcp") },
		Tooling:       toolSummary(store.Tools),
	})
	hostRuns := &hostrun.Starter{Catalogue: a.Catalogue, Hosts: svc.Hosts, SSH: svc.SSH, Runner: a.Runner}
	a.Scheduler = scheduler.NewLoop(store.Schedules, store.Users, a.Catalogue.Steps(), a.Runner)
	a.Scheduler.Hosts = hostRuns
	a.Workflows = workflow.New(store.Workflows, store.WorkflowRuns, a.Catalogue.Steps(), a.Runner)
	a.Workflows.Hosts = hostRuns
	a.Workflows.Users = store.Users
	a.Workflows.Projects = a.Catalogue
	a.Scheduler.Flows = a.Workflows
	a.Workflows.Events = svc.Events
	a.Runner.Events = svc.Events
	a.Guard = access.NewGuard(access.NewSessions(store.Sessions, store.Users,
		access.SessionPolicy{TTL: cfg.SessionTTL, Max: cfg.SessionMax}))
	a.Guard.AutoLogin = cfg.AutoLogin
	a.Guard.Tokens = store.Tokens
	a.Guard.Now = cfg.Clock()
	a.Guard.RunTokens = access.NewRunTokens()
	a.Board = &note.Board{
		Notes: store.Notes, Users: store.Users, Runs: store.Runs, WorkflowRuns: store.WorkflowRuns, RunSummaries: store.RunSummaries, Catalogue: a.Catalogue, Runner: a.Runner,
		Recordings: svc.Recordings, Events: svc.Events, Root: cfg.Root, Workspaces: store.Workspaces,
		Hosts: promptHosts(svc.Hosts, svc.SSH), BaseURL: cfg.URL,
		RunTokens: a.Guard.RunTokens, MCPURL: func() string { return MCPURL(a.Addr()) },
	}
	a.Workflows.Prompter = a.Board
	a.Workflows.Output = a.Board
	a.Scheduler.Board = a.Board
	if err := a.Board.Recover(context.Background()); err != nil {
		return err
	}

	// The modules: the tree in the binary lands in the data folder, then
	// every root is walked for a modinfo.php and each one is run once.
	//
	// A module that will not register is a warning and not a failure. The
	// instance serves its own pages either way, and a plugin that throws
	// on startup should not be the reason nobody can sign in.
	if err := phpscripts.Extract(cfg.ModulesDir()); err != nil {
		return err
	}
	// One dispatcher and one lookup on the instance, read by both the
	// plugins and the front end: a page's run and a module's run are the
	// same run, differing only in the tree the action file sits in.
	dispatcher := NewDispatcher(a.Board, store.Users, svc.Artifacts)
	moduleRuns := NewModuleRuns(store.Notes)

	moduleOpts := phpscript.NewOptions(cfg.ModulesDir())
	moduleOpts.Roots = append(moduleOpts.Roots, cfg.Root)
	moduleOpts.Registry = phpscript.NewStore(store.Modules, store.Sidebar)
	moduleOpts.Dispatcher = dispatcher
	moduleOpts.Runs = moduleRuns
	moduleOpts.Database = phpscript.NewDatabase(store.DB)
	moduleOpts.Artifacts = phpscript.NewTree(svc.Artifacts)
	moduleOpts.Tracer = svc.Telemetry.Tracer()
	moduleOpts.Version = version
	moduleOpts.OnError = func(err error) { slog.Warn("a module did not register", "err", err) }
	a.Modules = phpscript.New(moduleOpts)
	if err := a.Modules.Start(context.Background()); err != nil {
		return err
	}

	// The front end being ported. Its tree is extracted beside the modules'
	// and it answers under FrontBase; every row it draws it reads itself,
	// through the same pool, and FrontHost answers what is not a row.
	frontRoot, err := phpfront.Extract(cfg.DataDir)
	if err != nil {
		return err
	}
	frontOpts := phpfront.NewOptions(frontRoot)
	frontOpts.Base = FrontBase
	frontOpts.Version = version
	frontOpts.Login = a.Guard.LoginRequired()
	frontOpts.Now = cfg.Clock()
	frontOpts.Database = phpscript.NewDatabase(store.DB)
	frontOpts.Artifacts = phpscript.NewTree(svc.Artifacts)
	frontOpts.Tracer = svc.Telemetry.Tracer()
	frontOpts.Host = &FrontHost{
		Catalogue: a.Catalogue, Runner: a.Runner, Workflows: a.Workflows, Git: svc.Git, Events: svc.Events,
		Recordings: svc.Recordings, Integrations: svc.Integrations, Secrets: svc.Secrets, Root: cfg.Root,
		Artifacts: svc.Artifacts, Disk: svc.Disk, Now: cfg.Clock(),
		Hosts: svc.Hosts, SSH: svc.SSH, Facts: svc.Facts,
		Tools: svc.Tooling, Compose: cfg.ComposeProject,
		Keystore: svc.Keystore, Config: svc.Instance,
		Tracing: svc.Telemetry, Plugins: a.Modules,
		Cleaner: svc.Cleaner, Storage: svc.Measured,
		Dispatcher: dispatcher, Runs: moduleRuns, FrontRoot: frontRoot,
	}
	frontOpts.OnError = func(err error) { slog.Warn("the front end", "err", err) }
	a.Front, err = phpfront.New(frontOpts)
	if err != nil {
		return err
	}

	var addUser func(ctx context.Context, name, password string) error
	var removeUser func(ctx context.Context, name string) error
	if a.UsersFile != nil {
		addUser = func(ctx context.Context, name, password string) error {
			return a.UsersFile.Add(ctx, store.Users, store.Passwords, name, password)
		}
		removeUser = func(ctx context.Context, name string) error {
			return a.UsersFile.Remove(ctx, store.Users, store.Passwords, name)
		}
	}
	apiHandlers := api.New(api.Options{
		Spotify: svc.Spotify,
		Users:   store.Users, Projects: store.Projects, Workspaces: store.Workspaces, Runs: store.Runs, Artifacts: store.Artifacts, ArtifactStore: svc.Artifacts, Schedules: store.Schedules, Secrets: svc.Secrets, SecretReads: store.SecretReads,
		FolderRefs: store.FolderRefs,
		Workflows:  store.Workflows, WorkflowRuns: store.WorkflowRuns, Flows: a.Workflows, Events: svc.Events,
		Runner: a.Runner, Catalogue: a.Catalogue, Steps: a.Catalogue.Steps(), Git: svc.Git, Keys: svc.Keystore, Env: svc.Env, Recordings: svc.Recordings, Root: cfg.Root, Board: a.Board,
		Integrations: svc.Integrations, Hosts: svc.Hosts, HostRuns: hostRuns, Summarizer: &claude.Summarizer{Timeout: 3 * time.Minute},
		AddUser: addUser, RemoveUser: removeUser, Tokens: a.Guard.Tokens, Now: cfg.Clock(), Summaries: store.RunSummaries, Version: version,
		Cleaner: svc.Cleaner, Pins: store.Pins, Tools: store.Tools,
		Sidebar: store.Sidebar, Modules: a.Modules, Telemetry: svc.Telemetry,
	})
	pages, err := frontend.New(frontend.Options{
		Guard: a.Guard, Auth: access.NewAuthenticator(store.Users, store.Emails, store.Passwords), Users: store.Users, Projects: store.Projects, Workspaces: store.Workspaces, Runs: store.Runs, Artifacts: store.Artifacts, Schedules: store.Schedules, Secrets: svc.Secrets, SecretReads: store.SecretReads,
		FolderRefs: store.FolderRefs,
		Workflows:  store.Workflows, WorkflowRuns: store.WorkflowRuns, Flows: a.Workflows,
		Runner: a.Runner, Catalogue: a.Catalogue, Steps: a.Catalogue.Steps(), Git: svc.Git, Keys: svc.Keystore, Recordings: svc.Recordings, Root: cfg.Root, Version: version, Now: cfg.Clock(),
		ArtifactStore: svc.Artifacts, Disk: svc.Disk, Board: a.Board, ComposeProject: svc.Instance.ComposeProject, Cleaner: svc.Cleaner,
		Storage:      svc.Measured,
		Integrations: svc.Integrations, Hosts: svc.Hosts, Pins: store.Pins,
		Sidebar: store.Sidebar, Modules: a.Modules, Telemetry: svc.Telemetry,
		Tooling: svc.Tooling,
		Tools:   store.Tools,
		SSH:     svc.SSH, AddUsers: a.UsersFile != nil, BaseURL: cfg.URL,
		Facts:    svc.Facts,
		Instance: svc.Instance,
	})
	if err != nil {
		return err
	}

	routes := apiHandlers.Routes()
	r := chi.NewRouter()
	// A page is a quarter of a megabyte of HTML and the JSON behind it is
	// bigger, which puts most of a reader's wait in the transfer: /runs
	// answers in 2.2s and takes another 0.6s to arrive. gzip at level 5
	// costs single digit milliseconds on a body this size. text/event-stream
	// is not a compressible type here, so the SSE stream passes through, and
	// a response that set its own Content-Encoding is left alone.
	r.Use(middleware.Compress(5))
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
			return svc.Integrations.Values(ctx, integration.Forgejo)
		},
		Projects:   store.Projects.List,
		Workspaces: store.Workspaces.List,
		Root:       cfg.Root,
		Pull:       a.Runner.PullCheckout,
	}).ServeHTTP)
	// Spotify sends the browser back with the session cookie withheld
	// (SameSite=Strict); the page hands it on to the route behind the Guard.
	r.Get(spotify.CallbackPath, api.SpotifyCallback)
	r.Mount("/api/mcp", a.Guard.Require(access.RequireMCP(access.SameOrigin(server))))
	r.Mount("/api", a.Guard.Require(access.SameOrigin(routes)))
	// The ported front end, which answers at the root: it is the front end
	// the instance serves. Its static files sit beside the pages and not under
	// them, so an asset URL is the same name whatever holds the root.
	var ported http.Handler
	if a.Front != nil {
		front, err := a.Front.Routes(context.Background())
		if err != nil {
			return err
		}
		r.Mount(assets.Prefix, phpfront.Policy(a.Front.Assets()))
		guarded := a.Front.Guarded(a.Guard, access.SameOrigin(front))
		ported = phpfront.Policy(guarded)
	}
	// The telemetry dashboard. The tracer sits behind the Guard so a
	// trace is read by a signed in user, and behind its own Authorize,
	// which answers anyone but an admin with a 404.
	r.Mount(svc.Telemetry.Path(), a.Guard.Require(svc.Telemetry))
	// The runtime profiles, behind the Guard and the same admin check. A
	// trace names the slow request; a CPU profile names the function in
	// it, and only the running instance can produce one.
	r.Mount(PprofPath, a.Guard.Require(pprofRoutes()))
	if ported == nil {
		// No PHP front end in this build: the Go pages hold the root. A
		// build that cannot extract the tree still serves.
		r.Mount("/", pages.Routes())
		a.handler = svc.Telemetry.Middleware(r)
		return nil
	}
	// The Go pages, under a prefix and kept to be read against the ported
	// ones. GoBase says why a link from one of them leaves.
	//
	// StripPrefix as well as the mount, because chi's Mount takes the prefix
	// off the path it routes on and leaves the request's own URL whole, and a
	// page reads that URL to decide which sidebar row it is: no row would be
	// marked. The pages see themselves at the root, where the comparison
	// fetches them.
	r.Mount(GoBase, http.StripPrefix(GoBase, pages.Routes()))
	// What stays at the root because it is a page of neither front end: the
	// session, the readiness check, the theme's files, the byte streams the
	// ported pages link to, and a plugin's own pages. frontend's Root names
	// each of them and says why.
	pages.Root(r)
	r.Mount("/", ported)
	// The middleware wraps every route above, so one trace holds a request
	// and everything it ran: a handler's own spans, and the spans of a
	// module's PHP, which reports into the same tracer.
	a.handler = svc.Telemetry.Middleware(r)
	return nil
}
