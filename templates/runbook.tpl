{{ define "content" }}
{{- $p := .Project }}{{ $r := .Runbook }}{{ $run := running $r.Key }}{{ $can := canRun $p $r }}
<div class="crumbs"><a href="/">Overview</a><span><a href="/p/{{ $p.ID }}">{{ $p.Title }}</a></span><span>{{ $r.Name }}</span></div>
<div class="page-head">
	<div>
		<h2>{{ $r.Name }}{{ if $r.Interactive }} <span class="badge">interactive</span>{{ end }}</h2>
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
<div class="alert">You can read this runbook and its history. Your role, <b>{{ .User.Role }}</b>, is not allowed to run it.</div>
{{- end }}

{{- if $r.Summary }}
<pre class="summary">{{ $r.Summary }}</pre>
{{- end }}

<section class="block">
	<h3>Details</h3>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
			<tr><th>Project</th><td><a href="/p/{{ $p.ID }}">{{ $p.Title }}</a></td></tr>
			<tr><th>Task file</th><td>{{ $p.File }} <span class="badge">{{ $p.Kind }}</span></td></tr>
			<tr><th>Command</th><td><code>{{ range $p.Command }}{{ . }} {{ end }}{{ $r.Name }}</code></td></tr>
			{{- if $r.Aliases }}<tr><th>Aliases</th><td>{{ range $r.Aliases }}<span class="badge">{{ . }}</span> {{ end }}</td></tr>{{ end }}
			<tr><th>Who can run</th><td>{{ whoCanRun $p $r }}</td></tr>
			<tr><th>Input</th><td>{{ if $r.Interactive }}keyboard input is sent to the task{{ else }}output only{{ end }}</td></tr>
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
					<button type="button" class="btn btn-sm" data-history="/api/history/{{ .Project }}/{{ .Runbook }}/{{ .ID }}" data-title="{{ $p.Title }}" data-runbook="{{ .Runbook }}" data-when="{{ stamp .Started }}">View</button>
					<button type="button" class="btn btn-sm" data-history="/api/history/{{ .Project }}/{{ .Runbook }}/{{ .ID }}" data-replay data-title="{{ $p.Title }}" data-runbook="{{ .Runbook }}" data-when="{{ stamp .Started }}">Replay</button>
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
