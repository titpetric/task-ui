{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Workflows</h2>
		<p>Chains of steps: commands from any folders, prompts to Claude and summaries. Each step starts when the one before ends ok; a failure stops the rest, but for a summary. Running one needs permission to run every step.</p>
	</div>
	<div class="actions">
		<a class="btn btn-primary" href="/workflows/new"><svg class="ic"><use href="#i-plus"/></svg>Add workflow</a>
	</div>
</div>

{{- template "promptbox" . }}

<section class="block">
	{{- template "blockhead" (head "Workflows" "" (and (not .Workflows) "No workflows yet. Add one with a name and its steps, in order.")) }}
	{{- with .Workflows }}
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Workflow</th><th>Last run</th><th>By</th><th></th></tr></thead>
		<tbody>
		{{- range . }}
			<tr>
				<td class="key"><a href="/workflows/{{ .ID }}">{{ .Name }}</a>{{ with .Description }}<small class="faint wf-desc">{{ . }}</small>{{ end }}
					{{- range .Schedules }}<small class="faint wf-desc" title="{{ if .IsEnabled }}Scheduled{{ else }}A schedule, disabled{{ end }}, as {{ $.ScheduleUser . }}"><svg class="ic lead"><use href="#i-clock"/></svg><code>{{ .Expression }}</code>{{ if not .IsEnabled }} disabled{{ end }}</small>{{ end }}</td>
				<td class="nowrap" data-label="Last run">{{ if .Running }}<span class="state"><i class="dot running"></i>running</span>{{ else if .LastRun }}<span class="state"><i class="dot {{ .LastRun.State }}"></i>{{ .LastRun.State }}</span> <span class="faint" title="{{ stamp .LastRun.StartedAt }}">{{ ago .LastRun.StartedAt }}</span>{{ else }}<span class="faint">never run</span>{{ end }}</td>
				<td data-label="By">{{ template "user" (list $ .UserName) }}{{ if and .RunAs (ne .RunAs .UserName) }} <span class="faint nowrap">· runs as {{ .RunAs }}</span>{{ end }}</td>
				<td class="actions">
					{{- if .CanRun }}{{ if .Running }}
					<button type="button" class="btn btn-sm" data-workflow-stop="{{ .Running }}"><svg class="ic"><use href="#i-stop"/></svg>Stop</button>
					{{- else }}
					<button type="button" class="btn btn-primary btn-sm" data-workflow-run="{{ .ID }}"{{ if .Prompted }} data-workflow-prompt="{{ .Name }}"{{ end }}><svg class="ic"><use href="#i-play"/></svg>Run</button>
					{{- end }}{{ else if .RunDenied }}
					<span class="faint run-denied">{{ .RunDenied }}</span>{{ end }}
					{{- template "menu" (menu (print "Actions on " .Name)
						(and .CanChange (item "edit" "Edit" "href" (print "/workflows/" .ID "/edit")))
						(and .Running (item "eye" "The run going" "href" (print "/workflows/" .ID "/runs/" .Running)))
					) }}
				</td>
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>

{{- /* The newest runs of the workflows the user sees, a line each: its
     state on the dot by the workflow's name, linked to the run, and
     Claude's answer open under the line, with the chevron that folds it. */}}
<section class="block">
	{{- template "blockhead" (head "Recent runs" "" (and (not .WorkflowRuns) "No workflow has run yet.")) }}
	{{- with .WorkflowRuns }}
	<div class="table-wrap">
	<table class="table cards wf-runs">
		<thead><tr><th>Run</th><th>By</th><th>Started</th></tr></thead>
		<tbody>
		{{- range . }}
			<tr data-run-id="{{ .ID }}">
				<td class="key">{{ template "wffold" (list . true) }}<a href="/workflows/{{ .WorkflowID }}/runs/{{ .ID }}" title="{{ .State }}{{ with .Message }}, {{ . }}{{ end }}"><i class="dot {{ .State }}"></i>{{ .Name }}</a></td>
				<td data-label="By">{{ template "user" (list $ .UserName) }}</td>
				<td class="nowrap" data-label="Started" title="{{ stamp .StartedAt }}">{{ ago .StartedAt }}</td>
			</tr>
			{{- template "wfanswer" (list $ . 3 true) }}
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>
{{ end }}
