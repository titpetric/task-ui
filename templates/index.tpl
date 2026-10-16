{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Overview</h2>
		<p>Open a folder on the left to see its jobs, run one, or look through its history.</p>
	</div>
</div>

{{- $jobs := 0 }}{{ range .Folders }}{{ $jobs = len .Jobs | add $jobs }}{{ end }}
<div class="metrics">
	<div><b>Folders</b><span>{{ len .Folders }}</span></div>
	<div><b>Jobs</b><span>{{ $jobs }}</span></div>
	<div><b>Running</b><span>{{ len .Runs }}</span></div>
	<div><b>Your role</b><span>{{ .User.Role }}</span></div>
</div>

<section class="block" data-live>
	<h3>Running now <small>one run per job, instance wide</small></h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Folder</th><th>Job</th><th>Started by</th><th>Started</th><th></th></tr></thead>
		<tbody>
		{{- range .Runs }}
			<tr>
				<td><a href="/f/{{ .Folder }}">{{ .Folder }}</a></td>
				<td class="key"><span class="state"><i class="dot running"></i><a href="/f/{{ .Folder }}/{{ .Job }}">{{ .Job }}</a></span>{{ if .Forced }} <span class="badge signal">forced</span>{{ end }}</td>
				<td>{{ .User }}</td>
				<td class="nowrap" title="{{ stamp .Started }}">{{ ago .Started }}</td>
				<td class="actions">
					<form method="post" action="/api/runs/{{ .Folder }}/{{ .Job }}/stop" data-confirm="Kill {{ .Folder }} / {{ .Job }}, started by {{ .User }}?" data-confirm-ok="Kill">
						<input type="hidden" name="back" value="/">
						<button class="btn-icon danger" type="submit" title="Kill the run" aria-label="Kill the run"><svg class="ic"><use href="#i-stop"/></svg></button>
					</form>
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="5">Nothing is running. Runs started from any window show up here.</td></tr>
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

{{- if .HistoryEnabled }}
<section class="block">
	<h3>Recent runs</h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Started</th><th>Folder</th><th>Job</th><th>User</th><th>Duration</th><th>Result</th></tr></thead>
		<tbody>
		{{- range .Recent }}
			<tr>
				<td class="nowrap" title="{{ stamp .Started }}">{{ ago .Started }}</td>
				<td><a href="/f/{{ .Folder }}">{{ .Folder }}</a></td>
				<td class="key"><a href="/f/{{ .Folder }}/{{ .Job }}#history={{ .ID }}">{{ .Job }}</a></td>
				<td>{{ .User }}</td>
				<td class="nowrap">{{ if ne .State "running" }}{{ duration .Duration }}{{ end }}</td>
				<td>{{ template "state" . }}</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="6">No runs recorded yet. Run a job and it is recorded here.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{- end }}
{{ end }}
