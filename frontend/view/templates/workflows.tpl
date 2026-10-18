{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Workflows</h2>
		<p>Jobs from any folders, chained: Run starts the first, and each next one when the one before ended ok. The first that does not end ok ends the workflow, and the rest are skipped. Anyone can make one; running it needs the right to run every step.</p>
	</div>
</div>

<div class="table-wrap">
<table class="table cards">
	<thead><tr><th>Workflow</th><th>Steps</th><th>Last run</th><th>By</th><th></th></tr></thead>
	<tbody>
	{{- range .Workflows }}
		<tr>
			<td class="key"><a href="/workflows/{{ .ID }}">{{ .Name }}</a></td>
			<td data-label="Steps"><span class="chain">{{ range $i, $s := .Steps }}{{ if $i }}<span class="faint"> → </span>{{ end }}<span class="nowrap">{{ $s.Folder }} / {{ $s.Job }}</span>{{ end }}</span></td>
			<td class="nowrap" data-label="Last run">{{ if .Running }}<span class="state"><i class="dot running"></i>running</span>{{ else if .LastRun }}<span class="state"><i class="dot {{ .LastRun.State }}"></i>{{ .LastRun.State }}</span> <span class="faint" title="{{ stamp .LastRun.StartedAt }}">{{ ago .LastRun.StartedAt }}</span>{{ else }}<span class="faint">never run</span>{{ end }}</td>
			<td data-label="By">{{ .UserName }}</td>
			<td class="actions">
				{{- if .CanRun }}{{ if .Running }}
				<button type="button" class="btn btn-sm" data-workflow-stop="{{ .Running }}"><svg class="ic"><use href="#i-stop"/></svg>Stop</button>
				{{- else }}
				<button type="button" class="btn btn-primary btn-sm" data-workflow-run="{{ .ID }}"><svg class="ic"><use href="#i-play"/></svg>Run</button>
				{{- end }}{{ end }}
				<a class="btn btn-sm" href="/workflows/{{ .ID }}">Open</a>
			</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="5">No workflows yet. Make one below: a name and its jobs, in order.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>

<section class="block">
	<h3>New workflow</h3>
	{{ template "stepeditor" (list $ nil "/api/workflows" "POST" "/workflows" "Create") }}
</section>
{{ end }}
