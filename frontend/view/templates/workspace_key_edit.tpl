{{ define "content" }}
{{- $ws := .Workspace }}
{{- with $k := .Key }}
<div class="page-head">
	<div>
		<h2>Edit {{ $k.Name }}</h2>
		<p>{{ if $k.Comment }}{{ $k.Comment }}, {{ end }}<span class="dim">{{ $k.Kind }}</span> {{ $k.Fingerprint }}, an SSH key of the workspace {{ $ws.Folder }}. Choose whose runs in the workspace get it. With nothing checked, every run does; check roles or users to limit it to their runs.</p>
	</div>
</div>

{{ template "wskeylinks" (list $ $k (print "/settings/workspaces/" $ws.Folder) false) }}
{{- end }}
{{ end }}
