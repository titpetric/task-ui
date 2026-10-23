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
		{{- else if $can }}{{ template "runButtons" (list $ $p $r) }}{{ end }}
	</div>
</div>

{{- if $run }}
<div class="alert signal">
	<span class="state"><i class="dot running"></i><b>Running</b></span>, started by <b>{{ $run.UserName }}</b> {{ ago $run.StartedAt }}{{ if $run.Attached }}, {{ $run.Attached }} watching{{ end }}.
	{{- if $can }} Attach to follow it, or kill it.{{ else }} Attach to follow it.{{ end }}
</div>
{{- end }}
{{- if not $can }}
<div class="alert">You can see this job and its history, but you don't have permission to run it.</div>
{{- end }}

{{- if $r.Summary }}
<pre class="summary">{{ $r.Summary }}</pre>
{{- end }}

<section class="block">
	<h3>Job</h3>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
			<tr><th>Folder</th><td><a href="{{ $.Href $p.ID }}"><svg class="ic"><use href="#i-folder"/></svg> {{ $p.ID }}</a> <span class="dim">{{ $p.Title }}</span></td></tr>
			{{- with $r.Dir }}<tr><th>Directory</th><td><a href="{{ $.Href $p.ID }}/{{ . }}"><svg class="ic"><use href="#i-folder-open"/></svg> {{ . }}/</a> <span class="dim">the job runs here</span></td></tr>{{ end }}
			<tr><th>Source</th><td><svg class="ic"><use href="#i-file"/></svg> {{ $r.Source }}{{ if and (ne $r.Kind "script") (ne $r.Kind "skill") (not $r.Dir) }}{{ template "showsource" (list $p $r.Source) }}{{ end }}</td></tr>
			{{- if $p.Includes }}<tr><th>Includes</th><td>{{ range $p.Includes }}<div><svg class="ic"><use href="#i-file"/></svg> {{ . }}{{ template "showsource" (list $p .) }}</div>{{ end }}</td></tr>{{ end }}
			<tr><th>Command</th><td><code>{{ join $r.Command " " }}</code></td></tr>
			{{- if $r.DependsOn }}<tr><th>Depends on</th><td>{{ range $r.DependsOn }}{{ with $p.Job . }}<a href="{{ $.JobHref $p . }}">{{ .Name }}</a>{{ else }}{{ . }}{{ end }} {{ end }}</td></tr>{{ end }}
			<tr><th>Steps</th><td class="steps">{{ range $r.Steps }}<code>{{ . }}</code>{{ else }}<span class="faint">none listed</span>{{ end }}</td></tr>
			{{- if $r.Aliases }}<tr><th>Aliases</th><td>{{ range $r.Aliases }}<span class="badge">{{ . }}</span> {{ end }}</td></tr>{{ end }}
			{{- if $r.EnvFiles }}<tr><th>Env files</th><td>{{ range $r.EnvFiles }}<div><svg class="ic"><use href="#i-file"/></svg> {{ . }}</div>{{ end }}</td></tr>{{ end }}
			<tr><th>Who can run</th><td>{{ .WhoCanRun $p $r }}</td></tr>
			<tr><th>Hours</th><td>{{ template "hours" (list ($p.HoursOf $r) $) }}{{ if and $r.Hours $p.Hours }} <span class="faint">(the job's own; the folder runs {{ hoursLine $p.Hours }})</span>{{ end }}</td></tr>
			<tr><th>Input</th><td>{{ if $r.Interactive }}the job takes keyboard input{{ else }}output only{{ end }}</td></tr>
		</tbody>
	</table>
	</div>
</section>

<section class="block" data-live>
	<h3>History{{ if .Runs }} <small>{{ len .Runs }} runs</small>{{ end }}</h3>
	{{- with .Activity }}{{ template "activity" . }}{{ end }}
	<h4>Runs <small>newest first</small></h4>
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Started</th><th>Runs as</th><th>Duration</th><th>Result</th><th></th></tr></thead>
		<tbody>
		{{- range .Runs }}
			<tr>
				<td class="nowrap key" title="{{ ago .StartedAt }}">{{ stamp .StartedAt }}</td>
				<td data-label="By">{{ template "who" (list $ .) }}</td>
				<td class="nowrap" data-label="Took">{{ duration .Elapsed }}</td>
				<td data-label="Result">{{ template "state" . }}</td>
				<td class="actions">
					<a class="btn btn-sm" href="/runs/{{ .ID }}" data-attach="{{ .ID }}">{{ if not (ended .State) }}Attach{{ else }}View{{ end }}</a>
					{{- template "replay" . }}
					{{- if $.SummarizesRun $r . }}<button type="button" class="btn btn-sm" data-summarize-run="{{ .ID }}" data-question="{{ summaryQuestion }}" title="Claude summarizes the run's output, in this window">Summarize</button>
					{{- else }}{{ with $.RunReport .ID }}{{ if .Pending }}<span class="faint">summarizing</span>{{ end }}{{ end }}{{ end }}
					{{- if and (not (ended .State)) $can }}<button type="button" class="btn-icon danger" data-kill="{{ .ID }}" title="Kill the run" aria-label="Kill the run"><svg class="ic"><use href="#i-stop"/></svg></button>{{ end }}
				</td>
			</tr>
			{{- with $.RunReport .ID }}{{ if .Body }}
			<tr class="summary-row">
				<td colspan="5"><details class="summary-details">
					<summary>Summary{{ with .By }} <span class="faint">asked by {{ . }}</span>{{ end }}</summary>
					<div class="markdown-doc">{{ markdown .Body }}</div>
					<a class="btn btn-sm" href="{{ .Href }}">Report</a>
				</details></td>
			</tr>
			{{- end }}{{ end }}
		{{- else }}
			<tr class="empty"><td colspan="5">No runs yet.{{ if $can }} Press Run to start the first one.{{ end }}</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>

{{- if not $r.Dir }}
<section class="block">
	<h3>Schedules <small>cron expressions, in the server's time zone</small></h3>
	<div class="table-wrap fit">
	<table class="table fit-cards">
		<thead><tr><th>Expression</th><th>Next run</th><th class="hide-narrow">Last run</th><th>Runs as</th><th>State</th><th></th></tr></thead>
		<tbody>
		{{- range .Schedules }}
			<tr>
				<td class="key" data-label="Expression"><code>{{ .Expression }}</code></td>
				<td data-label="Next run">{{ template "nextrun" . }}</td>
				<td class="nowrap hide-narrow" data-label="Last run">{{ if .LastRunAt }}{{ ago .LastRunAt }}{{ else }}<span class="faint">never</span>{{ end }}</td>
				<td data-label="Runs as">{{ template "runas-row" (list $ .) }}</td>
				<td data-label="State">{{ if .IsEnabled }}<span class="badge ok">enabled</span>{{ else }}<span class="badge">disabled</span>{{ end }}</td>
				<td class="actions">
					{{- if $can }}
					<form data-api="/api/schedules/{{ .ID }}" data-method="PUT" data-next="reload"><input type="hidden" name="enabled" data-type="bool" value="{{ if .IsEnabled }}false{{ else }}true{{ end }}"><button class="btn btn-sm" type="submit">{{ if .IsEnabled }}Disable{{ else }}Enable{{ end }}</button></form>
					<button type="button" class="btn-icon danger" data-delete="/api/schedules/{{ .ID }}" data-confirm="Delete the schedule {{ .Expression }} for {{ $r.Name }}?" data-next="reload" title="Delete" aria-label="Delete schedule"><svg class="ic"><use href="#i-trash"/></svg></button>
					{{- end }}
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="6">Not scheduled.{{ if $can }} Add a cron expression below to run it on a schedule{{ if $.Users }}, as the user you pick{{ else }}, as you{{ end }}.{{ end }}</td></tr>
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
		{{- template "runas" (list $ "sch-user") }}
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Schedule</button>
		</div>
	</form>
	{{- end }}
</section>
{{- end }}
{{ end }}
