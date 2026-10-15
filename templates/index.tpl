{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Overview</h2>
		<p>Pick a runbook on the left to read what it does, run it, or look through its history.</p>
	</div>
</div>

{{- $runbooks := 0 }}{{ range .Projects }}{{ $runbooks = len .Runbooks | add $runbooks }}{{ end }}
<div class="metrics">
	<div><b>Projects</b><span>{{ len .Projects }}</span></div>
	<div><b>Runbooks</b><span>{{ $runbooks }}</span></div>
	<div><b>Running</b><span>{{ len .Runs }}</span></div>
	<div><b>Your role</b><span>{{ .User.Role }}</span></div>
</div>

<section class="block" data-live>
	<h3>Running now <small>one run per runbook, instance wide</small></h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Project</th><th>Runbook</th><th>Started by</th><th>Started</th><th></th></tr></thead>
		<tbody>
		{{- range .Runs }}
			<tr>
				<td>{{ .Project }}</td>
				<td class="key"><span class="state"><i class="dot running"></i><a href="/p/{{ .Project }}/{{ .Runbook }}">{{ .Runbook }}</a></span>{{ if .Forced }} <span class="badge signal">forced</span>{{ end }}</td>
				<td>{{ .User }}</td>
				<td class="nowrap" title="{{ stamp .Started }}">{{ ago .Started }}</td>
				<td class="actions">
					<form method="post" action="/api/runs/{{ .Project }}/{{ .Runbook }}/stop" data-confirm="Kill {{ .Project }} / {{ .Runbook }}, started by {{ .User }}?">
						<input type="hidden" name="back" value="/">
						<button class="btn btn-danger btn-sm" type="submit">Kill</button>
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
	<h3>Projects</h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Project</th><th class="hide-sm">Description</th><th>File</th><th class="text-end">Runbooks</th><th></th></tr></thead>
		<tbody>
		{{- range .Projects }}
			<tr>
				<td class="key"><a href="/p/{{ .ID }}">{{ .Title }}</a></td>
				<td class="desc hide-sm">{{ if .Error }}<span class="badge bad">error</span> {{ .Error }}{{ else }}{{ .Description }}{{ end }}</td>
				<td class="dim nowrap">{{ .File }} <span class="badge">{{ .Kind }}</span></td>
				<td class="text-end">{{ len .Runbooks }}</td>
				<td class="actions"><a class="btn btn-sm" href="/p/{{ .ID }}">Open</a></td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="5">No projects yet. Put a folder with an atkins.yml or Taskfile.yml under the root folder, or list one in task-ui.yml.</td></tr>
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
		<thead><tr><th>Started</th><th>Project</th><th>Runbook</th><th>User</th><th>Duration</th><th>Result</th></tr></thead>
		<tbody>
		{{- range .Recent }}
			<tr>
				<td class="nowrap" title="{{ stamp .Started }}">{{ ago .Started }}</td>
				<td>{{ .Project }}</td>
				<td class="key"><a href="/p/{{ .Project }}/{{ .Runbook }}">{{ .Runbook }}</a></td>
				<td>{{ .User }}</td>
				<td class="nowrap">{{ if ne .State "running" }}{{ duration .Duration }}{{ end }}</td>
				<td>{{ template "state" . }}</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="6">No runs recorded yet. Run a runbook and it is recorded here.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{- end }}
{{ end }}
