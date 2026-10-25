{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Artifacts</h2>
		<p>Files that runs write to <code>$TASK_UI_ARTIFACTS</code>, collected into one tree. When a run writes the same file again, the new file replaces it, and the history keeps the earlier version.</p>
	</div>
	{{- if .Can "artifact/delete" }}
	<div class="actions">
		{{ template "menu" (menu "Actions on every artifact"
			(danger "trash" "Delete all" "data-delete" "/api/artifacts" "data-confirm" "Delete all artifacts? This removes every file, its history and the record of every version." "data-confirm-ok" "Delete all" "data-next" "reload")
		) }}
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
	<table class="table file-table">
		<thead><tr><th>Name</th><th class="num">Size</th><th class="hide-sm">Changed</th><th></th></tr></thead>
		<tbody>
		{{- range .Tree }}
			<tr>
				<td class="key"><svg class="ic{{ if .Dir }} dir{{ end }}"><use href="#i-{{ if .Dir }}folder{{ else }}file{{ end }}"/></svg>
					{{- if .Dir }} <a href="{{ .Href }}">{{ .Name }}/</a>
					{{- else }} <a href="{{ .Href }}" data-artifact="{{ .Path }}">{{ .Name }}</a>{{ end }}</td>
				<td class="num nowrap">{{ size .Size }}</td>
				<td class="hide-sm nowrap" title="{{ stamp .ModTime }}">{{ ago .ModTime }}</td>
				<td class="actions">{{ template "menu" (menu (print "Actions on " .Name)
					(and (not .Dir) (item "eye" "View and versions" "href" .Href "data-artifact" .Path))
					(and (not .Dir) (item "download" "Download" "href" (print .Href "?download")))
					(and (not .Dir) (item "copy" "Copy the link" "data-copy" .Href))
					(and .Dir (item "download" "Download as .tgz" "href" .Archive "download" ""))
					(and .Dir (item "copy" "Copy the link to the .tgz" "data-copy" .Archive))
					(and ($.Can "artifact/edit") (danger "trash" "Delete from the tree" "data-delete" (print "/api" (or (and .Dir (print "/artifacts/files/" .Path)) .Href)) "data-confirm" (print "Delete " .Path (or (and .Dir "/ and everything in it") "") " from the tree? The history keeps its versions.") "data-confirm-ok" "Delete" "data-next" "reload"))
				) }}</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="4">This folder is empty.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>

<section class="block">
	<h3>History</h3>
	<form method="get" action="/artifacts" class="filters">
		<label for="art-folder">Folder
			<select id="art-folder" name="folder">
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
		{{- range .Artifacts }}{{ $a := . }}
			<tr>
				<td class="nowrap" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</td>
				{{- if .RunID }}
				<td>{{ with $.RunFolderHref .Folder "" }}<a href="{{ . }}">{{ $a.Folder }}</a>{{ else }}{{ .Folder }}{{ end }}</td>
				<td>{{ with $.RunJobHref .Folder "" .Job }}<a href="{{ . }}">{{ $a.Job }}</a>{{ else }}{{ .Job }}{{ end }}</td>
				<td><a class="faint mono" href="/runs/{{ .RunID }}" title="{{ .RunID }}">{{ slice .RunID 0 8 }}</a></td>
				{{- else }}
				<td colspan="3" class="dim">written over MCP or the API by {{ template "user" (list $ .UserName) }}</td>
				{{- end }}
				<td class="key">{{ if .CommitID }}<a href="/artifacts/files/{{ .Path }}?version={{ .CommitID }}" data-artifact="{{ .Path }}" data-version="{{ .CommitID }}" title="View this version">{{ .Name }}</a>{{ else }}<a href="/api/artifacts/{{ .ID }}">{{ .Name }}</a>{{ end }}</td>
				<td class="hide-sm">{{ if .CommitID }}<span class="faint mono" title="{{ .CommitID }}">{{ slice .CommitID 0 7 }}</span>{{ else }}<span class="faint">kept apart</span>{{ end }}</td>
				<td class="nowrap">{{ size .ByteSize }}</td>
				<td class="actions">{{ template "menu" (menu (print "Actions on " .Name)
					(and .CommitID (detail "Version" .CommitID "mono" ""))
					(and .RunID (detail "Run" .RunID "mono" ""))
					(detail "Created" (stamp .CreatedAt))
					(and .CommitID (item "eye" "View this version" "href" (print "/artifacts/files/" .Path "?version=" .CommitID) "data-artifact" .Path "data-version" .CommitID))
					(item "download" "Download this version" "href" (print "/api/artifacts/" .ID))
					(and ($.Can "artifact/delete") (danger "trash" "Delete the record" "data-delete" (print "/api/artifacts/" .ID) "data-confirm" (print "Delete the record of " .Name (or (and .RunID (print " from " .Folder " / " .Job)) "") "?") "data-confirm-ok" "Delete" "data-next" "reload"))
				) }}</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="8">No artifacts yet. Files a job writes to <code>$TASK_UI_ARTIFACTS</code> appear here.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{ end }}
