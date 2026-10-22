{{ define "content" }}
{{- $p := .Folder }}{{ $default := eq $p.ID "_" }}
<div class="page-head">
	<div>
		<h2><svg class="ic"><use href="#i-folder-open"/></svg> {{ $p.Title }}</h2>
		{{- if $p.Description }}<p>{{ $p.Description }}</p>{{ end }}
	</div>
	<div class="actions">
		<span class="dim nowrap">{{ range $p.Sources }}<span class="badge">{{ .Kind }}</span> {{ end }}</span>
	</div>
</div>
{{- template "actions" $ }}

{{- if $p.Error }}
<div class="alert bad"><b>Could not read a source.</b> {{ $p.Error }}</div>
{{- end }}

{{- template "promptbox" $ }}

{{- if $p.Sources }}
{{- template "sources" (list $ $p) }}
{{- else if and (not $.SubFolders) (not $default) }}
<div class="alert">task-ui found nothing to run in this folder: no pipeline file, Makefile, package scripts or scripts folder.</div>
{{- end }}

{{- if or (not $default) $.Tree }}
<section class="block">
	<h3>Files <small>{{ if $default }}the root{{ else }}{{ $p.ID }}/{{ end }}</small></h3>
	{{ template "tree" $ }}
</section>
{{- end }}

{{- if not $default }}
<div class="grid-2">
<section class="block">
	<h3>Folder</h3>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
			<tr><th>Path</th><td><svg class="ic"><use href="#i-folder"/></svg> {{ $p.ID }}/</td></tr>
			{{- range $p.Sources }}<tr><th>{{ .Kind }}</th><td><svg class="ic"><use href="#i-file"/></svg> {{ .File }}{{ if .Command }} <span class="dim">runs with</span> <code>{{ join .Command " " }} &lt;target&gt;</code>{{ end }}{{ if ne .Kind "script" }}{{ template "showsource" (list $p .File) }}{{ end }}</td></tr>{{ end }}
			{{- with $.Checkout }}<tr><th>Source</th><td><div class="mono-break">{{ .Remote }}</div><div class="dim">{{ .Branch }} <span class="faint">{{ .Commit }}</span> {{ .Subject }}{{ if not .When.IsZero }} <span class="faint">{{ ago .When }}</span>{{ end }}</div><div class="faint">pulled before every run</div></td></tr>{{ end }}
			{{- with $.Space }}<tr><th>Checkout size</th><td>{{ template "spacebar" . }}</td></tr>{{ end }}
			<tr><th>Who can run</th><td>{{ if $p.Allow }}admin, {{ join $p.Allow ", " }}{{ else }}admin, runner{{ end }}</td></tr>
			<tr><th>Hours</th><td>{{ template "hours" (list $p.Hours $) }}</td></tr>
			{{- if $p.Includes }}<tr><th>Includes</th><td>{{ range $p.Includes }}<div><svg class="ic"><use href="#i-file"/></svg> {{ . }}{{ template "showsource" (list $p .) }}</div>{{ end }}</td></tr>{{ end }}
			<tr><th>Scheduled</th><td>{{ if .Schedules }}{{ len .Schedules }} <a href="/schedules?folder={{ $p.ID }}">schedules</a>{{ else }}<span class="faint">none</span>, <a href="/schedules?folder={{ $p.ID }}">add one</a>{{ end }}</td></tr>
		</tbody>
	</table>
	</div>
</section>

<section class="block">
	<h3>Environment <small>dotenv files the sources read</small></h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>File</th><th>Variables</th><th>State</th>{{ if $.Can "env/view" }}<th></th>{{ end }}</tr></thead>
		<tbody>
		{{- range .EnvFiles }}
			<tr>
				<td class="key"><svg class="ic"><use href="#i-file"/></svg> {{ .Path }}</td>
				<td class="desc">{{ if .Keys }}{{ join .Keys ", " }}{{ else }}<span class="faint">none</span>{{ end }}</td>
				<td>{{ if .Managed }}<span class="badge ok">managed</span>{{ else if .Exists }}<span class="badge">in folder</span>{{ else }}<span class="badge bad">missing</span>{{ end }}</td>
				{{- if $.Can "env/view" }}
				<td class="actions"><a class="btn-icon" href="{{ $.Href $p.ID }}/env/{{ .Path }}" title="Edit {{ .Path }}" aria-label="Edit {{ .Path }}"><svg class="ic"><use href="#i-edit"/></svg></a></td>
				{{- end }}
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="4">No source uses an env file. Runs still get secrets through their environment.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
</div>
{{- end }}
{{ end }}
