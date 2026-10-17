{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Secrets</h2>
		<p>Variables added to the environment of every run. A secret for all folders reaches every job; a secret for one folder reaches that folder's jobs and wins over the shared one. Before a run, they are written into the env files the pipeline reads.</p>
	</div>
	<div class="actions">
		<a class="btn btn-primary" href="/settings/secrets/new">Add secret</a>
	</div>
</div>

<div class="table-wrap">
<table class="table">
	<thead><tr><th>Scope</th><th>Name</th><th>Value</th><th>Updated</th><th>Actions</th></tr></thead>
	<tbody>
	{{- range .Secrets }}
		<tr>
			<td class="nowrap">{{ if eq .Scope "all" }}<span class="badge signal">all folders</span>{{ else }}<a href="/f/{{ .Scope }}"><svg class="ic"><use href="#i-folder"/></svg> {{ .Scope }}</a>{{ end }}</td>
			<td class="key">{{ .Name }}</td>
			<td class="faint">{{ mask .Content }}</td>
			<td class="nowrap" title="{{ stamp .UpdatedAt }}">{{ ago .UpdatedAt }}</td>
			<td class="actions">
				<a class="btn-icon" href="/settings/secrets/edit?scope={{ .Scope }}&name={{ .Name }}" title="Replace" aria-label="Replace {{ .Name }}"><svg class="ic"><use href="#i-edit"/></svg></a>
				<button type="button" class="btn-icon danger" data-delete="/api/secrets/{{ .ID }}" data-confirm="Delete {{ .Name }} from {{ if eq .Scope "all" }}all folders{{ else }}{{ .Scope }}{{ end }}? Runs stop receiving it." data-next="reload" title="Delete" aria-label="Delete {{ .Name }}"><svg class="ic"><use href="#i-trash"/></svg></button>
			</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="5">No secrets yet. Add one for all folders, or for a single folder.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
{{ end }}
