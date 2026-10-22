{{ define "content" }}
{{- $ws := .Workspace }}
{{- with $k := .Key }}
<div class="page-head">
	<div>
		<h2>{{ $k.Name }}</h2>
		<p>{{ if $k.Comment }}{{ $k.Comment }}, {{ end }}<span class="dim">{{ $k.Type }}</span> {{ $k.Fingerprint }}, an SSH key of the workspace {{ $ws.Folder }}. Choose whose runs in the workspace get it. With nothing checked, every run does; check roles or users to limit it to their runs.</p>
	</div>
</div>

<form class="form" data-api="/api/workspaces/{{ $ws.Folder }}/keys/{{ $k.Name }}" data-method="PUT" data-next="/settings/workspaces/{{ $ws.Folder }}?notice=Grants+saved">
	{{- template "keygrants" (list (index $.KeyPolicies $k.Name) (index $.KeyUsers $k.Name) $.Users) }}
	<div class="form-foot">
		<a class="btn" href="/settings/workspaces/{{ $ws.Folder }}">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{- end }}
{{ end }}
