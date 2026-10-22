{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Instance</h2>
		<p>The settings the server runs with. The environment and the command line set these values; to change one, change it there and restart task-ui. This page never shows a password, token or key.</p>
	</div>
</div>

{{- with .Instance }}
<div class="table-wrap">
<table class="table settings instance">
	<tbody>
		<tr><td class="key nowrap">Base link</td><td class="mono">{{ $.BaseURL }}</td><td class="desc dim">{{ if .URL }}TASK_UI_URL{{ else }}TASK_UI_URL is empty: task-ui reads the link from each request.{{ end }}</td></tr>
		<tr><td class="key nowrap">Version</td><td class="mono">{{ $.Version }}</td><td class="desc dim"></td></tr>
		<tr><td class="key nowrap">Listen</td><td class="mono">{{ .Listen }}</td><td class="desc dim">TASK_UI_LISTEN</td></tr>
		<tr><td class="key nowrap">Root</td><td class="mono">{{ .Root }}</td><td class="desc dim">TASK_UI_ROOT: the folders with jobs to run.</td></tr>
		<tr><td class="key nowrap">Discovery</td><td>{{ if .Discover }}<span class="badge ok">on</span>{{ else }}<span class="badge">off</span>{{ end }}</td><td class="desc dim">TASK_UI_DISCOVER: list every folder under the root with something to run.</td></tr>
		<tr><td class="key nowrap">Data</td><td class="mono">{{ .DataDir }}</td><td class="desc dim">TASK_UI_DATA: the database, the recordings and the artifacts.</td></tr>
		<tr><td class="key nowrap">SSH folder</td><td class="mono">{{ .SSHDir }}</td><td class="desc dim">TASK_UI_SSH_DIR: the instance's SSH keys.</td></tr>
		<tr><td class="key nowrap">Password file</td><td class="mono">{{ with .Htpasswd }}{{ . }}{{ else }}<span class="faint">none</span>{{ end }}</td><td class="desc dim">TASK_UI_HTPASSWD</td></tr>
		<tr><td class="key nowrap">Sign in</td><td>{{ if .AutoLogin }}<span class="badge bad">auto login</span> {{ template "user" (list $ .AutoLogin) }}{{ else if $.LoginRequired }}<span class="badge ok">required</span>{{ else }}<span class="badge">open</span>{{ end }}</td><td class="desc dim">TASK_UI_AUTOLOGIN</td></tr>
		<tr><td class="key nowrap">Admins</td><td>{{ range .Admins }}{{ template "user" (list $ .) }} {{ else }}<span class="faint">none</span>{{ end }}</td><td class="desc dim">TASK_UI_ADMINS</td></tr>
		<tr><td class="key nowrap">Viewers</td><td>{{ range .Viewers }}{{ template "user" (list $ .) }} {{ else }}<span class="faint">none</span>{{ end }}</td><td class="desc dim">TASK_UI_VIEWERS: everyone else in the password file is a runner.</td></tr>
		<tr><td class="key nowrap">Agent</td><td>{{ with .Agent }}{{ template "user" (list $ .) }}{{ else }}<span class="faint">none</span>{{ end }}</td><td class="desc dim">TASK_UI_AGENT: the user a note names to start the agent job.</td></tr>
		<tr><td class="key nowrap">Agent job</td><td class="mono">{{ with .AgentJob }}{{ . }}{{ else }}<span class="faint">none</span>{{ end }}</td><td class="desc dim">TASK_UI_AGENT_JOB</td></tr>
		<tr><td class="key nowrap">Inventory</td><td>{{ if .Beszel }}<span class="badge ok">on</span>{{ else }}<span class="badge">off</span>{{ end }}</td><td class="desc dim">The Beszel integration lists the hosts of the Inventory. Set it up under Integrations.</td></tr>
		<tr><td class="key nowrap">MCP</td><td class="mono">{{ $.BaseURL }}/api/mcp</td><td class="desc dim">MCP clients connect here with a user's token.</td></tr>
	</tbody>
</table>
</div>
{{- end }}
{{ end }}
