{{ define "content" }}
{{- $admin := and .User .User.CanAdmin }}
<div class="page-head">
	<div>
		<h2>Instance</h2>
		<p>What the server runs with. The environment sets most of these: change one there and restart task-ui. The backlog length is a setting of its own, saved here{{ if not $admin }} by an admin{{ end }}.</p>
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
		<tr><td class="key">Retention</td><td class="value" data-label="Value"><a href="/settings/retention">Data retention</a></td><td class="var" data-label="Variable"><span class="faint">none</span></td><td class="desc" data-label="About"><span>How long recordings, artifact history and the read logs stay.</span></td></tr>
	{{ template "instance-table-foot" }}
</section>

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
	<div class="table-wrap">
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



