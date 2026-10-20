{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Projects</h2>
		<p>The folders under the root. A folder with something to run shows up in the sidebar; one cloned from a git repository is pulled before every run.</p>
	</div>
	<div class="actions">
		<a class="btn btn-primary" href="/settings/projects/new">Add project</a>
	</div>
</div>

<div class="table-wrap">
<table class="table">
	<thead><tr><th>Folder</th><th>Sources</th><th>Origin</th><th class="hide-sm">Last commit</th>{{ if $.Spaces }}<th class="hide-sm">Size</th>{{ end }}<th></th></tr></thead>
	<tbody>
	{{- range .Entries }}
		<tr>
			<td class="key nowrap">{{ if .Sources }}<a href="/f/{{ .ID }}"><svg class="ic"><use href="#i-folder"/></svg> {{ .ID }}</a>{{ else }}<svg class="ic"><use href="#i-folder"/></svg> {{ .ID }}{{ end }}</td>
			<td class="nowrap">{{ if .Error }}<span class="badge bad">error</span> {{ end }}{{ if .Sources }}{{ join .Sources ", " }} <span class="faint">{{ .Jobs }} jobs</span>{{ else }}<span class="faint">nothing to run</span>{{ end }}</td>
			<td class="mono-break">{{ with .Checkout }}{{ .Remote }} <span class="badge">{{ .Branch }}</span>{{ else }}<span class="faint">local folder</span>{{ end }}</td>
			<td class="desc hide-sm">{{ with .Checkout }}<span class="faint">{{ .Commit }}</span> {{ .Subject }}{{ if not .When.IsZero }} <span class="faint">{{ ago .When }}</span>{{ end }}{{ end }}</td>
			{{- if $.Spaces }}<td class="hide-sm">{{ template "spacemini" index $.Spaces .ID }}</td>{{ end }}
			<td class="actions">
				<a class="btn-icon" href="/settings/projects/{{ .ID }}" title="Configure {{ .ID }}" aria-label="Configure {{ .ID }}"><svg class="ic"><use href="#i-edit"/></svg></a>
				{{- if .Checkout }}
				<button type="button" class="btn-icon danger" data-delete="/api/projects/{{ .ID }}" data-confirm="Remove the folder {{ .ID }} and its checkout? It can be cloned again from {{ .Checkout.Remote }}." data-confirm-ok="Remove" data-next="reload" title="Remove the checkout" aria-label="Remove {{ .ID }}"><svg class="ic"><use href="#i-trash"/></svg></button>
				{{- end }}
			</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="6">The root is empty. Add a project from a git repository, or put a folder with something to run under {{ $.Root }}.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
<p class="faint" style="margin-top:0.6rem">Root: {{ .Root }}</p>
{{ end }}
