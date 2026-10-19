{{ define "content" }}
{{- $p := .Folder }}
<div class="page-head">
	<div>
		<h2><svg class="ic"><use href="#i-folder-open"/></svg> {{ $p.Title }}</h2>
		{{- if $p.Description }}<p>{{ $p.Description }}</p>{{ end }}
	</div>
	<div class="actions">
		<span class="dim nowrap">{{ range $p.Sources }}<span class="badge">{{ .Kind }}</span> {{ end }}</span>
	</div>
</div>

{{- if $p.Error }}
<div class="alert bad"><b>Could not read a source.</b> {{ $p.Error }}</div>
{{- end }}

{{- range $s := $p.Sources }}
<section class="block" data-live>
	<h3>{{ if eq (len $p.Sources) 1 }}Jobs{{ else }}{{ $s.Kind }}{{ end }} <small>{{ $s.File }}{{ if $s.Command }} · {{ join $s.Command " " }}{{ end }}</small></h3>
	<div class="jobcards">
	{{- range $r := $p.SourceJobs $s }}
		{{- $run := $.RunningRun $r.Key }}{{ $can := $.CanRun $p $r }}
		<article class="jobcard{{ if $run }} running{{ end }}">
			<header>
				<i class="shebang{{ if $run }} running{{ else }}{{ with $.LastRun $r.Key }} {{ .State }}{{ end }}{{ end }}" aria-hidden="true">#!</i>
				<a class="name" href="/f/{{ $p.ID }}/{{ $r.ID }}">{{ $r.Name }}</a>
				{{- if $r.Interactive }}<span class="badge">interactive</span>{{ end }}
			</header>
			{{- if $r.Description }}<p class="desc">{{ $r.Description }}</p>{{ end }}
			<div class="steps">
				{{- if $r.DependsOn }}<span class="dim">after {{ join $r.DependsOn ", " }}</span>{{ end }}
				{{- range $i, $step := $r.Steps }}{{ if lt $i 3 }}<code>{{ $step }}</code>{{ end }}{{ end }}
				{{- if gt (len $r.Steps) 3 }}<span class="faint">+{{ sub (len $r.Steps) 3 }} more</span>{{ end }}
			</div>
			<footer>
				<span class="last">
					{{- if $run }}<span class="state"><i class="dot running"></i>running, {{ $run.UserName }}</span>
					{{- else }}{{ with $.LastRun $r.Key }}{{ template "state" . }} <span class="faint">{{ ago .StartedAt }}</span>{{ else }}<span class="faint">never run</span>{{ end }}{{ end -}}
				</span>
				<span class="actions">
					{{- if $run }}
					<a class="btn btn-sm" href="/runs/{{ $run.ID }}" data-attach="{{ $run.ID }}"><svg class="ic"><use href="#i-eye"/></svg>Attach</a>
					{{- if $can }}<button type="button" class="btn-icon danger" data-kill="{{ $run.ID }}" title="Kill the run" aria-label="Kill the run"><svg class="ic"><use href="#i-stop"/></svg></button>{{ end }}
					{{- else if $can }}{{ template "runButtons" (list $p $r) }}
					{{- else }}<a class="btn btn-sm" href="/f/{{ $p.ID }}/{{ $r.ID }}">History</a>{{ end -}}
				</span>
			</footer>
		</article>
	{{- else }}
		<div class="alert">{{ if $s.Error }}{{ $s.Error }}{{ else }}No targets in {{ $s.File }}.{{ end }}</div>
	{{- end }}
	</div>
</section>
{{- else }}
<div class="alert">Nothing to run was found in this folder: no pipeline file, Makefile, package scripts or scripts folder.</div>
{{- end }}

<section class="block">
	<h3>Files <small>{{ $p.ID }}/</small></h3>
	{{ template "tree" $ }}
</section>

<div class="grid-2">
<section class="block">
	<h3>Folder</h3>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
			<tr><th>Path</th><td><svg class="ic"><use href="#i-folder"/></svg> {{ $p.ID }}/</td></tr>
			{{- range $p.Sources }}<tr><th>{{ .Kind }}</th><td><svg class="ic"><use href="#i-file"/></svg> {{ .File }}{{ if .Command }} <span class="dim">runs with</span> <code>{{ join .Command " " }} &lt;target&gt;</code>{{ end }}{{ if ne .Kind "script" }}{{ template "showsource" (list $p .File) }}{{ end }}</td></tr>{{ end }}
			{{- with $.Checkout }}<tr><th>Source</th><td><div class="mono-break">{{ .Remote }}</div><div class="dim">{{ .Branch }} <span class="faint">{{ .Commit }}</span> {{ .Subject }}{{ if not .When.IsZero }} <span class="faint">{{ ago .When }}</span>{{ end }}</div><div class="faint">pulled before every run</div></td></tr>{{ end }}
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
