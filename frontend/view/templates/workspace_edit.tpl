{{ define "content" }}
{{- $ws := .Workspace }}
{{- $nested := index .Nested $ws.Folder }}{{ $default := eq $ws.Folder "_" }}
<div class="page-head">
	<div>
		<h2>Configure {{ $ws.Folder }}</h2>
		<p>A workspace's title and description. Its folder, repository and branch stay: another checkout is another workspace.{{ if $default }} The default workspace is the root, and stays.{{ end }}</p>
	</div>
	{{- if not $default }}
	<div class="actions">
		{{- template "workspacedelete" (list $ws $nested "/settings/workspaces?notice=Workspace+removed") }}
	</div>
	{{- end }}
</div>

<form class="form" data-api="/api/workspaces/{{ $ws.Folder }}" data-method="PUT" data-next="/settings/workspaces?notice=Saved">
	<div class="field">
		<span class="label">Repository</span>
		<span class="mono-break">{{ if $default }}<span class="faint">the root, {{ .Root }}</span>{{ else }}{{ $ws.SourceURL }}{{ with $ws.Branch }} <span class="badge">{{ . }}</span>{{ end }}{{ end }}</span>
	</div>
	<div class="field">
		<label for="we-title">Title</label>
		<input type="text" id="we-title" name="title" value="{{ index .Form "title" }}" autocomplete="off">
	</div>
	<div class="field">
		<label for="we-description">Description</label>
		<input type="text" id="we-description" name="description" value="{{ index .Form "description" }}" autocomplete="off">
	</div>
	<div class="field">
		<span class="label">Projects <span class="faint">{{ if $default }}at the top of the root{{ else }}cloned inside it; removing the workspace removes them{{ end }}</span></span>
		<span>{{ range $i, $id := $nested }}{{ if $i }}, {{ end }}<a href="/settings/projects/{{ $id }}">{{ $id }}</a>{{ else }}<span class="faint">none yet</span>{{ end }}</span>
	</div>
	<div class="form-foot">
		<a class="btn" href="/settings/workspaces">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{ end }}
