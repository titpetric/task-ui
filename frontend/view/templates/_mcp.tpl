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
		<li>Get a token for your user. An admin runs this on the server; task-ui prints the token once and keeps only its hash:
			{{ template "copy-line" (list (printf "docker exec task task-ui token %s mcp" (or (and .User .User.Name) "<you>")) "the token command") }}</li>
		<li>Add the server to Claude Code:
			{{ template "copy-line" (list (printf "claude mcp add --transport http --scope user task %s/api/mcp --header \"Authorization: Bearer <token>\"" $base) "the claude mcp add command") }}</li>
		<li>Any other client takes the endpoint over HTTP and the header <code>Authorization: Bearer &lt;token&gt;</code>. <code>claude mcp list</code> shows <code>task</code> as connected.</li>
	</ol>
	<p class="faint">The tools and the rest: <a href="https://github.com/titpetric/task-ui/blob/main/docs/guide/mcp.md">the MCP chapter of the guide</a>.</p>
</section>
{{- end }}
