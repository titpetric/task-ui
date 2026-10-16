{{ define "content" }}
{{- $p := .Project }}
{{ template "crumbs" (list (crumb $p.ID "")) }}
<div class="page-head">
	<div>
		<h2><svg class="ic"><use href="#i-folder-open"/></svg> {{ $p.Title }}</h2>
		{{- if $p.Description }}<p>{{ $p.Description }}</p>{{ end }}
	</div>
	<div class="dim nowrap">{{ range $p.Sources }}<span class="badge">{{ .Kind }}</span> {{ end }}</div>
</div>

{{- if $p.Error }}
<div class="alert bad"><b>Could not read a source.</b> {{ $p.Error }}</div>
{{- end }}

{{- range $s := $p.Sources }}
<section class="block" data-live>
	<h3>{{ if eq (len $p.Sources) 1 }}Jobs{{ else }}{{ $s.Kind }}{{ end }} <small>{{ $s.File }}{{ if $s.Command }} · {{ join $s.Command " " }}{{ end }}</small></h3>
	<div class="table-wrap">
	<table class="table jobs">
		<thead><tr><th>Name</th><th>Description</th><th class="hide-sm steps">Steps</th><th class="hide-sm">Last run</th><th></th></tr></thead>
		<tbody>
		{{- range $r := $s.Jobs }}
			{{- $run := running $r.Key }}
			<tr>
				<td class="key"><a href="/f/{{ $p.ID }}/{{ $r.ID }}">{{ $r.Name }}</a>{{ if $r.Interactive }} <span class="badge">interactive</span>{{ end }}</td>
				<td class="desc">{{ $r.Description }}</td>
				<td class="hide-sm steps">
					{{- if $r.DependsOn }}<span class="dim">after {{ join $r.DependsOn ", " }}</span>{{ end }}
					{{- range $r.Steps }}<code>{{ . }}</code>{{ end }}
				</td>
				<td class="nowrap hide-sm">
					{{- if $run }}<span class="state"><i class="dot running"></i>running, {{ $run.User }}</span>
					{{- else }}{{ with last $p.ID $r.ID }}{{ template "state" . }} <span class="faint">{{ ago .Started }}</span>{{ else }}<span class="faint">never run</span>{{ end }}{{ end -}}
				</td>
				<td class="actions">
					{{- if canRun $p $r }}{{ template "runButtons" (list $p $r) }}{{ else }}<a class="btn btn-sm" href="/f/{{ $p.ID }}/{{ $r.ID }}">History</a>{{ end -}}
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="5">{{ if $s.Error }}{{ $s.Error }}{{ else }}No targets in {{ $s.File }}.{{ end }}</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{- else }}
<div class="alert">Nothing to run was found in this folder: no pipeline file, Makefile, package scripts or scripts folder.</div>
{{- end }}

<div class="grid-2">
<section class="block">
	<h3>Folder</h3>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
			<tr><th>Path</th><td><svg class="ic"><use href="#i-folder"/></svg> {{ $p.ID }}/</td></tr>
			{{- range $p.Sources }}<tr><th>{{ .Kind }}</th><td><svg class="ic"><use href="#i-file"/></svg> {{ .File }}{{ if .Command }} <span class="dim">runs with</span> <code>{{ join .Command " " }} &lt;target&gt;</code>{{ end }}</td></tr>{{ end }}
			{{- with $.Git }}<tr><th>Source</th><td><div class="mono-break">{{ .Remote }}</div><div class="dim">{{ .Branch }} <span class="faint">{{ .Commit }}</span> {{ .Subject }}{{ if not .When.IsZero }} <span class="faint">{{ ago .When }}</span>{{ end }}</div><div class="faint">pulled before every run</div></td></tr>{{ end }}
			<tr><th>Who can run</th><td>{{ if $p.Allow }}admin, {{ join $p.Allow ", " }}{{ else }}admin, runner{{ end }}</td></tr>
			{{- if $p.Includes }}<tr><th>Includes</th><td>{{ range $p.Includes }}<div><svg class="ic"><use href="#i-file"/></svg> {{ . }}</div>{{ end }}</td></tr>{{ end }}
		</tbody>
	</table>
	</div>
</section>

<section class="block">
	<h3>Environment <small>dotenv files the sources read</small></h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>File</th><th>Variables</th><th>State</th>{{ if $.User.CanAdmin }}<th></th>{{ end }}</tr></thead>
		<tbody>
		{{- range .EnvFiles }}
			<tr>
				<td class="key"><svg class="ic"><use href="#i-file"/></svg> {{ .Path }}</td>
				<td class="desc">{{ if .Keys }}{{ join .Keys ", " }}{{ else }}<span class="faint">none</span>{{ end }}</td>
				<td>{{ if .Managed }}<span class="badge ok">managed</span>{{ else if .Exists }}<span class="badge">in folder</span>{{ else }}<span class="badge bad">missing</span>{{ end }}</td>
				{{- if $.User.CanAdmin }}
				<td class="actions"><a class="btn-icon" href="/f/{{ $p.ID }}/env/{{ .Path }}" title="Edit {{ .Path }}" aria-label="Edit {{ .Path }}"><svg class="ic"><use href="#i-edit"/></svg></a></td>
				{{- end }}
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="4">No source declares an env file. Secrets still reach the runs through the environment.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
</div>
{{ end }}
