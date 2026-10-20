{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Artifacts</h2>
		<p>The files the runs leave in <code>$TASK_UI_ARTIFACTS</code>, gathered into one tree: a run that writes a file again replaces it, and the history keeps the version before.</p>
	</div>
	{{- if .User.CanAdmin }}
	<div class="actions">
		<button type="button" class="btn btn-danger" data-delete="/api/artifacts" data-confirm="Delete all the artifacts? Every file, its history and the record of every version go." data-confirm-ok="Delete all" data-next="reload"><svg class="ic"><use href="#i-trash"/></svg>Delete all</button>
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
	<div class="block-head">
		<h3>Files <small>/{{ .Browse }}</small></h3>
		<div class="actions">
			<a class="btn btn-sm" href="{{ .Archive }}" download><svg class="ic"><use href="#i-download"/></svg>{{ if .Browse }}This folder{{ else }}Everything{{ end }} as .tgz</a>
			<button type="button" class="btn-icon" data-copy="{{ .Archive }}" title="Copy the link to the .tgz" aria-label="Copy the link to the .tgz"><svg class="ic"><use href="#i-copy"/></svg></button>
		</div>
	</div>
	<div class="table-wrap">
	<table class="table artifact-files">
		<thead><tr><th>Name</th><th class="num">Size</th><th class="hide-sm">Changed</th><th></th></tr></thead>
		<tbody>
		{{- range .Tree }}
			<tr>
				<td class="key"><svg class="ic{{ if .Dir }} dir{{ end }}"><use href="#i-{{ if .Dir }}folder{{ else }}file{{ end }}"/></svg>
					{{- if .Dir }} <a href="{{ .Href }}">{{ .Name }}/</a>
					{{- else }} <a href="{{ .Href }}" data-artifact="{{ .Path }}">{{ .Name }}</a>{{ end }}</td>
				<td class="num nowrap">{{ size .Size }}</td>
				<td class="hide-sm nowrap" title="{{ stamp .ModTime }}">{{ ago .ModTime }}</td>
				<td class="actions">
					{{- if .Dir }}
					<a class="btn-icon" href="{{ .Archive }}" download title="Download {{ .Name }} as .tgz" aria-label="Download {{ .Name }} as .tgz"><svg class="ic"><use href="#i-download"/></svg></a>
					<button type="button" class="btn-icon" data-copy="{{ .Archive }}" title="Copy the link to the .tgz" aria-label="Copy the link to {{ .Name }} as .tgz"><svg class="ic"><use href="#i-copy"/></svg></button>
					{{- else }}
					<a class="btn-icon" href="{{ .Href }}" data-artifact="{{ .Path }}" title="View {{ .Name }} and its versions" aria-label="View {{ .Name }}"><svg class="ic"><use href="#i-eye"/></svg></a>
					<a class="btn-icon" href="{{ .Href }}?download" title="Download {{ .Name }}" aria-label="Download {{ .Name }}"><svg class="ic"><use href="#i-download"/></svg></a>
					<button type="button" class="btn-icon" data-copy="{{ .Href }}" title="Copy the link to {{ .Name }}" aria-label="Copy the link to {{ .Name }}"><svg class="ic"><use href="#i-copy"/></svg></button>
					{{- end }}
					{{- if $.User.CanAdmin }}
					<button type="button" class="btn-icon danger" data-delete="/api{{ if .Dir }}/artifacts/files/{{ .Path }}{{ else }}{{ .Href }}{{ end }}" data-confirm="Delete {{ .Path }}{{ if .Dir }}/ and everything in it{{ end }} from the tree? The history keeps its versions." data-confirm-ok="Delete" data-next="reload" title="Delete from the tree" aria-label="Delete {{ .Name }}"><svg class="ic"><use href="#i-trash"/></svg></button>
					{{- end }}
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="4">Empty.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
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
				<td class="key">{{ if .CommitID }}<a href="/artifacts/files/{{ .Path }}?version={{ .CommitID }}" data-artifact="{{ .Path }}" data-version="{{ .CommitID }}" title="View this version">{{ .Name }}</a>{{ else }}<a href="/api/artifacts/{{ .ID }}">{{ .Name }}</a>{{ end }}</td>
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
