{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Instance</h2>
		<p>What the server runs with. The environment sets these; change one there and restart task-ui.</p>
	</div>
</div>

{{- with .Instance }}
<div class="table-wrap">
<table class="table settings instance">
	<thead><tr><th>Setting</th><th>Value</th><th>Variable</th><th>Description</th></tr></thead>
	<tbody>
		<tr><td class="key nowrap">Base link</td><td class="mono">{{ $.BaseURL }}</td><td class="nowrap" data-label="Variable"><code>TASK_UI_URL</code></td><td class="desc dim">{{ if .URL }}The link pages and MCP clients use.{{ else }}Empty: task-ui reads the link from each request.{{ end }}</td></tr>
		<tr><td class="key nowrap">Version</td><td class="mono">{{ $.Version }}</td><td class="nowrap" data-label="Variable"><span class="faint">none</span></td><td class="desc dim"></td></tr>
		<tr><td class="key nowrap">Listen</td><td class="mono">{{ .Listen }}</td><td class="nowrap" data-label="Variable"><code>TASK_UI_LISTEN</code></td><td class="desc dim">The address the server listens on.</td></tr>
		<tr><td class="key nowrap">Root</td><td class="mono">{{ .Root }}</td><td class="nowrap" data-label="Variable"><code>TASK_UI_ROOT</code></td><td class="desc dim">The folders with jobs to run.</td></tr>
		<tr><td class="key nowrap">Discovery</td><td>{{ if .Discover }}<span class="badge ok">on</span>{{ else }}<span class="badge">off</span>{{ end }}</td><td class="nowrap" data-label="Variable"><code>TASK_UI_DISCOVER</code></td><td class="desc dim">List every folder under the root with something to run.</td></tr>
		<tr><td class="key nowrap">Data</td><td class="mono">{{ .DataDir }}</td><td class="nowrap" data-label="Variable"><code>TASK_UI_DATA</code></td><td class="desc dim">The database, recordings and artifacts.</td></tr>
		<tr><td class="key nowrap">SSH folder</td><td class="mono">{{ .SSHDir }}</td><td class="nowrap" data-label="Variable"><code>TASK_UI_SSH_DIR</code></td><td class="desc dim">The instance's SSH keys.</td></tr>
		<tr><td class="key nowrap">Password file</td><td class="mono">{{ with .Htpasswd }}{{ . }}{{ else }}<span class="faint">none</span>{{ end }}</td><td class="nowrap" data-label="Variable"><code>TASK_UI_HTPASSWD</code></td><td class="desc dim">Who can sign in.</td></tr>
		<tr><td class="key nowrap">Sign in</td><td>{{ if .AutoLogin }}<span class="badge bad">auto login</span> {{ template "user" (list $ .AutoLogin) }}{{ else if $.LoginRequired }}<span class="badge ok">required</span>{{ else }}<span class="badge">open</span>{{ end }}</td><td class="nowrap" data-label="Variable"><code>TASK_UI_AUTOLOGIN</code></td><td class="desc dim">Signs everyone in as one user; for demos only.</td></tr>
		<tr><td class="key nowrap">Admins</td><td>{{ range .Admins }}{{ template "user" (list $ .) }} {{ else }}<span class="faint">none</span>{{ end }}</td><td class="nowrap" data-label="Variable"><code>TASK_UI_ADMINS</code></td><td class="desc dim">Users with every permission.</td></tr>
		<tr><td class="key nowrap">Viewers</td><td>{{ range .Viewers }}{{ template "user" (list $ .) }} {{ else }}<span class="faint">none</span>{{ end }}</td><td class="nowrap" data-label="Variable"><code>TASK_UI_VIEWERS</code></td><td class="desc dim">Read only. Everyone else is a runner.</td></tr>
		<tr><td class="key nowrap">Agent</td><td>{{ with .Agent }}{{ template "user" (list $ .) }}{{ else }}<span class="faint">none</span>{{ end }}</td><td class="nowrap" data-label="Variable"><code>TASK_UI_AGENT</code></td><td class="desc dim">The user a note names to start the agent job.</td></tr>
		<tr><td class="key nowrap">Agent job</td><td class="mono">{{ with .AgentJob }}{{ . }}{{ else }}<span class="faint">none</span>{{ end }}</td><td class="nowrap" data-label="Variable"><code>TASK_UI_AGENT_JOB</code></td><td class="desc dim">The job that answers those notes.</td></tr>
		<tr><td class="key nowrap">Inventory</td><td>{{ if .Beszel }}<span class="badge ok">on</span>{{ else }}<span class="badge">off</span>{{ end }}</td><td class="nowrap" data-label="Variable"><span class="faint">none</span></td><td class="desc dim">Hosts from the Beszel <a href="/settings/integrations">integration</a>.</td></tr>
		<tr><td class="key nowrap">MCP</td><td class="mono">{{ $.BaseURL }}/api/mcp</td><td class="nowrap" data-label="Variable"><span class="faint">none</span></td><td class="desc dim">Where MCP clients connect, with a user's token.</td></tr>
	</tbody>
</table>
</div>
{{- end }}
{{ end }}
