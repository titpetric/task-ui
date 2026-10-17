{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Artifacts</h2>
		<p>Files the runs left behind. A job writes them to the directory named in <code>$TASK_UI_ARTIFACTS</code>, and they are collected when it ends.</p>
	</div>
</div>

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
	<thead><tr><th>Created</th><th>Folder</th><th>Job</th><th>Run</th><th>Name</th><th>Size</th><th></th></tr></thead>
	<tbody>
	{{- range .Artifacts }}
		<tr>
			<td class="nowrap" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</td>
			<td><a href="/f/{{ .Folder }}">{{ .Folder }}</a></td>
			<td><a href="/f/{{ .Folder }}/{{ .Job }}">{{ .Job }}</a></td>
			<td><a class="faint" href="/runs/{{ .RunID }}">{{ .RunID }}</a></td>
			<td class="key"><a href="/api/artifacts/{{ .ID }}">{{ .Name }}</a></td>
			<td class="nowrap">{{ size .ByteSize }}</td>
			<td class="actions">
				<a class="btn-icon" href="/api/artifacts/{{ .ID }}" title="Download" aria-label="Download {{ .Name }}"><svg class="ic"><use href="#i-download"/></svg></a>
				{{- if $.User.CanAdmin }}<button type="button" class="btn-icon danger" data-delete="/api/artifacts/{{ .ID }}" data-confirm="Delete {{ .Name }} of {{ .Folder }} / {{ .Job }}?" data-next="reload" title="Delete" aria-label="Delete {{ .Name }}"><svg class="ic"><use href="#i-trash"/></svg></button>{{ end }}
			</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="7">No artifacts yet. A job that writes to <code>$TASK_UI_ARTIFACTS</code> leaves its files here.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
{{ end }}
