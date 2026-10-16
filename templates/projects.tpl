{{ define "content" }}
{{ template "crumbs" (list (crumb "Settings" "/settings") (crumb "Projects" "")) }}
<div class="page-head">
	<div>
		<h2>Projects</h2>
		<p>The folders under the root. A folder with something to run shows up in the sidebar; one cloned from a git repository is pulled before every run.</p>
	</div>
	<div class="actions">
		<a class="btn btn-primary" href="/settings/projects/new">Add project</a>
	</div>
</div>

{{- if .Notice }}<div class="alert ok" role="status">{{ .Notice }}</div>{{ end }}
{{- if .Error }}<div class="alert bad" role="alert">{{ .Error }}</div>{{ end }}

<div class="table-wrap">
<table class="table">
	<thead><tr><th>Folder</th><th>Runs</th><th>Source</th><th class="hide-sm">Last commit</th><th>Actions</th></tr></thead>
	<tbody>
	{{- range .Entries }}
		<tr>
			<td class="key nowrap">{{ if .Sources }}<a href="/f/{{ .ID }}"><svg class="ic"><use href="#i-folder"/></svg> {{ .ID }}</a>{{ else }}<svg class="ic"><use href="#i-folder"/></svg> {{ .ID }}{{ end }}</td>
			<td class="nowrap">{{ if .Error }}<span class="badge bad">error</span> {{ end }}{{ if .Sources }}{{ join .Sources ", " }} <span class="faint">{{ .Jobs }} jobs</span>{{ else }}<span class="faint">nothing to run</span>{{ end }}</td>
			<td class="mono-break">{{ with .Checkout }}{{ .Remote }} <span class="badge">{{ .Branch }}</span>{{ else }}<span class="faint">local folder</span>{{ end }}</td>
			<td class="desc hide-sm">{{ with .Checkout }}<span class="faint">{{ .Commit }}</span> {{ .Subject }}{{ if not .When.IsZero }} <span class="faint">{{ ago .When }}</span>{{ end }}{{ end }}</td>
			<td class="actions">
				{{- if .Checkout }}
				<form method="post" action="/settings/projects/delete" data-confirm="Remove the folder {{ .ID }} and its checkout? It can be cloned again from {{ .Checkout.Remote }}." data-confirm-ok="Remove">
					<input type="hidden" name="id" value="{{ .ID }}">
					<button class="btn-icon danger" type="submit" title="Remove the checkout" aria-label="Remove {{ .ID }}"><svg class="ic"><use href="#i-trash"/></svg></button>
				</form>
				{{- end }}
			</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="5">The root is empty. Add a project from a git repository, or put a folder with something to run under {{ $.Root }}.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
<p class="faint" style="margin-top:0.6rem">Root: {{ .Root }}</p>
{{ end }}
