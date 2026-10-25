{{ define "content" }}
{{- $admin := and .User .User.CanAdmin }}
<div class="page-head">
	<div>
		<h2>Instance</h2>
		<p>What the server runs with. The environment sets most of these: change one there and restart task-ui. The backlog length and the parent environment are settings of their own, saved here{{ if not $admin }} by an admin{{ end }}.</p>
	</div>
</div>

{{- with .Instance }}
<section class="block instance-section" id="general">
	<h3>General</h3>
	<p class="section-intro">Where the instance answers, who signs in, and the clock it reads.</p>
	{{ template "instance-table-head" }}
		<tr><td class="key">Base link</td><td class="value mono" data-label="Value">{{ $.BaseURL }}</td><td class="var" data-label="Variable"><code>TASK_UI_URL</code></td><td class="desc" data-label="About"><span>{{ if .URL }}The link pages and MCP clients use.{{ else }}Empty: task-ui reads the link from each request.{{ end }}</span></td></tr>
		<tr><td class="key">Version</td><td class="value mono ro" data-label="Value">{{ $.Version }}</td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>The build that serves this page.</span></td></tr>
		<tr><td class="key">Listen</td><td class="value mono" data-label="Value">{{ .Listen }}</td><td class="var" data-label="Variable"><code>TASK_UI_LISTEN</code></td><td class="desc" data-label="About"><span>The address the server listens on.</span></td></tr>
		<tr><td class="key">Time zone</td><td class="value mono ro" data-label="Value">{{ with .TimeZone }}{{ . }}{{ else }}<span class="faint">local</span>{{ end }}</td><td class="var" data-label="Variable"><code>TZ</code></td><td class="desc" data-label="About"><span>The clock of the server, which the hours of a folder and the schedules read.</span></td></tr>
		<tr><td class="key">Sign in</td><td class="value" data-label="Value">{{ if .AutoLogin }}<span class="badge bad">auto login</span> {{ template "user" (list $ .AutoLogin) }}{{ else if $.LoginRequired }}<span class="badge ok">required</span>{{ else }}<span class="badge">open</span>{{ end }}</td><td class="var" data-label="Variable"><code>TASK_UI_AUTOLOGIN</code></td><td class="desc" data-label="About"><span>Auto login signs everyone in as one user; for demos only.</span></td></tr>
		<tr><td class="key">Session lifetime</td><td class="value mono" data-label="Value">{{ with .SessionTTL }}{{ . }}{{ else }}<span class="faint">default</span>{{ end }}</td><td class="var" data-label="Variable"><code>TASK_UI_SESSION_TTL</code></td><td class="desc" data-label="About"><span>How long a browser stays signed in.</span></td></tr>
		<tr><td class="key">Password file</td><td class="value mono" data-label="Value">{{ with .Htpasswd }}{{ . }}{{ else }}<span class="faint">none</span>{{ end }}</td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>Who can sign in: <code>.htpasswd</code> in the data directory.</span></td></tr>
		<tr><td class="key">Roles</td><td class="value" data-label="Value"><a href="/settings/users">Users</a></td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>The admin flag, the role and MCP of each user live on their profile, and <code>task-ui user</code> sets them.</span></td></tr>
		<tr><td class="key">MCP</td><td class="value mono" data-label="Value">{{ $.BaseURL }}/api/mcp</td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>Where MCP clients connect, with a user's token.</span></td></tr>
	{{ template "instance-table-foot" }}
</section>

<section class="block instance-section" id="runs">
	<h3>Runs</h3>
	<p class="section-intro">How many runs do work at once, and how many may wait for a free slot. Terminals take no slot.</p>
	{{ template "instance-table-head" }}
		{{- with .Backlog }}
		<tr><td class="key">Parallel runs</td><td class="value mono ro" data-label="Value">{{ .Parallel }}<small class="faint">{{ .Going }} going, {{ .Waiting }} waiting</small></td><td class="var" data-label="Variable"><span class="faint">computed</span></td><td class="desc" data-label="About"><span>One run per CPU of the machine; a run over that waits, queued, in order.</span></td></tr>
		<tr><td class="key">Automatic backlog</td><td class="value mono ro" data-label="Value">{{ .Auto }}</td><td class="var" data-label="Variable"><code>backlog_length_auto</code></td><td class="desc" data-label="About"><span>20 runs per CPU: the length of the line while an admin sets no backlog length.</span></td></tr>
		<tr><td class="key">Backlog length</td><td class="value" data-label="Value">
			{{- if $admin }}
			<form class="inline-setting" data-api="/api/instance" data-method="PUT" data-next="/settings/instance?notice=Backlog+length+saved#runs">
				<input type="text" name="backlog_length" value="{{ if .Configured }}{{ .Configured }}{{ end }}" placeholder="{{ .Auto }}" inputmode="numeric" autocomplete="off" spellcheck="false" aria-label="Backlog length, empty for the automatic {{ .Auto }}">
				<button class="btn btn-sm" type="submit">Save</button>
			</form>
			{{- else }}<span class="mono">{{ if .Configured }}{{ .Configured }}{{ else }}<span class="faint">automatic</span>{{ end }}</span>{{ end }}
			<small class="faint">in effect: {{ .Length }}</small></td>
			<td class="var" data-label="Variable"><code>backlog_length</code></td><td class="desc" data-label="About"><span>How many runs may wait for a slot, up to 100000; empty is the automatic length. A start past it answers <code>429</code>.</span></td></tr>
		{{- end }}
		<tr><td class="key">Keep attachable</td><td class="value mono" data-label="Value">{{ with .Keep }}{{ . }}{{ else }}<span class="faint">default</span>{{ end }}</td><td class="var" data-label="Variable"><code>TASK_UI_KEEP</code></td><td class="desc" data-label="About"><span>How long a finished run stays in memory to attach to; its row and recording stay.</span></td></tr>
		<tr><td class="key">Stop grace</td><td class="value mono" data-label="Value">{{ with .Grace }}{{ . }}{{ else }}<span class="faint">default</span>{{ end }}</td><td class="var" data-label="Variable"><code>TASK_UI_GRACE</code></td><td class="desc" data-label="About"><span>How long a stopped run gets before the kill.</span></td></tr>
		<tr><td class="key">Artifact history</td><td class="value mono" data-label="Value">{{ if .ArtifactHistory }}{{ .ArtifactHistory }} versions{{ else }}every version{{ end }}</td><td class="var" data-label="Variable"><code>TASK_UI_ARTIFACT_HISTORY</code></td><td class="desc" data-label="About"><span>How many versions the artifact history keeps before it squashes to one.</span></td></tr>
		<tr><td class="key">Retention</td><td class="value" data-label="Value"><a href="/settings/retention">Data retention</a></td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>How long runs, recordings, prompt threads, workflow runs, artifact history and the secret read log stay.</span></td></tr>
	{{ template "instance-table-foot" }}
</section>

{{- if .ParentEnv }}
<section class="block instance-section" id="environment">
	<h3>Parent environment</h3>
	<p class="section-intro">How much of the server's own environment a run starts with. The folder's secrets and the integrations' variables reach a run under every choice; a script reads any other secret with <code>task-ui get-secret NAME</code>, as <a href="https://github.com/titpetric/task-ui/blob/main/docs/guide/settings.md#secrets-in-scripts">Secrets in scripts</a> shows.</p>
	{{- if $admin }}
	<form class="form parent-env" data-api="/api/instance" data-method="PUT" data-next="/settings/instance?notice=Parent+environment+saved#environment" data-parent-env>
		<div class="field">
			<span class="label">Forward <span class="faint">applies to the next run</span></span>
			<label class="check"><input type="radio" name="parent_env" value="none"{{ if eq .ParentEnv "none" }} checked{{ end }}> <strong>None</strong></label>
			<small class="faint">Secure: the base variables alone, less the sensitive ones.</small>
			<label class="check"><input type="radio" name="parent_env" value="auto"{{ if eq .ParentEnv "auto" }} checked{{ end }}> <strong>Auto</strong></label>
			<small class="faint">Filtered: every variable but the sensitive ones.</small>
			<label class="check"><input type="radio" name="parent_env" value="all"{{ if eq .ParentEnv "all" }} checked{{ end }}> <strong>All</strong></label>
			<small class="faint">Every variable, sensitive ones too.</small>
		</div>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
	<p class="faint">Base variables: {{ range $i, $n := .EnvBase }}{{ if $i }}, {{ end }}<code>{{ $n }}</code>{{ end }}. A name holding {{ range $i, $n := .EnvSensitive }}{{ if $i }}, {{ end }}<code>{{ $n }}</code>{{ end }} is sensitive.</p>
	{{- with .Env }}
	<div class="table-wrap">
	<table class="table env-table cards">
		<thead><tr><th>Variable</th><th>Forwarded</th><th>Flag</th></tr></thead>
		<tbody>
		{{- range . }}
		<tr data-env-none="{{ .None }}" data-env-auto="{{ .Auto }}"><td class="mono">{{ .Name }}</td>
			<td data-label="Forwarded" data-env-forwarded>{{ if .Forwarded $.Instance.ParentEnv }}<span class="badge ok">yes</span>{{ else }}<span class="badge">no</span>{{ end }}</td>
			<td data-label="Flag">{{ if .Sensitive }}<span class="badge bad" title="The name looks like a credential's">sensitive</span>{{ end }}</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- end }}
	{{- else }}
	<p>In effect: <strong>{{ .ParentEnv }}</strong>; an admin changes it here.</p>
	{{- end }}
</section>
{{- end }}

<section class="block instance-section" id="paths">
	<h3>Paths</h3>
	<p class="section-intro">Where task-ui keeps its data, the folders it runs and its keys.</p>
	{{ template "instance-table-head" }}
		<tr><td class="key">Workspace folder</td><td class="value mono" data-label="Value">{{ .Root }}</td><td class="var" data-label="Variable"><code>TASK_UI_WORKSPACE</code></td><td class="desc" data-label="About"><span>The folders with jobs to run, by default <code>workspaces</code> in the data directory.</span></td></tr>
		<tr><td class="key">Discovery</td><td class="value" data-label="Value">{{ if .Discover }}<span class="badge ok">on</span>{{ else }}<span class="badge">off</span>{{ end }}</td><td class="var" data-label="Variable"><code>TASK_UI_DISCOVER</code></td><td class="desc" data-label="About"><span>List every folder under the root with something to run.</span></td></tr>
		<tr><td class="key">Data</td><td class="value mono" data-label="Value">{{ .DataDir }}</td><td class="var" data-label="Variable"><code>TASK_UI_DATA</code></td><td class="desc" data-label="About"><span>The database, recordings and artifacts.</span></td></tr>
		<tr><td class="key">Data on the host</td><td class="value mono" data-label="Value">{{ with .HostData }}{{ . }}{{ else }}<span class="faint">the same</span>{{ end }}</td><td class="var" data-label="Variable"><code>TASK_UI_HOST_DATA</code></td><td class="desc" data-label="About"><span>The data directory as the host's docker daemon sees it; runs get it, and their own directory in <code>TASK_UI_HOST_PWD</code>.</span></td></tr>
		<tr><td class="key">Key folder</td><td class="value mono" data-label="Value">{{ .SSHDir }}</td><td class="var" data-label="Variable"><code>TASK_UI_SSH_DIR</code></td><td class="desc" data-label="About"><span>Every key of the instance, and the backups of rotated keys in <code>backup/</code>.</span></td></tr>
		<tr><td class="key">Compose project</td><td class="value mono" data-label="Value">{{ with .ComposeProject }}{{ . }}{{ else }}<span class="faint">none</span>{{ end }}</td><td class="var" data-label="Variable"><code>TASK_UI_COMPOSE_PROJECT</code></td><td class="desc" data-label="About"><span>The docker compose project task-ui runs in, whose up and down ask first.</span></td></tr>
	{{ template "instance-table-foot" }}
</section>
{{- end }}

{{- with .Instance }}{{ with .Storage }}
<section class="block instance-section" id="storage">
	<h3>Storage</h3>
	<p class="section-intro">What the instance keeps on disk: its database, the recordings and the artifacts of the runs, the tool history of each workspace's home and the instance's own copy, and the checkouts under each workspace's <code>src</code>. Measured now and then; an ellipsis stands for a measure still under way.</p>
	<div class="table-wrap">
	<table class="table storage-table cards">
		<colgroup><col class="c-item"><col class="c-where"><col class="c-size"><col class="c-files"></colgroup>
		<thead><tr><th>Item</th><th>Where</th><th class="num">Size</th><th class="num">Files</th></tr></thead>
		<tbody>
		{{- range .Rows }}
		<tr><td class="key">{{ .Name }}</td><td class="where mono" data-label="Where">{{ .Where }}</td><td class="num mono" data-label="Size">{{ if .Pending }}<span class="faint pending" aria-label="not measured yet"></span>{{ else }}{{ size .Size }}{{ end }}</td><td class="num mono" data-label="Files">{{ if .Pending }}<span class="faint pending" aria-label="not measured yet"></span>{{ else }}{{ .Files }}{{ end }}</td></tr>
		{{- end }}
		</tbody>
		<tfoot><tr><th>Total</th><td></td><td class="num mono" data-label="Size">{{ if .Total.Pending }}<span class="faint pending" aria-label="not measured yet"></span>{{ else }}{{ size .Total.Size }}{{ end }}</td><td class="num mono" data-label="Files">{{ if .Total.Pending }}<span class="faint pending" aria-label="not measured yet"></span>{{ else }}{{ .Total.Files }}{{ end }}</td></tr></tfoot>
	</table>
	</div>
</section>
{{- end }}{{ end }}

{{- if .Can "key/view" }}
{{ template "instance-keys" . }}
{{- end }}

{{- with .Instance }}
<section class="block instance-section" id="integrations">
	<h3>Integrations</h3>
	<p class="section-intro">The services task-ui reads from.</p>
	{{ template "instance-table-head" }}
		<tr><td class="key">Inventory</td><td class="value" data-label="Value">{{ if .Beszel }}<span class="badge ok">on</span>{{ else }}<span class="badge">off</span>{{ end }}</td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>Hosts from the Beszel <a href="/settings/integrations">integration</a>.</span></td></tr>
	{{ template "instance-table-foot" }}
</section>
{{- end }}
{{ end }}

{{/* The table every section of the Instance page holds, one width per
     column across the page: the setting, its value, the variable that
     sets it and what it does. */}}
{{ define "instance-table-head" }}<div class="table-wrap">
	<table class="table instance-table cards">
		<colgroup><col class="c-key"><col class="c-value"><col class="c-var"><col class="c-desc"></colgroup>
		<thead><tr><th>Setting</th><th>Value</th><th>Variable</th><th>Description</th></tr></thead>
		<tbody>{{ end }}
{{ define "instance-table-foot" }}</tbody>
	</table>
	</div>{{ end }}

{{/* The keys of the instance on its page: the encryption key and the
     key git pulls with, each with its actions behind a menu. */}}
{{ define "instance-keys" }}
{{- $git := "" }}{{ range .Keys }}{{ if .Git }}{{ $git = . }}{{ end }}{{ end }}
<section class="block instance-section" id="keys">
	<h3>Keys <small>in the key folder</small></h3>
	<p class="section-intro">The encryption key of the instance's secrets and the SSH key git pulls and clones with. <a href="/settings/keys">Every key</a></p>
	<div class="table-wrap fit">
	<table class="table instance-keys key-table cards">
		{{ template "keyhead" }}
		<tbody>
			{{- with .InstanceKey }}{{ template "keyrow" (list $ . "/settings/instance") }}
			{{- else }}
			<tr><td class="key">Encryption key</td><td colspan="4" data-label="Key"><span class="badge bad">missing</span> <span class="faint">The integrations' settings and the secrets need one.</span></td>
				<td class="actions">{{ if $.Can "key/edit" }}<form class="inline" data-api="/api/keys/encrypt/rotate" data-method="POST" data-next="/settings/instance#keys" data-confirm="Generate the instance's encryption key?" data-confirm-ok="Generate"><button class="btn btn-sm" type="submit">Generate</button></form> <a class="btn btn-sm" href="/settings/keys/upload">Upload</a>{{ end }}</td></tr>
			{{- end }}
			{{- with $git }}{{ template "keyrow" (list $ . "/settings/instance") }}
			{{- else }}
			<tr><td class="key">Git key</td><td colspan="4" data-label="Key"><span class="faint">none: git uses ssh's own</span></td>
				<td class="actions">{{ if $.Can "key/edit" }}<a class="btn btn-sm" href="/settings/keys">Choose</a>{{ end }}</td></tr>
			{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{ end }}



