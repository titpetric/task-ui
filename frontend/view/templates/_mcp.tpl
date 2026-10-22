{{/* copy-line is a value in a code field with a copy button: (list value label). */}}
{{ define "copy-line" }}
{{- $value := index . 0 }}{{ $label := index . 1 -}}
<div class="copy-line"><code>{{ $value }}</code><button type="button" class="btn-icon" data-copy="{{ $value }}" title="Copy {{ $label }}" aria-label="Copy {{ $label }}"><svg class="ic"><use href="#i-copy"/></svg></button></div>
{{- end }}

{{/* mcp-connect tells how to connect an MCP client to this instance, with the base link to copy. */}}
{{ define "mcp-connect" }}
{{- $base := .BaseURL }}
<section class="block mcp-connect" id="mcp">
	<h3>Connect an MCP client</h3>
	<p>task-ui is an MCP server: an agent such as Claude Code sees the jobs and workflows, runs them and reads how they went, as the user of its token.</p>
	<dl class="copy-list">
		<dt>Base link</dt>
		<dd>{{ template "copy-line" (list $base "the base link") }}</dd>
		<dt>MCP endpoint</dt>
		<dd>{{ template "copy-line" (list (printf "%s/api/mcp" $base) "the MCP endpoint") }}</dd>
	</dl>
	<ol class="mcp-steps">
		<li>{{ if .User }}Enable MCP on <a href="/settings/users/{{ .User.ID }}#mcp">your user settings</a>{{ else }}Enable MCP on your user settings{{ end }} and copy the token: it shows once and lasts 30 days, until you rotate it there. Enabling takes Allow Edits on Users, so an admin may do it for you.</li>
		<li>Add the server to Claude Code:
			{{ template "copy-line" (list (printf "claude mcp add --transport http --scope user task %s/api/mcp --header \"Authorization: Bearer <token>\"" $base) "the claude mcp add command") }}</li>
		<li>Any other client takes the endpoint over HTTP and the header <code>Authorization: Bearer &lt;token&gt;</code>. <code>claude mcp list</code> shows <code>task</code> as connected.</li>
	</ol>
	<p class="faint">The tools and the rest: <a href="https://github.com/titpetric/task-ui/blob/main/docs/guide/mcp.md">the MCP chapter of the guide</a>.</p>
</section>
{{- end }}


{{/* mcp-access is a user's MCP token on their settings page: (list page user). The token shows once, in #mcp-token, from the answer of Enable or Rotate. */}}
{{ define "mcp-access" }}
{{- $p := index . 0 }}{{ $u := index . 1 }}{{ $m := $p.AccountMCP }}
{{- $api := printf "/api/users/%s/mcp" $u.ID }}
<section class="block mcp-access" id="mcp">
	<h3>MCP access</h3>
	{{- if $m.Enabled }}
	<dl class="copy-list" data-before-reveal>
		<dt>Created</dt>
		<dd title="{{ stamp $m.CreatedAt }}">{{ ago $m.CreatedAt }}</dd>
		<dt>Expires</dt>
		<dd>{{ if $m.ExpiresAt }}<span class="badge{{ with $m.Tone $p.Now }} {{ . }}{{ end }}" title="{{ stamp $m.ExpiresAt }}">{{ if $m.Expired $p.Now }}expired {{ end }}{{ stamp $m.ExpiresAt }}</span>{{ else }}never, a token of <code>task-ui token</code>{{ end }}</dd>
		<dt>Last used</dt>
		<dd{{ with $m.UsedAt }} title="{{ stamp . }}"{{ end }}>{{ if $m.UsedAt }}{{ ago $m.UsedAt }}{{ else }}<span class="faint">never</span>{{ end }}</dd>
	</dl>
	{{- if $m.CanManage }}
	<div class="actions" data-before-reveal>
		<form class="inline" data-api="{{ $api }}" data-method="POST" data-reveal="#mcp-token" data-confirm="Rotate the MCP token of {{ $u.Name }}? The current one stops working at once; the new one lasts 30 days." data-confirm-ok="Rotate"><button class="btn" type="submit">Rotate</button></form>
		<button type="button" class="btn btn-danger" data-delete="{{ $api }}" data-confirm="Disable MCP for {{ $u.Name }}? The token stops working at once." data-confirm-ok="Disable" data-next="reload">Disable</button>
	</div>
	{{- end }}
	{{- else }}
	<p data-before-reveal>Off: no MCP client signs in as {{ $u.Name }} with a token from here. Enable makes a token that lasts 30 days and shows it once.</p>
	{{- if $m.CanEnable }}
	<form class="inline" data-api="{{ $api }}" data-method="POST" data-reveal="#mcp-token" data-before-reveal><button class="btn btn-primary" type="submit">Enable MCP</button></form>
	{{- else }}
	<p class="faint">Enabling takes Allow Edits on Users: ask an admin.</p>
	{{- end }}
	{{- end }}
	<div id="mcp-token" class="mcp-token" hidden>
		<div class="alert">Copy the token now: it is not shown again. It lasts until <b data-fill="{expires_at}"></b>; rotate it here before then.</div>
		<dl class="copy-list">
			<dt>Token</dt>
			<dd><div class="copy-line"><code data-fill="{token}"></code><button type="button" class="btn-icon" data-copy="" data-fill="{token}" title="Copy the token" aria-label="Copy the token"><svg class="ic"><use href="#i-copy"/></svg></button></div></dd>
			<dt>Claude Code</dt>
			{{- $add := printf "claude mcp add --transport http --scope user task %s/api/mcp --header \"Authorization: Bearer {token}\"" $p.BaseURL }}
			<dd><div class="copy-line"><code data-fill="{{ $add }}"></code><button type="button" class="btn-icon" data-copy="" data-fill="{{ $add }}" title="Copy the claude mcp add command" aria-label="Copy the claude mcp add command"><svg class="ic"><use href="#i-copy"/></svg></button></div></dd>
		</dl>
		<a class="btn" href="/settings/users/{{ $u.ID }}">Done</a>
	</div>
</section>
{{- end }}
