package phpfront_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/oida"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpfront"
)

// host answers what the real one answers off the instance.
//
// Its methods are grouped the way the interface is: a folder's and the
// artifact tree's in host_folder_test.go, the live runs' in host_run_test.go,
// the Inventory's in host_inventory_test.go, the tooling report's in
// host_tooling_test.go and the settings' in host_settings_test.go.
type host struct {
	published  []string
	dispatched []string
}

func (h *host) Whoami(context.Context) string { return "ops" }

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
		"folderDir=/src/infra",
		// A tree asks about every row it draws in one call, and reads the
		// answer by ID: an empty ID is asked of nobody, so the list the
		// chrome hands over is not filtered on this side first.
		"folderDirs=2 /src/infra /malta/src/deploy -",
		// A service crosses with its registry entry and the state of its
		// settings, and with no secret in it: `set` says the token has a
		// value and nothing says what it is.
		`integrations=[{"id":"claude","title":"Claude Code","description":"The token Claude Code signs in with.","credential":"OAuth token","steps":["Run claude setup-token."],"fields":[{"name":"token","label":"OAuth token","help":"","placeholder":"","secret":true,"optional":false,"who":false,"options":[],"generated":false,"env":"CLAUDE_CODE_OAUTH_TOKEN","default":"","toggle":false,"default_on":false,"empty":false,"source":"","fixed":false},{"name":"who","label":"Who","help":"","placeholder":"","secret":false,"optional":true,"who":true,"options":[],"generated":false,"env":"","default":"","toggle":false,"default_on":false,"empty":false,"source":"","fixed":false}],"configured":true,"values":{"who":"admin, ops"},"set":{"token":true},"checked":["admin","ops"],"enabled":[],"error":""}]`,
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
	assert.Contains(t, body, "folderDirs=0 - - -", "no root places any folder, and a tree reads every row as having none")
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
	// Every interpreter span is kept. The front end runs at trace_level
	// info, which marks a frame as detail, and the instance's own minimum
	// of a millisecond would drop a frame of this fixture: what is asserted
	// here is what the interpreter reports, not how fast the machine is.
	opts.DebugSpanMinDuration = 0
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

// A frame too short to read is dropped and counted.
//
// This is the division of labour the trace level is for: the interpreter
// reports a span per frame and marks it as detail, and the recorder's minimum
// decides which of them a reader sees. A page that calls two hundred methods
// is two hundred spans of a microsecond without it, with the request they
// belong to somewhere underneath.
func TestFrontSpansDropped(t *testing.T) {
	t.Parallel()

	opts := oida.NewOptions("task-ui-test")
	opts.ReadEnv = false
	opts.Enabled = true
	// A second: no frame of this fixture runs that long, so every one of
	// them is detail the trace drops.
	opts.DebugSpanMinDuration = time.Second
	tracer, err := oida.New(opts)
	assert.NoError(t, err)

	assert.Contains(t, serveTraced(t, &host{}, "/spans", tracer), "drawn=ops")

	var names []string
	var dropped int
	for _, trace := range tracer.Snapshot().Log {
		dropped += trace.DroppedSpans
		for _, span := range trace.Spans {
			names = append(names, span.Name)
		}
	}

	assert.NotContains(t, names, "new Page", "a frame under the minimum is detail the trace drops")
	assert.NotContains(t, names, "page.draw")
	assert.True(t, dropped >= 2, "and the trace counts what it dropped")
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
