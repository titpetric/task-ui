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
		<thead><tr><th>Workflow</th><th>Steps</th><th>Last run</th><th>By</th><th></th></tr></thead>
		<tbody>
		{{- range . }}
			<tr>
				<td class="key"><a href="/workflows/{{ .ID }}">{{ .Name }}</a>{{ with .Description }}<small class="faint wf-desc">{{ . }}</small>{{ end }}</td>
				<td data-label="Steps"><span class="chain">{{ template "chain" (list $ .Steps) }}</span></td>
				<td class="nowrap" data-label="Last run">{{ if .Running }}<span class="state"><i class="dot running"></i>running</span>{{ else if .LastRun }}<span class="state"><i class="dot {{ .LastRun.State }}"></i>{{ .LastRun.State }}</span> <span class="faint" title="{{ stamp .LastRun.StartedAt }}">{{ ago .LastRun.StartedAt }}</span>{{ else }}<span class="faint">never run</span>{{ end }}</td>
				<td data-label="By">{{ template "user" (list $ .UserName) }}</td>
				<td class="actions">
					{{- if .CanRun }}{{ if .Running }}
					<button type="button" class="btn btn-sm" data-workflow-stop="{{ .Running }}"><svg class="ic"><use href="#i-stop"/></svg>Stop</button>
					{{- else }}
					<button type="button" class="btn btn-primary btn-sm" data-workflow-run="{{ .ID }}"{{ if .Prompted }} data-workflow-prompt="{{ .Name }}"{{ end }}><svg class="ic"><use href="#i-play"/></svg>Run</button>
					{{- end }}{{ end }}
					{{- template "menu" (menu (print "Actions on " .Name)
						(item "flow" "Open" "href" (print "/workflows/" .ID))
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

{{- /* The newest runs of the workflows the user sees, each a thread on
     one line: what it started with and its reply, linked to the run. */}}
<section class="block">
	{{- template "blockhead" (head "Recent runs" "" (and (not .WorkflowRuns) "No workflow has run yet.")) }}
	{{- with .WorkflowRuns }}
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Run</th><th>By</th><th>Result</th><th>Input and answer</th></tr></thead>
		<tbody>
		{{- range . }}
			<tr>
				<td class="key"><a href="/workflows/{{ .WorkflowID }}/runs/{{ .ID }}">{{ .Name }}</a><small class="faint wf-desc" title="{{ stamp .StartedAt }}">{{ ago .StartedAt }}</small></td>
				<td data-label="By">{{ template "user" (list $ .UserName) }}</td>
				<td class="nowrap" data-label="Result"><span class="state"><i class="dot {{ .State }}"></i>{{ .State }}</span></td>
				<td class="wf-excerpt"{{ if or .Question .Agent }} data-label="Input and answer"{{ end }}>{{ template "wfexcerpt" . }}</td>
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>
{{ end }}
