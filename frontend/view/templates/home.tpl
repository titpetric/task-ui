{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Dashboard</h2>
		<p>Every terminal in progress, as it looks right now, for everyone. Open one to attach, or open a folder on the left to start a run.</p>
	</div>
</div>

{{- $jobs := 0 }}{{ range .Folders }}{{ $jobs = len .Jobs | add $jobs }}{{ end }}
<div class="metrics">
	<div><b>Folders</b><span>{{ len .Folders }}</span></div>
	<div><b>Jobs</b><span>{{ $jobs }}</span></div>
	<div><b>Running</b><span>{{ len .Running }}</span></div>
	<div><b>Scheduled</b><span>{{ len .Schedules }}</span></div>
	<div><b>Your role</b><span>{{ .User.Role }}</span></div>
</div>

<section class="block" data-live>
	<h3>Running now <small>one run per job, instance wide</small></h3>
	{{- if .Runs }}
	<div class="tiles">
	{{- range .Runs }}
		<a class="tile" href="/runs/{{ .ID }}" data-attach="{{ .ID }}">
			<div class="tile-head">
				<i class="dot {{ .State }}"></i>
				<b>{{ .Folder }} / {{ .Job }}</b>
				<span class="spacer"></span>
				<span class="faint">{{ if .Attached }}{{ .Attached }} watching{{ end }}</span>
				<span class="avatar sm" title="{{ .UserName }}">{{ .Initials }}</span>
			</div>
			<div class="frame" data-frame="{{ .ID }}" role="img" aria-label="Terminal of {{ .Folder }} / {{ .Job }}">{{ .FrameSVG }}</div>
		</a>
	{{- end }}
	</div>
	{{- else }}
	<div class="alert">Nothing is running. Open a folder and press Run, or run a workflow, and the terminal shows up here for everyone.</div>
	{{- end }}
</section>

<section class="block">
	<h3>Control centre <small>the jobs folders share get a column; the rest are under other</small></h3>
	{{- with $c := .Control }}
	<div class="table-wrap">
	<table class="table control">
		<thead><tr><th>Folder</th>{{ range $c.Columns }}<th class="col">{{ . }}</th>{{ end }}<th>Other</th></tr></thead>
		<tbody>
		{{- range $row := $c.Rows }}
			<tr>
				<td class="key nowrap"><a href="/f/{{ $row.Folder.ID }}">{{ $row.Folder.ID }}</a></td>
				{{- range $i, $job := $row.Cells }}
				<td class="col" data-label="{{ index $c.Columns $i }}">{{ with $job }}{{ template "controljob" (list $ $row.Folder . false) }}{{ end }}</td>
				{{- end }}
				<td class="other" data-label="Other">{{ range $row.Extra }}{{ template "controljob" (list $ $row.Folder . true) }}{{ end }}</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="{{ add (len $c.Columns) 2 }}">No jobs yet. Add a project from a git repository in Settings, or put a folder with something to run under the root.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>
{{ end }}

{{/* controljob is a job's button in the control centre: Run, an eye on
     the run in progress, or a #! to its page for a user who may not run
     it. In the other column the button carries the job's name. */}}
{{- define "controljob" }}
{{- $ := index . 0 }}{{ $f := index . 1 }}{{ $r := index . 2 }}{{ $named := index . 3 }}
{{- $run := $.RunningRun $r.Key }}
{{- if $run -}}
<a class="cc{{ if $named }} named{{ end }} running" href="/runs/{{ $run.ID }}" data-attach="{{ $run.ID }}" title="Attach to {{ $f.ID }} / {{ $r.Name }}, running"><svg class="ic"><use href="#i-eye"/></svg>{{ if $named }}<span>{{ $r.Name }}</span>{{ end }}</a>
{{- else if $.CanRun $f $r -}}
<button type="button" class="cc{{ if $named }} named{{ end }}" data-run data-folder="{{ $f.ID }}" data-job="{{ $r.ID }}" data-title="{{ $f.ID }} / {{ $r.Name }}" title="Run {{ $f.ID }} / {{ $r.Name }}"><svg class="ic"><use href="#i-play"/></svg>{{ if $named }}<span>{{ $r.Name }}</span>{{ end }}</button>
{{- else -}}
<a class="cc{{ if $named }} named{{ end }} idle" href="/f/{{ $f.ID }}/{{ $r.ID }}" title="{{ $f.ID }} / {{ $r.Name }}: your role may not run it"><i class="shebang" aria-hidden="true">#!</i>{{ if $named }}<span>{{ $r.Name }}</span>{{ end }}</a>
{{- end }}
{{- end }}
