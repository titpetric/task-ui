//go:build integration

package frontend_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// Settings, Integrations shows a form per integration: a secret only as
// set, never its value. Admins only.
func TestIntegrations(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	assert.NoError(t, f.opts.Integrations.Save(context.Background(), integration.Claude, map[string]string{"token": "sk-secret", "who": "admin"}))

	body := f.page(t, asAdmin, "/settings/integrations")
	assert.Contains(t, body, "Claude Code")
	assert.Contains(t, body, `name="values.who" value="admin" data-type="joined" checked>`)
	assert.Contains(t, body, `name="values.who" value="viewer" data-type="joined">`)
	assert.Contains(t, body, `data-peek="integration.claude" checked`)
	assert.Contains(t, body, `data-confirm-ok="Clear"`)
	assert.Contains(t, body, `<span class="badge ok">set</span>`)
	assert.NotContains(t, body, "sk-secret")
	assert.Contains(t, body, `data-api="/api/integrations/beszel"`)
	assert.Contains(t, body, `data-api="/api/integrations/forgejo"`)
	assert.Contains(t, body, `<option value="ssh" selected>ssh</option>`)
	assert.NotContains(t, body, `data-job="@forgejo-sync"`, "no sync while the integration is not set up")
	assert.Contains(t, body, "No sync ran yet")

	// Set up, the form gets Sync now and Dry run, and the line names the
	// last run of @forgejo-sync.
	assert.NoError(t, f.opts.Integrations.Save(context.Background(), integration.Forgejo, map[string]string{"url": "https://git.example", "token": "t0k3n", "clone": "https"}))
	run := &model.Run{Folder: model.Default, Job: "@forgejo-sync", UserName: "admin", State: model.StateOK}
	assert.NoError(t, f.opts.Runs.Create(context.Background(), run))
	body = f.page(t, asAdmin, "/settings/integrations")
	assert.Contains(t, body, `data-job="@forgejo-sync" data-title="internal / forgejo-sync"`)
	assert.Contains(t, body, `data-run-args='["--dry-run"]'`)
	assert.Contains(t, body, `<option value="https" selected>https</option>`)
	assert.Contains(t, body, `href="/runs/`+run.ID+`"`)
	assert.NotContains(t, body, "t0k3n")

	assert.Equal(t, http.StatusForbidden, f.get(t, asOps, "/settings/integrations").Status)
	assert.Contains(t, f.page(t, asAdmin, "/settings"), `href="/settings/integrations"`)
	assert.NotContains(t, f.page(t, asAdmin, "/settings/secrets"), "integration:claude")
}

// Inventory shows the Beszel hosts, or how to set them up, and the run
// panel for an admin.
func TestInventory(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	assert.Contains(t, f.page(t, asAdmin, "/inventory"), "The Beszel integration isn't set up yet")

	hub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/collections/system_stats/records":
			_, _ = w.Write([]byte(`{"items":[{"system":"a","stats":{"m":15.6,"mu":12.4,"d":200,"du":60}}]}`))
		case "/api/collections/containers/records":
			_, _ = w.Write([]byte(`{"items":[{"system":"a","name":"db","cpu":2,"memory":100}]}`))
		default:
			_, _ = w.Write([]byte(`{"items":[{"id":"a","name":"dev1","host":"10.10.10.11","status":"up","info":{"cpu":95,"mp":80,"dp":30,"u":7200,"t":4,"m":"EPYC","k":"6.12","v":"0.12"}},{"id":"b","name":"old","host":"10.10.10.12","status":"down","info":{"t":2}}]}`))
		}
	}))
	t.Cleanup(hub.Close)
	assert.NoError(t, f.opts.Integrations.Save(context.Background(), integration.Beszel, map[string]string{"url": hub.URL, "token": "t1"}))

	admin := f.page(t, asAdmin, "/inventory")
	assert.Contains(t, admin, `data-inventory-run`)
	assert.Contains(t, admin, `name="hosts" value="10.10.10.11" checked><span>debian@10.10.10.11</span>`)
	assert.Contains(t, admin, `data-command="uptime"`)
	assert.NotContains(t, admin, `name="summarize"`)
	assert.NoError(t, f.opts.Integrations.Save(context.Background(), integration.Claude, map[string]string{"token": "sk-1", "who": "admin"}))
	assert.Contains(t, f.page(t, asAdmin, "/inventory"), `name="summarize" value="1"`)
	body := f.page(t, asViewer, "/inventory")
	assert.NotContains(t, body, `data-inventory-run`)
	assert.Contains(t, body, "dev1")
	// The host's icon carries its state; there is no State column.
	assert.Contains(t, body, `<svg class="ic host-state up" role="img" aria-label="up">`)
	assert.NotContains(t, body, `<th>State</th>`)
	// The hosts refresh themselves, and say the time of their drawing.
	assert.Contains(t, body, `<section class="block" data-hosts-live>`)
	assert.Contains(t, body, `<p class="faint live-stamp">Updated 22:00</p>`)
	assert.Contains(t, body, `class="meter high"`)
	assert.Contains(t, body, `class="meter warm"`)
	assert.Contains(t, body, `<td class="center">4</td>`)
	assert.Contains(t, body, `<code title="ssh debian@10.10.10.11">debian@10.10.10.11</code>`)
	assert.Contains(t, body, `data-copy="ssh debian@10.10.10.11"`)
	assert.Contains(t, body, `<td class="key nowrap" title="10.10.10.11, up"><svg class="ic host-state up" role="img" aria-label="up"><use href="#i-server"/></svg> dev1</td>`)
	assert.Contains(t, body, "12.4 / 15.6</td>")
	assert.Contains(t, body, "60 / 200</td>")
	assert.NotContains(t, body, "EPYC")
	assert.Contains(t, body, `href="/inventory" class="active"`)
	assert.Equal(t, http.StatusNotFound, f.get(t, asViewer, "/hosts").Status)

	// The cluster over the table: a pool a resource of the hosts that
	// report, the host down greyed in the legend.
	assert.Contains(t, body, `Cluster <small>2 hosts, 1 up, <span class="bad">1 down</span></small>`)
	assert.Contains(t, body, `<span class="pool-figure">3.8 / 4 cores</span><span class="pool-pct">95%</span>`)
	assert.Contains(t, body, `<span class="seg c0" style="flex-grow: 100.000" title="dev1: 12.4 GiB of 15.6 GiB"><i style="width: 79.5%"></i></span>`)
	assert.Contains(t, body, `<span class="pool-figure">1 running</span>`)
	assert.Contains(t, body, `<li class="left"><i class="swatch"></i>old</li>`)
	assert.Contains(t, body, "1 host is down or sends no measures")
}

// The run panel shows the user's last Run, a row a host, linked to its
// run, with the batch in Runs; each run page of the batch lists the
// prompt threads about its first run, and names its host.
func TestInventory_LastBatch(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	hub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[{"id":"a","name":"dev1","host":"10.10.10.11","status":"up","info":{"t":4}}]}`))
	}))
	t.Cleanup(hub.Close)
	ctx := context.Background()
	assert.NoError(t, f.opts.Integrations.Save(ctx, integration.Beszel, map[string]string{"url": hub.URL, "token": "t1"}))
	assert.NotContains(t, f.page(t, asAdmin, "/inventory"), "Your last run")

	batch := model.NewID()
	var runs []*model.Run
	for _, host := range []string{"dev1", "dev2"} {
		now := time.Now()
		run := &model.Run{Folder: model.Default, Job: model.HostJob(host), Target: host, BatchID: batch, Args: `["uptime"]`, UserID: f.users[asAdmin].ID, UserName: asAdmin, State: model.StateOK, StartedAt: &now, FinishedAt: &now, CastPath: "x.cast"}
		assert.NoError(t, f.opts.Runs.Create(ctx, run))
		runs = append(runs, run)
	}
	thread := &model.Note{Channel: model.ChannelNotes, UserID: f.users[asAdmin].ID, UserName: asAdmin, Body: "Summarize the runs on these hosts.", State: model.NoteDone, WorkspaceID: model.Default, Dir: model.Default, AboutRunID: runs[0].ID}
	assert.NoError(t, f.opts.Board.Notes.Create(ctx, thread))

	body := f.page(t, asAdmin, "/inventory")
	assert.Contains(t, body, "Your last run <small><code>uptime</code> on 2 hosts")
	assert.Contains(t, body, `<a class="btn btn-sm" href="/runs?batch=`+batch+`">In Runs</a>`)
	assert.Contains(t, body, `<a href="/runs/`+runs[1].ID+`"><code>dev2</code></a>`)
	assert.Contains(t, body, `data-output="`+runs[0].ID+`"`)
	assert.NotContains(t, body, "data-batch-live")
	assert.Contains(t, body, "Summarize the runs on these hosts.")
	assert.NotContains(t, f.page(t, asViewer, "/inventory"), "Your last run")

	page := f.page(t, asAdmin, "/runs/"+runs[1].ID)
	assert.Contains(t, page, "<h2>@host:dev2")
	assert.Contains(t, page, "Summarize the runs on these hosts.")
	listed := f.page(t, asAdmin, "/runs?batch="+batch)
	assert.Contains(t, listed, runs[0].ID)
	assert.Contains(t, listed, runs[1].ID)
}

// The job pickers of Scheduled and of a workflow's steps list the host
// jobs under Hosts, @host:* and one a host with an ssh login, for a
// user with the hosts permission alone; every job has the args field.
func TestHostJobs(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	hub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[{"id":"a","name":"dev1","host":"10.10.10.11","status":"up","info":{"t":4}}]}`))
	}))
	t.Cleanup(hub.Close)
	assert.NoError(t, f.opts.Integrations.Save(context.Background(), integration.Beszel, map[string]string{"url": hub.URL, "token": "t1"}))

	for _, path := range []string{"/schedules", "/workflows/new"} {
		admin := f.page(t, asAdmin, path)
		assert.Contains(t, admin, `<optgroup label="Hosts">`, path)
		assert.Contains(t, admin, `<option value="@host:*" data-folder="internal"`, path)
		assert.Contains(t, admin, `<option value="@host:10.10.10.11" data-folder="internal"`, path)
		assert.Contains(t, admin, `<option value="internal"`, path)
		ops := f.page(t, asOps, path)
		assert.NotContains(t, ops, `label="Hosts"`, path)
		assert.NotContains(t, ops, `@host:`, path)
	}
	assert.Contains(t, f.page(t, asAdmin, "/schedules"), `name="args" data-type="args"`)
	assert.Contains(t, f.page(t, asOps, "/workflows/new"), `data-step-args`)
}

// A queued run of a batch says it waits, on the inventory panel and on
// its page, with Attach; a failed one says why.
func TestInventory_Queued(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	f.withKey(t)
	hub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[{"id":"a","name":"dev1","host":"10.10.10.11","status":"up","info":{"t":4}}]}`))
	}))
	t.Cleanup(hub.Close)
	ctx := context.Background()
	assert.NoError(t, f.opts.Integrations.Save(ctx, integration.Beszel, map[string]string{"url": hub.URL, "token": "t1"}))
	now := time.Now()
	run := &model.Run{Folder: model.Default, Job: model.HostJob("dev1"), Target: "dev1", BatchID: model.NewID(), Args: `["uptime"]`, UserID: f.users[asAdmin].ID, UserName: asAdmin, State: model.StateQueued, StartedAt: &now}
	assert.NoError(t, f.opts.Runs.Create(ctx, run))

	panel := f.page(t, asAdmin, "/inventory")
	assert.Contains(t, panel, "Queued: waits for the run on dev1 before it.")
	assert.Contains(t, panel, `<i class="dot queued"></i>queued</span>`)
	assert.Contains(t, panel, "data-batch-live")
	page := f.page(t, asAdmin, "/runs/"+run.ID)
	assert.Contains(t, page, "It waits, queued, for the run of its job before it to end")
	assert.Contains(t, page, ">Attach</button>")
	assert.Contains(t, f.page(t, asAdmin, "/runs?state=queued"), run.ID)

	failed := &model.Run{Folder: model.Default, Job: model.HostJob("dev1"), Target: "dev1", UserName: asAdmin, State: model.StateFailed, Message: model.QueuedLost, StartedAt: &now, FinishedAt: &now}
	assert.NoError(t, f.opts.Runs.Create(ctx, failed))
	assert.Contains(t, f.page(t, asAdmin, "/runs/"+failed.ID), `<span class="dim">the instance restarted before it ran</span></td></tr>`)
}
