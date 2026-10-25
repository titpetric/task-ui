{{ define "content" }}
{{- $ws := .Workspace }}
<div class="page-head">
	<div>
		<h2>Generate key</h2>
		<p>Generates an SSH key of the workspace {{ $ws.Folder }} in {{ .KeysDir }}, named <span class="mono">workspace_{{ $ws.Folder }}_&lt;name&gt;_&lt;algorithm&gt;</span>. Its runs get it in their ssh agent. A name is unique across the folder.</p>
	</div>
</div>

{{ template "wskeygenerate" (list . (print "/settings/workspaces/" $ws.Folder) false) }}
{{ end }}
