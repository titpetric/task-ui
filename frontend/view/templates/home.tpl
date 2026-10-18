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
	<h3>Recent runs</h3>
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Started</th><th>Folder</th><th>Job</th><th>By</th><th>Duration</th><th>Result</th><th></th></tr></thead>
		<tbody>
		{{- range .Recents }}
			<tr>
				<td class="nowrap" data-label="Started" title="{{ stamp .StartedAt }}">{{ ago .StartedAt }}</td>
				<td data-label="Folder"><a href="/f/{{ .Folder }}">{{ .Folder }}</a></td>
				<td class="key"><a href="/f/{{ .Folder }}/{{ .Job }}">{{ .Job }}</a></td>
				<td data-label="By">{{ template "who" . }}</td>
				<td class="nowrap" data-label="Took">{{ duration .Elapsed }}</td>
				<td data-label="Result">{{ template "state" . }}</td>
				<td class="actions"><a class="btn btn-sm" href="/runs/{{ .ID }}" data-attach="{{ .ID }}">{{ if eq .State "running" }}Attach{{ else }}View{{ end }}</a></td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="7">No runs yet. Run a job and it is recorded here.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>

<section class="block">
	<h3>Folders</h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Folder</th><th>Name</th><th class="hide-sm">Description</th><th>Sources</th><th class="text-end">Jobs</th></tr></thead>
		<tbody>
		{{- range .Folders }}
			<tr>
				<td class="key nowrap"><a href="/f/{{ .ID }}"><svg class="ic"><use href="#i-folder"/></svg> {{ .ID }}/</a></td>
				<td>{{ .Title }}</td>
				<td class="desc hide-sm">{{ if .Error }}<span class="badge bad">error</span> {{ .Error }}{{ else }}{{ .Description }}{{ end }}</td>
				<td class="nowrap">{{ range .Sources }}<span class="badge">{{ .Kind }}</span> {{ end }}</td>
				<td class="text-end">{{ len .Jobs }}</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="5">No folders yet. Add one from a git repository in Settings, or put a folder with something to run under the root.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{ end }}
