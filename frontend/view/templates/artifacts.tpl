{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Artifacts</h2>
		<p>The files the runs leave in <code>$TASK_UI_ARTIFACTS</code>, gathered into one tree: a run that writes a file again replaces it, and the history keeps the version before.</p>
	</div>
	{{- if .User.CanAdmin }}
	<div class="actions">
		<button type="button" class="btn btn-danger" data-delete="/api/artifacts" data-confirm="Empty the artifacts? Every file, its history and the record of every version go." data-confirm-ok="Empty" data-next="reload">Empty</button>
	</div>
	{{- end }}
</div>

{{- with .Space }}
<section class="block">
	<h3>Space</h3>
	{{ template "spacebar" . }}
</section>
{{- end }}

<section class="block">
	<h3>Files{{ if .Browse }} <small>{{ .Browse }}</small>{{ end }}</h3>
	{{ template "tree" . }}
</section>

<section class="block">
	<h3>History</h3>
	<form method="get" action="/artifacts" class="filters">
		<label>Folder
			<select name="folder">
				<option value="">all folders</option>
				{{- range .Folders }}<option value="{{ .ID }}"{{ if eq (index $.Form "folder") .ID }} selected{{ end }}>{{ .ID }}</option>{{ end }}
			</select>
		</label>
		<label>Job <input type="text" name="job" value="{{ index .Form "job" }}" spellcheck="false" autocomplete="off"></label>
		<button class="btn" type="submit">Filter</button>
	</form>

	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Created</th><th>Folder</th><th>Job</th><th>Run</th><th>Name</th><th class="hide-sm">Version</th><th>Size</th><th></th></tr></thead>
		<tbody>
		{{- range .Artifacts }}
			<tr>
				<td class="nowrap" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</td>
				<td><a href="/f/{{ .Folder }}">{{ .Folder }}</a></td>
				<td><a href="/f/{{ .Folder }}/{{ .Job }}">{{ .Job }}</a></td>
				<td><a class="faint mono" href="/runs/{{ .RunID }}" title="{{ .RunID }}">{{ slice .RunID 0 8 }}</a></td>
				<td class="key"><a href="/api/artifacts/{{ .ID }}">{{ .Name }}</a></td>
				<td class="hide-sm">{{ if .CommitID }}<span class="faint mono" title="{{ .CommitID }}">{{ slice .CommitID 0 7 }}</span>{{ else }}<span class="faint">kept apart</span>{{ end }}</td>
				<td class="nowrap">{{ size .ByteSize }}</td>
				<td class="actions">
					<a class="btn-icon" href="/api/artifacts/{{ .ID }}" title="Download this version" aria-label="Download {{ .Name }}"><svg class="ic"><use href="#i-download"/></svg></a>
					{{- if $.User.CanAdmin }}<button type="button" class="btn-icon danger" data-delete="/api/artifacts/{{ .ID }}" data-confirm="Delete the record of {{ .Name }} from {{ .Folder }} / {{ .Job }}?" data-next="reload" title="Delete" aria-label="Delete {{ .Name }}"><svg class="ic"><use href="#i-trash"/></svg></button>{{ end }}
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="8">No artifacts yet. A job that writes to <code>$TASK_UI_ARTIFACTS</code> leaves its files here.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{ end }}
