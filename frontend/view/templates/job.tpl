{{ define "content" }}
{{- $p := .Folder }}{{ $r := .Job }}{{ $run := .RunningRun $r.Key }}{{ $can := .CanRun $p $r }}
<div class="page-head">
	<div>
		<h2>{{ $r.Name }} <span class="badge">{{ $r.Kind }}</span>{{ if $r.Interactive }} <span class="badge">interactive</span>{{ end }}</h2>
		<p>{{ $r.Description }}</p>
	</div>
	<div class="actions">
		{{- if $run }}
		<a class="btn btn-primary" href="/runs/{{ $run.ID }}" data-attach="{{ $run.ID }}">Attach</a>
		{{- if $can }}<button type="button" class="btn btn-danger" data-kill="{{ $run.ID }}">Kill</button>{{ end }}
		{{- else if $can }}{{ template "runButtons" (list $p $r) }}{{ end }}
	</div>
</div>

{{- if $run }}
<div class="alert signal">
	<span class="state"><i class="dot running"></i><b>Running</b></span>, started by <b>{{ $run.UserName }}</b> {{ ago $run.StartedAt }}{{ if $run.Attached }}, {{ $run.Attached }} watching{{ end }}.
	{{- if $can }} Attach to follow it, or Run force to kill it and start again.{{ else }} Attach to follow it.{{ end }}
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
			<tr><th>Who can run</th><td>{{ .WhoCanRun $p $r }}</td></tr>
			<tr><th>Input</th><td>{{ if $r.Interactive }}keyboard input is sent to the job{{ else }}output only{{ end }}</td></tr>
		</tbody>
	</table>
	</div>
</section>

<section class="block" data-live>
	<h3>History{{ if .Runs }} <small>{{ len .Runs }} runs</small>{{ end }}</h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Started</th><th>By</th><th>Duration</th><th>Result</th><th></th></tr></thead>
		<tbody>
		{{- range .Runs }}
			<tr>
				<td class="nowrap" title="{{ ago .StartedAt }}">{{ stamp .StartedAt }}</td>
				<td>{{ template "who" . }}</td>
				<td class="nowrap">{{ duration .Elapsed }}</td>
				<td>{{ template "state" . }}</td>
				<td class="actions">
					<a class="btn btn-sm" href="/runs/{{ .ID }}" data-attach="{{ .ID }}">{{ if eq .State "running" }}Attach{{ else }}View{{ end }}</a>
					{{- if and (eq .State "running") $can }}<button type="button" class="btn-icon danger" data-kill="{{ .ID }}" title="Kill the run" aria-label="Kill the run"><svg class="ic"><use href="#i-stop"/></svg></button>{{ end }}
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="5">No runs yet.{{ if $can }} Press Run to start the first one.{{ end }}</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>

<section class="block">
	<h3>Scheduled <small>cron expressions, in the server's time zone</small></h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Expression</th><th>Next run</th><th>Last run</th><th>By</th><th>State</th><th></th></tr></thead>
		<tbody>
		{{- range .Schedules }}
			<tr>
				<td class="key"><code>{{ .Expression }}</code></td>
				<td class="nowrap">{{ if .IsEnabled }}{{ stamp .NextRunAt }}{{ else }}<span class="faint">-</span>{{ end }}</td>
				<td class="nowrap">{{ if .LastRunAt }}{{ ago .LastRunAt }}{{ else }}<span class="faint">never</span>{{ end }}</td>
				<td>{{ .UserName }}</td>
				<td>{{ if .IsEnabled }}<span class="badge ok">enabled</span>{{ else }}<span class="badge">disabled</span>{{ end }}</td>
				<td class="actions">
					{{- if $can }}
					<form data-api="/api/schedules/{{ .ID }}" data-method="PUT" data-next="reload"><input type="hidden" name="enabled" data-type="bool" value="{{ if .IsEnabled }}false{{ else }}true{{ end }}"><button class="btn btn-sm" type="submit">{{ if .IsEnabled }}Disable{{ else }}Enable{{ end }}</button></form>
					<button type="button" class="btn-icon danger" data-delete="/api/schedules/{{ .ID }}" data-confirm="Delete the schedule {{ .Expression }} for {{ $r.Name }}?" data-next="reload" title="Delete" aria-label="Delete schedule"><svg class="ic"><use href="#i-trash"/></svg></button>
					{{- end }}
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="6">Not scheduled.{{ if $can }} Add an expression below to run it on a schedule, under your name.{{ end }}</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- if $can }}
	<form class="form inline" data-api="/api/schedules" data-method="POST" data-next="reload">
		<input type="hidden" name="folder" value="{{ $p.ID }}">
		<input type="hidden" name="job" value="{{ $r.ID }}">
		<div class="field">
			<label for="sch-expr">Expression</label>
			<input type="text" id="sch-expr" name="expression" placeholder="0 3 * * *  or  @daily" spellcheck="false" autocomplete="off" required>
		</div>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Schedule</button>
		</div>
	</form>
	{{- end }}
</section>
{{ end }}
