{{/* mcp-connect tells how to connect an MCP client to this instance, with the base link to copy. */}}
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
		<li>{{ if .User }}Enable MCP on <a href="/users/{{ .User.Name }}#mcp">your profile</a>{{ else }}Enable MCP on your profile{{ end }} and copy the token. It shows once and lasts 30 days, or until you rotate it. A token needs MCP agent user on, which an admin sets, and enabling needs Allow Edits on Users; if you don't have them, ask someone who does.</li>
		<li>Add the server to Claude Code:
			{{ template "copyable" (list (printf "claude mcp add --transport http --scope user task %s/api/mcp --header \"Authorization: Bearer <token>\"" $base) "the claude mcp add command") }}</li>
		<li>Other clients connect to the endpoint over HTTP with the header <code>Authorization: Bearer &lt;token&gt;</code>. Check with <code>claude mcp list</code>: it shows <code>task</code> as connected.</li>
	</ol>
	<p class="faint">For the list of tools, read <a href="https://github.com/titpetric/task-ui/blob/main/docs/guide/mcp.md">the MCP chapter of the guide</a>.</p>
</section>
{{- end }}


{{/* mcp-access is a user's MCP token on their profile: (list page user). The token shows once, in #mcp-token, from the response to Enable or Rotate. */}}
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
		<dd>{{ if $m.ExpiresAt }}<span class="badge{{ with $m.Tone $p.Now }} {{ . }}{{ end }}" title="{{ stamp $m.ExpiresAt }}">{{ if $m.Expired $p.Now }}expired {{ end }}{{ stamp $m.ExpiresAt }}</span>{{ else }}never: a token from <code>task-ui token</code>{{ end }}</dd>
		<dt>Last used</dt>
		<dd{{ with $m.UsedAt }} title="{{ stamp . }}"{{ end }}>{{ if $m.UsedAt }}{{ ago $m.UsedAt }}{{ else }}<span class="faint">never</span>{{ end }}</dd>
	</dl>
	{{- if $m.CanManage }}
	<div class="actions" data-before-reveal>
		<form class="inline" data-api="{{ $api }}" data-method="POST" data-reveal="#mcp-token" data-confirm="Rotate the MCP token of {{ $u.Name }}? The current token stops working right away. The new one lasts 30 days." data-confirm-ok="Rotate"><button class="btn" type="submit">Rotate</button></form>
		<button type="button" class="btn btn-danger" data-delete="{{ $api }}" data-confirm="Disable MCP for {{ $u.Name }}? The token stops working right away." data-confirm-ok="Disable" data-next="reload">Disable</button>
	</div>
	{{- end }}
	{{- else }}
	<p data-before-reveal>MCP is off for {{ $u.Name }}. Enable creates a token that lasts 30 days and shows it once.</p>
	{{- if not $m.Agent }}
	<p class="faint">A token needs MCP agent user, under <a href="#roles">Roles</a>, which an admin turns on.</p>
	{{- else if $m.CanEnable }}
	<form class="inline" data-api="{{ $api }}" data-method="POST" data-reveal="#mcp-token" data-before-reveal><button class="btn btn-primary" type="submit">Enable MCP</button></form>
	{{- else }}
	<p class="faint">Enabling needs Allow Edits on Users. Ask someone who has it.</p>
	{{- end }}
	{{- end }}
	<div id="mcp-token" class="mcp-token" hidden>
		<div class="alert">Copy the token now; you won't see it again. It works until <b data-fill="{expires_at}"></b>. Rotate it here before then.</div>
		<dl class="copy-list">
			<dt>Token</dt>
			<dd><span class="copyable"><code data-fill="{token}"></code><button type="button" class="btn-icon" data-copy="" data-fill="{token}" title="Copy the token" aria-label="Copy the token"><svg class="ic"><use href="#i-copy"/></svg></button></span></dd>
			<dt>Claude Code</dt>
			{{- $add := printf "claude mcp add --transport http --scope user task %s/api/mcp --header \"Authorization: Bearer {token}\"" $p.BaseURL }}
			<dd><span class="copyable"><code data-fill="{{ $add }}"></code><button type="button" class="btn-icon" data-copy="" data-fill="{{ $add }}" title="Copy the claude mcp add command" aria-label="Copy the claude mcp add command"><svg class="ic"><use href="#i-copy"/></svg></button></span></dd>
		</dl>
		<a class="btn" href="/users/{{ $u.Name }}#mcp">Done</a>
	</div>
</section>
{{- end }}
