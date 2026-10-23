{{ define "content" }}
{{- $run := .Run }}{{ $can := .CanRun .Folder .Job }}
<div class="page-head">
	<div>
		<h2>{{ $run.Name }} <span class="badge {{ tone $run.State }}">{{ $run.State }}{{ if and (ne $run.State "ok") (ne $run.State "running") }} {{ $run.ExitCode }}{{ end }}</span></h2>
		<p>Started by {{ $run.UserName }} {{ ago $run.StartedAt }}{{ if $run.IsScheduled }} by the scheduler{{ end }}{{ if $run.IsForced }}, with Run force{{ end }}.{{ if eq $run.State "running" }} It's still running: attach to follow it{{ if $can }}, or stop it{{ end }}.{{ else }} The run has finished; its full output is below.{{ end }}</p>
	</div>
	<div class="actions">
		<button type="button" class="btn btn-primary" data-attach="{{ $run.ID }}">{{ if eq $run.State "running" }}Attach{{ else }}View{{ end }}</button>
		{{- if and $run.Live (eq $run.State "running") $can }}
		<button type="button" class="btn" data-stop="{{ $run.ID }}">Stop</button>
		<button type="button" class="btn btn-danger" data-kill="{{ $run.ID }}">Kill</button>
		{{- end }}
		{{- if $run.CastPath }}<a class="btn" href="/api/runs/{{ $run.ID }}/cast" download="{{ $run.ID }}.cast">Recording</a>{{ end }}
		{{- if and $run.CastPath (ne $run.State "running") }}<button type="button" class="btn" data-replay="{{ $run.ID }}" data-replay-title="{{ $run.Name }}" data-replay-who="{{ $run.UserName }}" title="Replay the recording" aria-label="Replay the recording">|&gt;</button>{{ end }}
	</div>
</div>

{{- if and $run.Live (eq $run.State "running") }}
<section class="block" data-live>
	<h3>Terminal <small>as it looks now, {{ if $run.Attached }}{{ $run.Attached }} watching{{ else }}nobody attached{{ end }}</small></h3>
	<a class="tile wide" href="#" data-attach="{{ $run.ID }}"><div class="frame" data-frame="{{ $run.ID }}" role="img" aria-label="Terminal of {{ $run.Name }}">{{ $run.FrameSVG }}</div></a>
</section>
{{- end }}

{{- if and $run.CastPath (ne $run.State "running") }}
<section class="block">
	<h3>Output <small>every line the run printed</small></h3>
	<pre class="output" data-output="{{ $run.ID }}" data-cols="{{ $run.TermCols }}" data-rows="{{ $run.TermRows }}"><span class="faint">Reading the recording…</span></pre>
</section>
{{- end }}

{{- template "promptbox" . }}

<div class="grid-2">
<section class="block">
	<h3>Run</h3>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
			<tr><th>Folder</th><td>{{ with $.RunFolderHref $run.Folder "" }}<a href="{{ . }}"><svg class="ic"><use href="#i-folder"/></svg> {{ $run.Folder }}</a>{{ else }}{{ $run.Folder }}{{ end }}</td></tr>
			{{- with $run.Dir }}<tr><th>Directory</th><td><a href="{{ $.RunFolderHref $run.Folder . }}"><svg class="ic"><use href="#i-folder-open"/></svg> {{ . }}/</a> <span class="dim">the job ran here</span></td></tr>{{ end }}
			<tr><th>Job</th><td>{{ with $.RunJobHref $run.Folder $run.Dir $run.Job }}<a href="{{ . }}">{{ $run.Job }}</a>{{ else }}{{ $run.Job }}{{ end }}</td></tr>
			{{- with $run.ArgList }}<tr><th>Arguments</th><td>{{ range . }}<code>{{ . }}</code> {{ end }}<span class="dim">after --, one argument each</span></td></tr>{{ end }}
			<tr><th>By</th><td>{{ template "who" (list $ $run) }}</td></tr>
			<tr><th>Started</th><td>{{ stamp $run.StartedAt }}</td></tr>
			<tr><th>Finished</th><td>{{ if $run.FinishedAt }}{{ stamp $run.FinishedAt }}{{ else }}<span class="faint">in progress</span>{{ end }}</td></tr>
			<tr><th>Duration</th><td>{{ duration $run.Elapsed }}</td></tr>
			<tr><th>Result</th><td>{{ template "state" $run }}{{ if $run.Message }} <span class="dim">{{ $run.Message }}</span>{{ end }}</td></tr>
			{{- if $run.Reason }}<tr><th>Reason</th><td>{{ $run.Reason }} <span class="badge bad">out of hours</span></td></tr>{{ end }}
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
					{{- if $.Can "artifact/delete" }}<button type="button" class="btn-icon danger" data-delete="/api/artifacts/{{ .ID }}" data-confirm="Delete {{ .Name }}?" data-next="reload" title="Delete" aria-label="Delete {{ .Name }}"><svg class="ic"><use href="#i-trash"/></svg></button>{{ end }}
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="4">No artifacts. To keep files from a run, the job writes them to the directory in $TASK_UI_ARTIFACTS.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
</div>
{{ end }}
