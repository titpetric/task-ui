package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

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
	"github.com/titpetric/task-ui/internal/assets"
	"github.com/titpetric/task-ui/mcp"
	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/note"
	"github.com/titpetric/task-ui/phpfront"
	"github.com/titpetric/task-ui/phpscript"
	"github.com/titpetric/task-ui/phpscripts"
	"github.com/titpetric/task-ui/recording"
	"github.com/titpetric/task-ui/runner"
	"github.com/titpetric/task-ui/scheduler"
	"github.com/titpetric/task-ui/skills"
	"github.com/titpetric/task-ui/spotify"
	"github.com/titpetric/task-ui/sshprobe"
	"github.com/titpetric/task-ui/storage"
	"github.com/titpetric/task-ui/tooling"
	"github.com/titpetric/task-ui/workflow"
)

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
	tools := storage.NewToolStorage(db)
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
	modules := storage.NewModuleStorage(db)
	sidebar := storage.NewSidebarStorage(db)
	// The sidebar's top block as it stood before the table held it, on the
	// first start of an instance. A table with a row in it is left alone,
	// so the order a user saved survives a restart.
	if err := sidebar.EnsureDefaults(context.Background(), model.SidebarDefaults()); err != nil {
		return err
	}
	telemetry, err := NewTelemetry(context.Background(), storage.NewSettingStorage(db), "task-ui")
	if err != nil {
		return err
	}
	a.Telemetry = telemetry

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
	// The meter walks the artifact tree, the reader signs in to a host and
	// the tool prober runs atkins: both front ends hold the same one, so the
	// work is done once however a reader came to the page, and the two show
	// the same numbers.
	disk := &diskusage.Meter{TTL: 10 * time.Minute}
	facts := &hostinfo.Reader{Run: prober.Output, Timeout: 20 * time.Second, TTL: 5 * time.Minute}
	tooled := &tooling.Prober{Timeout: 5 * time.Second, TTL: time.Minute}
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
		Tooling:       toolSummary(tools),
	})
	hostRuns := &hostrun.Starter{Catalogue: a.Catalogue, Hosts: hosts, SSH: prober, Runner: a.Runner}
	a.Scheduler = scheduler.NewLoop(schedules, a.Users, a.Catalogue.Steps(), a.Runner)
	a.Scheduler.Hosts = hostRuns
	a.Workflows = workflow.New(workflows, workflowRuns, a.Catalogue.Steps(), a.Runner)
	a.Workflows.Hosts = hostRuns
	a.Workflows.Users = a.Users
	a.Workflows.Projects = a.Catalogue
	a.Scheduler.Flows = a.Workflows
	events := event.NewBus()
	a.Workflows.Events = events
	a.Runner.Events = events
	a.Guard = access.NewGuard(access.NewSessions(storage.NewUserSessionStorage(a.DB), a.Users,
		access.SessionPolicy{TTL: cfg.SessionTTL, Max: cfg.SessionMax}))
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
	dispatcher := NewDispatcher(a.Board, a.Users, store)
	moduleRuns := NewModuleRuns(storage.NewNoteStorage(db))

	moduleOpts := phpscript.NewOptions(cfg.ModulesDir())
	moduleOpts.Roots = append(moduleOpts.Roots, cfg.Root)
	moduleOpts.Registry = phpscript.NewStore(modules, sidebar)
	moduleOpts.Dispatcher = dispatcher
	moduleOpts.Runs = moduleRuns
	moduleOpts.Database = phpscript.NewDatabase(db)
	moduleOpts.Artifacts = phpscript.NewTree(store)
	moduleOpts.Tracer = telemetry.Tracer()
	moduleOpts.Version = version
	moduleOpts.OnError = func(err error) { slog.Warn("a module did not register", "err", err) }
	a.Modules = phpscript.New(moduleOpts)
	if err := a.Modules.Start(context.Background()); err != nil {
		return err
	}

	// What the server runs with and what it keeps on disk, both read by the
	// two front ends: the Instance page of either draws the same
	// configuration, and the meter is one so a page drawn twice walks the
	// disk once.
	instance := &view.Instance{
		URL: cfg.URL, Listen: cfg.Listen, Root: cfg.Root, DataDir: cfg.DataDir, SSHDir: cfg.SSHDir, Htpasswd: cfg.Htpasswd, HostData: hostDataShown(cfg.DataDir, hostData),
		Discover: cfg.Discover, AutoLogin: cfg.AutoLogin, TimeZone: timeZone(cfg.Clock()()),
		SessionTTL: cfg.SessionTTL.String(), SessionMax: cfg.SessionMax.String(), Keep: cfg.Keep.String(), Grace: cfg.Grace.String(),
		ArtifactHistory: cfg.ArtifactHistory, ComposeProject: self,
	}
	measured := &diskusage.Inventory{TTL: 30 * time.Second, Wait: 2 * time.Second, Items: storageItems(cfg, workspaces, os.Getenv("HOME"))}

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
	frontOpts.Database = phpscript.NewDatabase(db)
	frontOpts.Artifacts = phpscript.NewTree(store)
	frontOpts.Tracer = telemetry.Tracer()
	frontOpts.Host = &FrontHost{
		Catalogue: a.Catalogue, Runner: a.Runner, Workflows: a.Workflows, Git: g, Events: events,
		Recordings: recordings, Integrations: integrations, Secrets: secrets, Root: cfg.Root,
		Artifacts: store, Disk: disk, Now: cfg.Clock(),
		Hosts: hosts, SSH: prober, Facts: facts,
		Tools: tooled, Compose: cfg.ComposeProject,
		Keystore: keys, Config: instance,
		Tracing: telemetry, Plugins: a.Modules,
		Cleaner: cleaner, Storage: measured,
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
			return a.UsersFile.Add(ctx, a.Users, a.Passwords, name, password)
		}
		removeUser = func(ctx context.Context, name string) error {
			return a.UsersFile.Remove(ctx, a.Users, a.Passwords, name)
		}
	}
	pins := storage.NewPinStorage(db)
	spotifyService := &spotify.Service{
		Settings: func(ctx context.Context) (map[string]string, error) {
			return integrations.Values(ctx, integration.Spotify)
		},
		Tokens:  spotify.SecretTokens{Secrets: secrets},
		Client:  &spotify.Client{},
		OAuth:   &spotify.OAuth{},
		States:  &spotify.States{},
		BaseURL: cfg.URL,
	}
	apiHandlers := api.New(api.Options{
		Spotify: spotifyService,
		Users:   a.Users, Projects: projects, Workspaces: workspaces, Runs: runs, Artifacts: artifacts, ArtifactStore: store, Schedules: schedules, Secrets: secrets, SecretReads: secretReads,
		FolderRefs: refs,
		Workflows:  workflows, WorkflowRuns: workflowRuns, Flows: a.Workflows, Events: events,
		Runner: a.Runner, Catalogue: a.Catalogue, Steps: a.Catalogue.Steps(), Git: g, Keys: keys, Env: env, Recordings: recordings, Root: cfg.Root, Board: a.Board,
		Integrations: integrations, Hosts: hosts, HostRuns: hostRuns, Summarizer: &claude.Summarizer{Timeout: 3 * time.Minute},
		AddUser: addUser, RemoveUser: removeUser, Tokens: a.Guard.Tokens, Now: cfg.Clock(), Summaries: summaries, Version: version,
		Cleaner: cleaner, Pins: pins, Tools: tools,
		Sidebar: sidebar, Modules: a.Modules, Telemetry: telemetry,
	})
	pages, err := frontend.New(frontend.Options{
		Guard: a.Guard, Auth: access.NewAuthenticator(a.Users, a.Emails, a.Passwords), Users: a.Users, Projects: projects, Workspaces: workspaces, Runs: runs, Artifacts: artifacts, Schedules: schedules, Secrets: secrets, SecretReads: secretReads,
		FolderRefs: refs,
		Workflows:  workflows, WorkflowRuns: workflowRuns, Flows: a.Workflows,
		Runner: a.Runner, Catalogue: a.Catalogue, Steps: a.Catalogue.Steps(), Git: g, Keys: keys, Recordings: recordings, Root: cfg.Root, Version: version, Now: cfg.Clock(),
		ArtifactStore: store, Disk: disk, Board: a.Board, ComposeProject: self, Cleaner: cleaner,
		Storage:      measured,
		Integrations: integrations, Hosts: hosts, Pins: pins,
		Sidebar: sidebar, Modules: a.Modules, Telemetry: telemetry,
		Tooling: tooled,
		Tools:   tools,
		SSH:     prober, AddUsers: a.UsersFile != nil, BaseURL: cfg.URL,
		Facts:    facts,
		Instance: instance,
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
			return integrations.Values(ctx, integration.Forgejo)
		},
		Projects:   projects.List,
		Workspaces: workspaces.List,
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
	r.Mount(telemetry.Path(), a.Guard.Require(telemetry))
	// The runtime profiles, behind the Guard and the same admin check. A
	// trace names the slow request; a CPU profile names the function in
	// it, and only the running instance can produce one.
	r.Mount(PprofPath, a.Guard.Require(pprofRoutes()))
	if ported == nil {
		// No PHP front end in this build: the Go pages hold the root. A
		// build that cannot extract the tree still serves.
		r.Mount("/", pages.Routes())
		a.handler = telemetry.Middleware(r)
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
	a.handler = telemetry.Middleware(r)
	return nil
}
