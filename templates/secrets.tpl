{{ define "content" }}
{{ template "crumbs" (list (crumb "Settings" "/settings") (crumb "Secrets" "")) }}
<div class="page-head">
	<div>
		<h2>Secrets</h2>
		<p>Variables added to the environment of every run. A secret for all folders reaches every job; a secret for one folder reaches that folder's jobs and wins over the shared one. Before a run, they are written into the env files the pipeline reads.</p>
	</div>
	<div class="actions">
		<a class="btn btn-primary" href="/settings/secrets/new">Add secret</a>
	</div>
</div>

{{- if .Notice }}<div class="alert ok" role="status">{{ .Notice }}</div>{{ end }}
{{- if .Error }}<div class="alert bad" role="alert">{{ .Error }}</div>{{ end }}

<div class="table-wrap">
<table class="table">
	<thead><tr><th>Scope</th><th>Key</th><th>Value</th><th>Actions</th></tr></thead>
	<tbody>
	{{- range .Secrets }}
		<tr>
			<td class="nowrap">{{ if eq .Scope "all" }}<span class="badge signal">all folders</span>{{ else }}<a href="/f/{{ .Scope }}"><svg class="ic"><use href="#i-folder"/></svg> {{ .Scope }}</a>{{ end }}</td>
			<td class="key">{{ .Key }}</td>
			<td class="faint">{{ mask .Value }}</td>
			<td class="actions">
				<a class="btn-icon" href="/settings/secrets/edit?scope={{ .Scope }}&key={{ .Key }}" title="Edit" aria-label="Edit {{ .Key }}"><svg class="ic"><use href="#i-edit"/></svg></a>
				<form method="post" action="/settings/secrets/delete" data-confirm="Delete {{ .Key }} from {{ if eq .Scope "all" }}all folders{{ else }}{{ .Scope }}{{ end }}? Runs stop receiving it." data-confirm-ok="Delete">
					<input type="hidden" name="scope" value="{{ .Scope }}">
					<input type="hidden" name="key" value="{{ .Key }}">
					<button class="btn-icon danger" type="submit" title="Delete" aria-label="Delete {{ .Key }}"><svg class="ic"><use href="#i-trash"/></svg></button>
				</form>
			</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="4">No secrets yet. Add one for all folders, or for a single folder.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
{{ end }}
