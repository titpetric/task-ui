{{ define "mcp-connect" }}
{{- $base := .BaseURL }}
<section class="block mcp-connect" id="mcp">
	<h3>Connect an MCP client</h3>
	<p>task-ui is an MCP server. An agent such as Claude Code can list jobs and workflows, run them and read the results, with the permissions of the user whose token it uses.</p>
	<dl class="copy-list">
		<dt>Base link</dt>
		<dd>{{ template "copyable" (list $base "the base link") }}</dd>
		<dt>MCP endpoint</dt>
		<dd>{{ template "copyable" (list (printf "%s/api/mcp" $base) "the MCP endpoint") }}</dd>
	</dl>
	<ol class="mcp-steps">
		<li>{{ if .User }}Make a token under <a href="/users/{{ .User.Name }}/edit#mcp">your MCP access</a>{{ else }}Make a token under your MCP access{{ end }} and copy it. It shows once and lasts 30 days, or until you delete it; make one per client, so retiring a client takes nothing else with it. A token needs MCP agent user on, which an admin sets, and making one needs Allow Edits on Users; if you don't have them, ask someone who does.</li>
		<li>Add the server to Claude Code:
			{{ template "copyable" (list (printf "claude mcp add --transport http --scope user task %s/api/mcp --header \"Authorization: Bearer <token>\"" $base) "the claude mcp add command") }}</li>
		<li>Other clients connect to the endpoint over HTTP with the header <code>Authorization: Bearer &lt;token&gt;</code>. Check with <code>claude mcp list</code>: it shows <code>task</code> as connected.</li>
	</ol>
	<p class="faint"><a href="https://github.com/titpetric/task-ui/blob/main/docs/guide/mcp.md">The MCP chapter of the guide</a></p>
</section>
{{- end }}


{{ define "mcp-access" }}
{{- $p := index . 0 }}{{ $u := index . 1 }}{{ $m := $p.AccountMCP }}
{{- $api := printf "/api/users/%s/mcp" $u.ID }}
<section class="block mcp-access" id="mcp">
	<h3>MCP access</h3>
	{{- if $m.Enabled }}
	<div class="table-wrap" data-before-reveal>
	<table class="table cards">
		<thead><tr><th>Token</th><th>Created</th><th>Expires</th><th>Last used</th><th></th></tr></thead>
		<tbody>
		{{- range $m.Tokens }}
			<tr>
				<td class="key nowrap" data-label="Token">{{ .Name }}</td>
				<td class="nowrap" data-label="Created" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</td>
				<td class="nowrap" data-label="Expires">{{ if .ExpiresAt }}<span class="badge{{ with .Tone $p.Now }} {{ . }}{{ end }}" title="{{ stamp .ExpiresAt }}">{{ if .Expired $p.Now }}expired {{ end }}{{ stamp .ExpiresAt }}</span>{{ else }}<span class="faint">never: a token from <code>task-ui token</code></span>{{ end }}</td>
				<td class="nowrap" data-label="Last used"{{ with .UsedAt }} title="{{ stamp . }}"{{ end }}>{{ if .UsedAt }}{{ ago .UsedAt }}{{ else }}<span class="faint">never</span>{{ end }}</td>
				<td class="actions">{{- if $m.CanManage }}<button type="button" class="btn-icon danger" data-delete="{{ $api }}/{{ .ID }}" data-confirm="Delete the MCP token {{ .Name }} of {{ $u.Name }}? Whatever client holds it stops working right away. The other tokens keep working." data-confirm-ok="Delete" data-next="reload" title="Delete the token {{ .Name }}" aria-label="Delete the token {{ .Name }}"><svg class="ic"><use href="#i-trash"/></svg></button>{{- end }}</td>
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- else }}
	<p data-before-reveal>{{ $u.Name }} holds no MCP token. New token makes one that lasts 30 days and shows it once.</p>
	{{- end }}
	{{- if not $m.Agent }}
	<p class="faint">A token needs MCP agent user, under <a href="#roles">Roles</a>.</p>
	{{- else if $m.CanManage }}
	<div class="actions" data-before-reveal>
		<form class="inline" data-api="{{ $api }}" data-method="POST" data-reveal="#mcp-token"><button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-plus"/></svg>New token</button></form>
	</div>
	<p class="faint" data-before-reveal>One token per client, so retiring a client is one Delete and the others keep signing in.</p>
	{{- else }}
	<p class="faint">Making a token needs Allow Edits on Users.</p>
	{{- end }}
	<div id="mcp-token" class="mcp-token" hidden>
		<div class="alert">Copy the token now; you won't see it again. <b data-fill="{name}"></b> works until <b data-fill="{expires_at}"></b>. Make another here, or delete this one, whenever you like.</div>
		<dl class="copy-list">
			<dt>Token</dt>
			<dd><span class="copyable"><code data-fill="{token}"></code><button type="button" class="btn-icon" data-copy="" data-fill="{token}" title="Copy the token" aria-label="Copy the token"><svg class="ic"><use href="#i-copy"/></svg></button></span></dd>
			<dt>Claude Code</dt>
			{{- $add := printf "claude mcp add --transport http --scope user task %s/api/mcp --header \"Authorization: Bearer {token}\"" $p.BaseURL }}
			<dd><span class="copyable"><code data-fill="{{ $add }}"></code><button type="button" class="btn-icon" data-copy="" data-fill="{{ $add }}" title="Copy the claude mcp add command" aria-label="Copy the claude mcp add command"><svg class="ic"><use href="#i-copy"/></svg></button></span></dd>
			<dt>MCP settings</dt>
			{{- $settings := printf "{\"mcpServers\":{\"task\":{\"type\":\"http\",\"url\":\"%s/api/mcp\",\"headers\":{\"Authorization\":\"Bearer {token}\"}}}}" $p.BaseURL }}
			<dd><span class="copyable"><code data-fill="{{ $settings }}"></code><button type="button" class="btn-icon" data-copy="" data-fill="{{ $settings }}" title="Copy the MCP settings" aria-label="Copy the MCP settings"><svg class="ic"><use href="#i-copy"/></svg></button></span></dd>
		</dl>
		<a class="btn" href="/users/{{ $u.Name }}/edit?notice=MCP+token+made#mcp">Done</a>
	</div>
</section>
{{- end }}
