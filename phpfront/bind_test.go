package phpfront_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/titpetric/oida"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpfront"
)

// host answers what the real one answers off the instance.
type host struct {
	published  []string
	dispatched []string
}

func (h *host) Whoami(context.Context) string { return "ops" }

func (h *host) Folders(context.Context) []phpfront.Folder {
	return []phpfront.Folder{{
		ID:    "infra",
		Title: "Infra",
		Dir:   "/src/infra",
		Allow: []string{"ops"},
		Jobs:  []phpfront.Job{{ID: "build", Name: "build"}, {ID: "task:test", Name: "test", Allow: []string{"ann"}}},
	}}
}

// The job of a run is the folder, the directory inside it and the job, all
// three: the shell of a directory inside the folder is a different job from
// the shell of the folder, and only this answers either.
func (h *host) Job(_ context.Context, folder, dir, id string) phpfront.Job {
	if folder != "infra" || dir != "api" || id != "@shell" {
		return phpfront.Job{}
	}
	return phpfront.Job{
		Found:       true,
		ID:          "@shell",
		Name:        "shell",
		Dir:         "api",
		Shell:       true,
		Allow:       []string{"admin"},
		FolderAllow: []string{"ops"},
	}
}

func (h *host) Projects(_ context.Context, workspace string) []string {
	if workspace != "malta" {
		return nil
	}
	return []string{"malta/deploy"}
}

func (h *host) FolderDir(_ context.Context, id string, workspaces []string) string {
	if id == "infra" {
		return "/src/infra"
	}
	if len(workspaces) == 0 {
		return ""
	}
	return "/" + workspaces[0] + "/src/" + id
}

func (h *host) IsDir(_ context.Context, id, rel string) bool {
	return id == "infra" && rel == "api"
}

func (h *host) Commit(_ context.Context, dir string) string {
	if dir != "/src/infra" {
		return ""
	}
	return "2024-05-01 10:00:00 +0000 UTC"
}

func (h *host) Running(context.Context) []phpfront.Running {
	return []phpfront.Running{{Key: "infra/build", ID: "01hz", State: "running"}}
}

func (h *host) Flows(context.Context) []phpfront.Flow {
	return []phpfront.Flow{{WorkflowID: "01hw", RunID: "01hr"}}
}

func (h *host) Dirty(_ context.Context, dir string) int {
	if dir == "/src/infra" {
		return 3
	}
	return 0
}

func (h *host) Rules(context.Context) phpfront.Rules {
	return phpfront.Rules{
		Rules:       []phpfront.Rule{{Key: "view", Title: "Allow View"}},
		Sections:    []phpfront.Rule{{Key: "note", Title: "Notes"}},
		Permissions: []phpfront.Rule{{Key: "run", Title: "Run Jobs"}},
		Paths:       []string{"note/all", "note/view", "run"},
		Checks:      []string{"note/view", "run"},
		Extras:      []phpfront.Rule{{Key: "note/pin", Title: "Pin"}},
	}
}

func (h *host) Publish(_ context.Context, topic string, _ map[string]any) error {
	h.published = append(h.published, topic)
	return nil
}

func (h *host) Session(_ context.Context, id string) phpfront.Session {
	if id != "01hz" {
		return phpfront.Session{}
	}
	return phpfront.Session{
		Live:     true,
		Attached: 2,
		Message:  "waits for a free slot, number 1 in line",
		State:    "running",
		Title:    "build the image",
	}
}

func (h *host) Frame(_ context.Context, id string) string {
	if id != "01hz" {
		return ""
	}
	return "<svg></svg>"
}

func (h *host) Screen(_ context.Context, id, castPath string, cols int) phpfront.Screen {
	if castPath != "run.cast" {
		return phpfront.Screen{}
	}
	return phpfront.Screen{SVG: "<svg></svg>", Lines: 42, Complete: true}
}

func (h *host) Expired(_ context.Context, castPath, state string) bool {
	return castPath == "gone.cast" && state == "ok"
}

func (h *host) Claude(context.Context) bool { return true }

func (h *host) Tree(_ context.Context, dir string) []phpfront.TreeItem {
	if dir != "lint" {
		return nil
	}
	return []phpfront.TreeItem{
		{Name: "nested", Dir: true, Size: 4096, ModTime: "2026-11-02 10:32:12.414 +0000 UTC"},
		{Name: "out.txt", Size: 31},
	}
}

func (h *host) Space(context.Context) phpfront.Space {
	return phpfront.Space{Measured: true, Files: 2048, History: 512, Ignored: 64}
}

// A directory nothing measured is not in the answer, which is what an
// instance with no meter answers for every one of them.
func (h *host) Spaces(_ context.Context, dirs []string) map[string]phpfront.Space {
	out := map[string]phpfront.Space{}
	for _, dir := range dirs {
		if dir != "/root/src/infra" {
			continue
		}
		out[dir] = phpfront.Space{Measured: true, Files: 1024, History: 256, Ignored: 32}
	}
	return out
}

// The hosts permission decides whether task-ui reads a host now, so the
// double answers a login only where the reader has it.
func (h *host) Inventory(_ context.Context, hosts bool) phpfront.Inventory {
	out := phpfront.Inventory{
		Hosts: []phpfront.System{{
			ID: "sys1", Name: "web1", Host: "10.0.0.1", Status: "up", Up: true,
			CPUs: 8, Agent: "0.9.1", Uptime: "3 days",
			CPU: 12.5, Memory: 48, Disk: 71,
			Measured: true, MemoryTotal: 32, MemoryUsed: 15.5, DiskTotal: 500, DiskUsed: 355,
		}},
		Cluster: phpfront.Cluster{
			Up: 1, Pools: []phpfront.Pool{{
				Name: "CPU", Unit: "cores", Figure: "1 / 8 cores", Used: 1, Total: 8, Percent: 12.5,
				Segments: []phpfront.Segment{{Host: "web1", Width: 100, Fill: 12.5, Title: "web1: 1 of 8 cores"}},
			}},
			Hosts: []phpfront.ClusterHost{{Name: "web1"}},
		},
		History: []phpfront.Series{{ID: "sys1", Samples: []phpfront.Sample{{At: "2026-11-02 10:00:00 +0000 UTC", CPU: 9, Sent: 1024}}}},
	}
	if hosts {
		out.Logins = []phpfront.Login{{Host: "10.0.0.1", Login: "debian@10.0.0.1"}}
	}
	return out
}

func (h *host) Reachable(context.Context) []string {
	return []string{"10.0.0.1"}
}

func (h *host) HostFacts(_ context.Context, name string, hosts bool) phpfront.HostFacts {
	if name != "web1" {
		return phpfront.HostFacts{}
	}
	out := phpfront.HostFacts{
		Found:   true,
		System:  phpfront.System{ID: "sys1", Name: "web1", Host: "10.0.0.1", Up: true, CPUs: 8},
		Unsized: "task-ui has no ssh login to the host, for docker ps",
		Containers: []phpfront.Container{{
			Name: "api", Image: "api:1", State: "running", Status: "Up 2 days",
			CPU: 3.5, Memory: 512, MemoryShare: 1.5, Net: 2048, Measured: true,
		}},
		Measured:    true,
		Filesystems: []phpfront.Filesystem{{Mount: "/", Size: 1 << 30, Used: 1 << 29, Free: 1 << 29, Percent: 50}},
	}
	if hosts {
		out.Login = "debian@10.0.0.1"
		out.Read = "2026-11-02 10:32:12.414 +0000 UTC"
	}
	return out
}

func (h *host) Tooling(context.Context) []phpfront.ToolingScope {
	return []phpfront.ToolingScope{{
		Name: "task-ui", Repository: "task-ui", Database: "/data/atkins.db",
		Days: 30, From: "2026-10-04", To: "2026-11-02", Runs: 3, TotalMs: 4500,
		Commands: []phpfront.ToolingCommand{{
			Command: "go test", Tool: "go", Runs: 3, TotalMs: 4500,
			AverageMs: 1500, MaxMs: 2500, Failures: 1,
			LastRunAt: "2026-11-02 10:32:12.414 +0000 UTC",
		}},
		Usage:        []phpfront.ToolingUsage{{Date: "2026-11-02", Tool: "go", Runs: 3, TotalMs: 4500}},
		CommandUsage: []phpfront.ToolingUsage{{Date: "2026-11-02", Command: "go test", Tool: "go", Runs: 3, TotalMs: 4500}},
		Samples: []phpfront.ToolingSample{{
			At: "2026-11-02 10:32:12.414 +0000 UTC", Command: "go test", Tool: "go",
			Job: "test", DurationMs: 2500, ExitCode: 1, Failed: true,
		}},
	}}
}

// The folder pages read the catalogue and the disk, which this fake has none
// of: each answers the nothing an instance with no such folder answers.
func (h *host) Place(context.Context, string, string) phpfront.Place {
	return phpfront.Place{}
}

func (h *host) Groups(context.Context, string, string, []string, bool, bool) []phpfront.Group {
	return nil
}

func (h *host) Listing(context.Context, string, string, []string) []phpfront.Entry {
	return nil
}

func (h *host) RootTree(context.Context, bool, []string) []phpfront.Entry {
	return nil
}

func (h *host) ModTime(context.Context, string) string {
	return ""
}

func (h *host) Status(context.Context, string) phpfront.Status {
	return phpfront.Status{}
}

func (h *host) Checkout(context.Context, string) phpfront.Checkout {
	return phpfront.Checkout{}
}

func (h *host) Hours(context.Context, string) phpfront.Hours {
	return phpfront.Hours{}
}

func (h *host) Keys(context.Context) phpfront.Keystore {
	return phpfront.Keystore{}
}

func (h *host) KeyFolders(context.Context) []phpfront.KeyFolder {
	return nil
}

func (h *host) Instance(context.Context, bool) phpfront.Instance {
	return phpfront.Instance{}
}

func (h *host) Usage(context.Context) phpfront.Retention {
	return phpfront.Retention{}
}

func (h *host) Telemetry(context.Context) phpfront.Telemetry {
	return phpfront.Telemetry{}
}

// A service answers its registry entry and what its settings hold; the token
// it is set up with is not in the answer at all.
func (h *host) Services(context.Context) []phpfront.Integration {
	return []phpfront.Integration{{
		ID: "claude", Title: "Claude Code", Description: "The token Claude Code signs in with.",
		Credential: "OAuth token", Steps: []string{"Run claude setup-token."},
		Fields: []phpfront.IntegrationField{
			{Name: "token", Label: "OAuth token", Secret: true, Env: "CLAUDE_CODE_OAUTH_TOKEN"},
			{Name: "who", Label: "Who", Optional: true, Who: true},
		},
		Configured: true,
		Values:     map[string]string{"who": "admin, ops"},
		Set:        map[string]bool{"token": true},
		Checked:    []string{"admin", "ops"},
		Enabled:    map[string]bool{},
	}}
}

// The vault answers the variables it opened and the first one it would not,
// which is the shape the page draws: the table and the error above it.
func (h *host) Variables(context.Context) phpfront.Variables {
	return phpfront.Variables{Values: map[string]string{"TOKEN": "", "REGION": "eu"}}
}

// One variable as its form opens it. A secret draws no value, so the row the
// form edits answers content only where the variable holds some.
func (h *host) Variable(_ context.Context, id string) phpfront.Variable {
	if id != "REGION" {
		return phpfront.Variable{}
	}
	return phpfront.Variable{Content: "eu"}
}

// The environment the integrations add is the reader's: a plain variable and a
// credential, which answers its name and no value.
func (h *host) IntegrationEnv(context.Context) []phpfront.EnvVar {
	return []phpfront.EnvVar{
		{Name: "CLAUDE_CODE_OAUTH_TOKEN", Redacted: true},
		{Name: "CLAUDE_MODEL", Value: "opus"},
	}
}

func (h *host) Modules(context.Context) []phpfront.Plugin {
	return nil
}

func (h *host) Entries(context.Context, []string) []phpfront.Project {
	return nil
}

func (h *host) Dispatch(_ context.Context, comment, action, folder string) (string, error) {
	h.dispatched = append(h.dispatched, action+":"+comment+":"+folder)
	return "01NOTE", nil
}

func (h *host) RunOf(_ context.Context, note string) string {
	if note == "" {
		return ""
	}
	return "01RUN-" + note
}

// Every host call answers across the boundary, and a Go struct arrives as a
// map keyed by its json tags.
func TestBind(t *testing.T) {
	t.Parallel()

	own := &host{}
	body := serve(t, own, "/host")

	// A Go struct arrives as an associative array keyed by its json tags, in
	// declaration order, which is what a template indexes by.
	for _, want := range []string{
		"base=/development",
		"version=test",
		"whoami=ops",
		"dirty=3",
		`folders=[{"id":"infra","title":"Infra","dir":"/src/infra","allow":["ops"],"sources":[],"error":"","errors":[],"jobs":[{"id":"build","name":"build","dir":"","shell":false,"host":false,"allow":[]},{"id":"task:test","name":"test","dir":"","shell":false,"host":false,"allow":["ann"]}]}]`,
		// The gate of a run takes both allow lists and the kind, and the
		// catalogue cannot answer for a directory inside a folder, so the
		// job of a run is its own call and answers the whole gate.
		`job={"found":true,"id":"@shell","name":"shell","dir":"api","shell":true,"host":false,"allow":["admin"],"folder_allow":["ops"],"kind":"","source":"","source_lines":0,"description":"","summary":"","command":[],"steps":[],"aliases":[],"depends":[],"interactive":false,"takes_args":false,"args_command":"","key":""}`,
		`nojob={"found":false,`,
		`running=[{"key":"infra/build","id":"01hz","state":"running"}]`,
		// What the workflow runner holds, which is what running means for a
		// workflow: the table says so of a run a dead process left behind.
		`flows=[{"workflow_id":"01hw","run_id":"01hr"}]`,
		`"checks":["note/view","run"]`,
		`"sections":[{"key":"note","title":"Notes"}]`,
		// A rule a section added is keyed by the path a check names, so the
		// matrix draws its row without composing the path itself.
		`"extras":[{"key":"note/pin","title":"Pin"}]`,
		"published=true",
		// Indexing by key is what a template and the DAOs do, and what a
		// host object silently answers nothing for.
		"indexed=true",
		`keys=["rules","sections","permissions","paths","checks","extras"]`,
		"folder=infra allow=ops jobs=build/build,task:test/test",
		"section=note:Notes",
		"checks=note/view,run",
		// A session is the run the runner holds, which a page lays over the
		// row. An ID the runner holds nothing for is not live, and then the
		// row is the whole answer.
		`session={"live":true,"attached":2,"message":"waits for a free slot, number 1 in line","state":"running","exit_code":0,"title":"build the image","finished_at":""}`,
		`nosession={"live":false,`,
		"frame=<svg></svg>",
		"expired=true",
		"kept=false",
		"claude=true",
		`tree=[{"name":"nested","dir":true,"size":4096,"mod_time":"2026-11-02 10:32:12.414 +0000 UTC"},{"name":"out.txt","dir":false,"size":31,"mod_time":""}]`,
		"notree=[]",
		`space={"measured":true,"files":2048,"history":512,"ignored":64}`,
		`spaces={"/root/src/infra":{"measured":true,"files":1024,"history":256,"ignored":32}}`,
		// A service crosses with its registry entry and the state of its
		// settings, and with no secret in it: `set` says the token has a
		// value and nothing says what it is.
		`integrations=[{"id":"claude","title":"Claude Code","description":"The token Claude Code signs in with.","credential":"OAuth token","steps":["Run claude setup-token."],"fields":[{"name":"token","label":"OAuth token","help":"","placeholder":"","secret":true,"optional":false,"who":false,"options":[],"generated":false,"env":"CLAUDE_CODE_OAUTH_TOKEN","default":"","toggle":false,"default_on":false,"empty":false,"source":""},{"name":"who","label":"Who","help":"","placeholder":"","secret":false,"optional":true,"who":true,"options":[],"generated":false,"env":"","default":"","toggle":false,"default_on":false,"empty":false,"source":""}],"configured":true,"values":{"who":"admin, ops"},"set":{"token":true},"checked":["admin","ops"],"enabled":[],"error":""}]`,
		`"hosts":[{"id":"sys1","name":"web1","host":"10.0.0.1","status":"up","up":true,"cpus":8,"agent":"0.9.1","uptime":"3 days","cpu":12.5,"memory":48,"disk":71,"measured":true,"memory_total":32,"memory_used":15.5,"disk_total":500,"disk_used":355}]`,
		`"logins":{"10.0.0.1":"debian@10.0.0.1"}`,
		`"pools":[{"name":"CPU","unit":"cores","figure":"1 / 8 cores","used":1,"total":8,"percent":12.5,"segments":[{"host":"web1","color":0,"width":100,"fill":12.5,"title":"web1: 1 of 8 cores"}]}]`,
		`"history":{"sys1":[{"at":"2026-11-02 10:00:00 +0000 UTC","cpu":9,"memory":0,"memory_used":0,"disk":0,"disk_used":0,"sent":1024,"recv":0,"extra":[]}]}`,
		// Without the hosts permission the page still draws, and nothing
		// reads a host: an empty array, which keyed is an empty object.
		"watching=[]",
		`"containers":[{"name":"api","image":"api:1","state":"running","status":"Up 2 days","cpu":3.5,"memory":512,"memory_share":1.5,"net":2048,"size_rw":0,"size_virtual":0,"measured":true,"sized":false}]`,
		`"filesystems":[{"mount":"/","device":"","type":"","size":1073741824,"used":536870912,"free":536870912,"percent":50}]`,
		`nohost={"found":false,`,
		"login=debian@10.0.0.1",
		"records=1",
	} {
		assert.Contains(t, body, want)
	}
	assert.Contains(t, body, `<h1 id="title">Title</h1>`, "markdown renders through the instance's renderer")
	assert.Contains(t, body, "plain=<p>see <tt>notes</tt>", "no prefix leaves a relative link as its text")
	assert.Contains(t, body, `linked=<p>see <a href="/artifacts/files/lint/out.txt">`, "a prefix links a relative path under the directory it was written in")
	assert.Contains(t, body, "run=01RUN-01NOTE|\n", "no note is a run that has not started")
	assert.Equal(t, []string{"runs"}, own.published)
}

// A dispatch answers the note it opened and carries the action file and the
// folder through verbatim: the folder is the thread's scratch space in the
// artifact tree and the action is a path in this tree.
func TestBindDispatch(t *testing.T) {
	t.Parallel()

	own := &host{}
	body := serve(t, own, "/dispatch")

	assert.Contains(t, body, "dispatched=01NOTE")
	assert.Equal(t, []string{"modules/refinement/action/refinement.md:use ULID keys:01THREAD"}, own.dispatched)
}

// A nil Host answers every call empty, which is an instance with no chrome and
// no panic.
func TestBind_NoHost(t *testing.T) {
	t.Parallel()

	body := serve(t, nil, "/host")
	assert.Contains(t, body, "whoami=\n")
	assert.Contains(t, body, "dirty=0")
	assert.Contains(t, body, "folders=[]", "an empty array, so a foreach over it draws the empty line")
	assert.Contains(t, body, `nojob={"found":false,`, "no catalogue is a job nobody may run")
	assert.Contains(t, body, `rules={"rules":[],"sections":[],"permissions":[],"paths":[],"checks":[],"extras":[]}`)
	assert.Contains(t, body, "published=false")
	assert.Contains(t, body, `session={"live":false,`, "a session nothing holds, so a page reads the row alone")
	assert.Contains(t, body, "frame=\n")
	assert.Contains(t, body, "expired=false")
	assert.Contains(t, body, "claude=false")
	assert.Contains(t, body, "tree=[]", "an empty array, so a foreach over a directory draws the empty line")
	assert.Contains(t, body, `space={"measured":false,`, "no meter says nothing about space, rather than that it takes none")
	assert.Contains(t, body, `"off":true`, "no inventory is the page's unset state, not an error")
	assert.Contains(t, body, `nohost={"found":false,`)
	assert.Contains(t, body, "run=|\n", "no lookup is a run that has not started, twice over")
}

// Dispatch is the exception to the nil Host answering empty: an instance that
// starts no runs refuses the call, so a page cannot record a note that was
// never opened. The status is the failure of the request, which is why this
// does not go through serve.
func TestBindDispatch_NoHost(t *testing.T) {
	t.Parallel()

	opts := phpfront.NewOptions("testdata/host")
	opts.OnError = func(err error) { t.Errorf("router: %v", err) }

	front, err := phpfront.New(opts)
	assert.NoError(t, err)

	router, err := front.Routes(context.Background())
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/dispatch", nil))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// A page's trace holds the PHP that drew it.
//
// The interpreter measures a frame, a new, an include and a template, and it
// does so only while the runtime carries an observer: the front end registers
// the one phpscript's telemetry gives it, and without that a page's trace held
// the queries the database stdlib opens and not a line of the PHP above them.
func TestFrontSpans(t *testing.T) {
	t.Parallel()

	opts := oida.NewOptions("task-ui-test")
	opts.ReadEnv = false
	opts.Enabled = true
	tracer, err := oida.New(opts)
	assert.NoError(t, err)

	assert.Contains(t, serveTraced(t, &host{}, "/spans", tracer), "drawn=ops")

	var names []string
	for _, trace := range tracer.Snapshot().Log {
		for _, span := range trace.Spans {
			names = append(names, span.Name)
		}
	}

	assert.Contains(t, names, "new Page", "the allocation of a class is a span")
	assert.Contains(t, names, "page.draw", "and a frame, named after the receiver the call was written on")
}

// serve runs one request against the host fixture tree.
func serve(t *testing.T, own phpfront.Host, path string) string {
	t.Helper()
	return serveTraced(t, own, path, nil)
}

// serveTraced is serve, recording into a tracer. A nil tracer records nothing,
// which is an instance with telemetry switched off.
func serveTraced(t *testing.T, own phpfront.Host, path string, tracer *oida.Tracer) string {
	t.Helper()

	opts := phpfront.NewOptions("testdata/host")
	opts.Base = "/development"
	opts.Version = "test"
	opts.Host = own
	opts.Tracer = tracer
	opts.OnError = func(err error) { t.Errorf("router: %v", err) }

	front, err := phpfront.New(opts)
	assert.NoError(t, err)

	router, err := front.Routes(context.Background())
	assert.NoError(t, err)

	// The trace is the request's, so the middleware is what puts one on the
	// context the page and every binding under it read.
	var handler http.Handler = router
	if tracer != nil {
		handler = tracer.Middleware(router)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	assert.Equal(t, http.StatusOK, w.Code, w.Body.String())

	return strings.ReplaceAll(w.Body.String(), "\r\n", "\n")
}
