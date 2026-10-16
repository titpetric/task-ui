{{ define "content" }}
{{- $p := .Project }}{{ $r := .Job }}{{ $run := running $r.Key }}{{ $can := canRun $p $r }}
{{ template "crumbs" (list (crumb $p.ID (print "/f/" $p.ID)) (crumb $r.Name "")) }}
<div class="page-head">
	<div>
		<h2>{{ $r.Name }} <span class="badge">{{ $r.Kind }}</span>{{ if $r.Interactive }} <span class="badge">interactive</span>{{ end }}</h2>
		<p>{{ $r.Description }}</p>
	</div>
	<div class="actions">
		{{- if $can }}{{ template "runButtons" (list $p $r) }}{{ end }}
	</div>
</div>

{{- if $run }}
<div class="alert signal" data-running-alert>
	<span class="state"><i class="dot running"></i><b>Running</b></span>, started by <b>{{ $run.User }}</b> {{ ago $run.Started }}.
	{{- if $can }} Run force stops it and starts again.{{ end }}
</div>
{{- end }}
{{- if not $can }}
<div class="alert">You can read this job and its history. Your role, <b>{{ .User.Role }}</b>, is not allowed to run it.</div>
{{- end }}

{{- if $r.Summary }}
<pre class="summary">{{ $r.Summary }}</pre>
{{- end }}

<section class="block">
	<h3>Job</h3>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
			<tr><th>Folder</th><td><a href="/f/{{ $p.ID }}"><svg class="ic"><use href="#i-folder"/></svg> {{ $p.ID }}</a> <span class="dim">{{ $p.Title }}</span></td></tr>
			<tr><th>Source</th><td><svg class="ic"><use href="#i-file"/></svg> {{ $r.Source }}</td></tr>
			<tr><th>Command</th><td><code>{{ join $r.Command " " }}</code></td></tr>
			{{- if $r.DependsOn }}<tr><th>Depends on</th><td>{{ range $r.DependsOn }}{{ with $p.Job . }}<a href="/f/{{ $p.ID }}/{{ .ID }}">{{ .Name }}</a>{{ else }}{{ . }}{{ end }} {{ end }}</td></tr>{{ end }}
			<tr><th>Steps</th><td class="steps">{{ range $r.Steps }}<code>{{ . }}</code>{{ else }}<span class="faint">none listed</span>{{ end }}</td></tr>
			{{- if $r.Aliases }}<tr><th>Aliases</th><td>{{ range $r.Aliases }}<span class="badge">{{ . }}</span> {{ end }}</td></tr>{{ end }}
			{{- if $r.EnvFiles }}<tr><th>Env files</th><td>{{ range $r.EnvFiles }}<div><svg class="ic"><use href="#i-file"/></svg> {{ . }}</div>{{ end }}</td></tr>{{ end }}
			<tr><th>Who can run</th><td>{{ whoCanRun $p $r }}</td></tr>
			<tr><th>Input</th><td>{{ if $r.Interactive }}keyboard input is sent to the job{{ else }}output only{{ end }}</td></tr>
		</tbody>
	</table>
	</div>
</section>

<section class="block" data-live>
	<h3>History{{ if .History }} <small>{{ len .History }} runs</small>{{ end }}</h3>
	{{- if not .HistoryEnabled }}
	<div class="alert">History is off. Set TASK_UI_HISTORY=true to record runs.</div>
	{{- else }}
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Started</th><th>User</th><th>Duration</th><th>Result</th><th></th></tr></thead>
		<tbody>
		{{- range .History }}
			<tr>
				<td class="nowrap" title="{{ ago .Started }}">{{ stamp .Started }}</td>
				<td>{{ .User }}{{ if .Forced }} <span class="badge signal">forced</span>{{ end }}</td>
				<td class="nowrap">{{ if ne .State "running" }}{{ duration .Duration }}{{ end }}</td>
				<td>{{ template "state" . }}</td>
				<td class="actions">
					<button type="button" class="btn btn-sm" data-history="/api/history/{{ .Project }}/{{ .Job }}/{{ .ID }}" data-title="{{ $p.ID }}" data-job="{{ .Job }}" data-when="{{ stamp .Started }}">View</button>
					<button type="button" class="btn btn-sm" data-history="/api/history/{{ .Project }}/{{ .Job }}/{{ .ID }}" data-replay data-title="{{ $p.ID }}" data-job="{{ .Job }}" data-when="{{ stamp .Started }}">Replay</button>
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="5">No runs yet.{{ if $can }} Press Run to start the first one.{{ end }}</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>
{{ end }}
