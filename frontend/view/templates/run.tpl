{{ define "content" }}
{{- $run := .Run }}{{ $can := .CanRun .Folder .Job }}
<div class="page-head">
	<div>
		<h2>{{ $run.Folder }} / {{ $run.Job }} {{ template "state" $run }}</h2>
		<p>Started by {{ $run.UserName }} {{ ago $run.StartedAt }}{{ if $run.IsScheduled }} by the scheduler{{ end }}{{ if $run.IsForced }}, with Run force{{ end }}.{{ if $run.Live }} The session is live: attach to follow it{{ if $can }}, or end it{{ end }}.{{ else }} The session is over; what it showed is in the recording.{{ end }}</p>
	</div>
	<div class="actions">
		<button type="button" class="btn btn-primary" data-attach="{{ $run.ID }}">{{ if $run.Live }}Attach{{ else }}View{{ end }}</button>
		{{- if and $run.Live (eq $run.State "running") $can }}
		<button type="button" class="btn" data-stop="{{ $run.ID }}">Stop</button>
		<button type="button" class="btn btn-danger" data-kill="{{ $run.ID }}">Kill</button>
		{{- end }}
		{{- if $run.CastPath }}<a class="btn" href="/api/runs/{{ $run.ID }}/cast" download="{{ $run.ID }}.cast">Recording</a>{{ end }}
	</div>
</div>

{{- if $run.Live }}
<section class="block" data-live>
	<h3>Terminal <small>as it looks now, {{ if $run.Attached }}{{ $run.Attached }} watching{{ else }}nobody attached{{ end }}</small></h3>
	<a class="tile wide" href="#" data-attach="{{ $run.ID }}"><img class="frame" src="/api/runs/{{ $run.ID }}/frame.svg" alt="Terminal of {{ $run.Folder }} / {{ $run.Job }}" data-frame="{{ $run.ID }}"></a>
</section>
{{- end }}

<div class="grid-2">
<section class="block">
	<h3>Run</h3>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
			<tr><th>Folder</th><td><a href="/f/{{ $run.Folder }}"><svg class="ic"><use href="#i-folder"/></svg> {{ $run.Folder }}</a></td></tr>
			<tr><th>Job</th><td><a href="/f/{{ $run.Folder }}/{{ $run.Job }}">{{ $run.Job }}</a></td></tr>
			<tr><th>By</th><td>{{ template "who" $run }}</td></tr>
			<tr><th>Started</th><td>{{ stamp $run.StartedAt }}</td></tr>
			<tr><th>Finished</th><td>{{ if $run.FinishedAt }}{{ stamp $run.FinishedAt }}{{ else }}<span class="faint">in progress</span>{{ end }}</td></tr>
			<tr><th>Duration</th><td>{{ duration $run.Elapsed }}</td></tr>
			<tr><th>Result</th><td>{{ template "state" $run }}{{ if $run.Message }} <span class="dim">{{ $run.Message }}</span>{{ end }}</td></tr>
			<tr><th>Terminal</th><td>{{ $run.TermCols }}×{{ $run.TermRows }}</td></tr>
			<tr><th>ID</th><td class="faint">{{ $run.ID }}</td></tr>
		</tbody>
	</table>
	</div>
</section>

<section class="block">
	<h3>Artifacts <small>files left under $TASK_UI_ARTIFACTS</small></h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Name</th><th>Size</th><th>Type</th><th></th></tr></thead>
		<tbody>
		{{- range .Artifacts }}
			<tr>
				<td class="key"><a href="/api/artifacts/{{ .ID }}">{{ .Name }}</a></td>
				<td class="nowrap">{{ size .ByteSize }}</td>
				<td class="dim">{{ .ContentType }}</td>
				<td class="actions">
					<a class="btn-icon" href="/api/artifacts/{{ .ID }}" title="Download" aria-label="Download {{ .Name }}"><svg class="ic"><use href="#i-download"/></svg></a>
					{{- if $.User.CanAdmin }}<button type="button" class="btn-icon danger" data-delete="/api/artifacts/{{ .ID }}" data-confirm="Delete {{ .Name }}?" data-next="reload" title="Delete" aria-label="Delete {{ .Name }}"><svg class="ic"><use href="#i-trash"/></svg></button>{{ end }}
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="4">No artifacts. A job writes files to the directory in $TASK_UI_ARTIFACTS to leave some.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
</div>
{{ end }}
