{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Workflows</h2>
		<p>A workflow runs jobs from any folders, one after another. Run starts the first step; each next step starts when the one before it ends ok. If a step fails, the workflow stops and skips the rest. A step can also run another workflow. Anyone can create a workflow, but to run it you need permission to run every step.</p>
	</div>
	<div class="actions">
		<a class="btn btn-primary" href="/workflows/new">Add workflow</a>
	</div>
</div>

<div class="table-wrap">
<table class="table cards">
	<thead><tr><th>Workflow</th><th>Steps</th><th>Last run</th><th>By</th><th></th></tr></thead>
	<tbody>
	{{- range .Workflows }}
		<tr>
			<td class="key"><a href="/workflows/{{ .ID }}">{{ .Name }}</a></td>
			<td data-label="Steps"><span class="chain">{{ template "chain" (list $ .Steps) }}</span></td>
			<td class="nowrap" data-label="Last run">{{ if .Running }}<span class="state"><i class="dot running"></i>running</span>{{ else if .LastRun }}<span class="state"><i class="dot {{ .LastRun.State }}"></i>{{ .LastRun.State }}</span> <span class="faint" title="{{ stamp .LastRun.StartedAt }}">{{ ago .LastRun.StartedAt }}</span>{{ else }}<span class="faint">never run</span>{{ end }}</td>
			<td data-label="By">{{ template "user" (list $ .UserName) }}</td>
			<td class="actions">
				{{- if .CanRun }}{{ if .Running }}
				<button type="button" class="btn btn-sm" data-workflow-stop="{{ .Running }}"><svg class="ic"><use href="#i-stop"/></svg>Stop</button>
				{{- else }}
				<button type="button" class="btn btn-primary btn-sm" data-workflow-run="{{ .ID }}"{{ if .Prompted }} data-workflow-prompt="{{ .Name }}"{{ end }}><svg class="ic"><use href="#i-play"/></svg>Run</button>
				{{- end }}{{ end }}
				<a class="btn btn-sm" href="/workflows/{{ .ID }}">Open</a>
			</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="5">No workflows yet. Add one with a name and its jobs, in order.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
{{ end }}
