{{ define "content" }}
{{- with $w := .Workflow }}
<div class="page-head">
	<div>
		<h2>{{ $w.Name }}</h2>
		{{- with $w.Description }}
		<p>{{ . }}</p>
		{{- end }}
		<p class="faint">{{ len $w.Steps }} step{{ if ne (len $w.Steps) 1 }}s{{ end }}, by {{ $w.UserName }}</p>
	</div>
	<div class="actions" data-actions>
		{{- if $w.CanRun }}{{ if $w.Running }}
		<button type="button" class="btn" data-workflow-stop="{{ $w.Running }}"><svg class="ic"><use href="#i-stop"/></svg>Stop</button>
		{{- else }}
		<button type="button" class="btn btn-primary" data-workflow-run="{{ $w.ID }}"{{ if $w.Prompted }} data-workflow-prompt="{{ $w.Name }}"{{ end }}><svg class="ic"><use href="#i-play"/></svg>Run</button>
		{{- end }}{{ end }}
		{{- if $w.CanChange }}
		<a class="btn" href="/workflows/{{ $w.ID }}/edit"><svg class="ic"><use href="#i-edit"/></svg>Edit</a>
		{{- end }}
	</div>
</div>

{{- /* The page follows the workflow's topic: data-sse names it, and the
     events update the parts below in place, no fetch. */}}
<section class="block" data-sse="workflow:{{ $w.ID }}" data-user="{{ $.User.Name }}">
	<h3>Pipeline</h3>
	{{- /* The steps; while a run is going, its states and screens live. A
	     finished run opens with View under Runs. */}}
	{{- with $latest := $w.LastRun }}
	<p class="faint" data-latest>Latest run <span class="state"><i class="dot {{ $latest.State }}"></i>{{ $latest.State }}</span>{{ if $latest.Message }} {{ $latest.Message }},{{ end }} <span title="{{ stamp $latest.StartedAt }}">{{ ago $latest.StartedAt }}</span> by {{ $latest.UserName }}</p>
	{{- else }}
	<p class="faint" data-latest>Not run yet.</p>
	{{- end }}
	<ol class="pipeline" data-pipeline>
		{{- range $i, $l := $.Pipeline }}{{ template "pipenode" (list $l $i) }}{{ end }}
	</ol>
</section>

<section class="block">
	<h3>Runs</h3>
	{{- with $.Activity }}{{ template "activity" . }}{{ end }}
	<form method="get" action="/workflows/{{ $w.ID }}" class="filters">
		{{- template "pagesize" $ }}
		<button class="btn" type="submit">Show</button>
	</form>
	<div class="table-wrap">
	<table class="table cards wf-runs">
		<thead><tr><th>Run</th><th>By</th><th>Steps</th><th></th></tr></thead>
		<tbody data-runs{{ if $.Summarize }} data-summarize{{ end }}>
		{{- range $.WorkflowRuns }}
			<tr data-run-id="{{ .ID }}">
				<td class="key" data-result>{{ template "wffold" (list .) }}<a href="/workflows/{{ $w.ID }}/runs/{{ .ID }}" title="{{ stamp .StartedAt }}"><i class="dot {{ .State }}" title="{{ .State }}"></i>{{ ago .StartedAt }}</a>{{ with .Message }}<small class="faint wf-desc">{{ . }}</small>{{ end }}</td>
				<td data-label="By">{{ template "user" (list $ .UserName) }}</td>
				<td data-label="Steps"><span class="chain">{{ range $i, $l := $.Links .Steps }}{{ if $i }}<span class="faint then" aria-hidden="true"></span>{{ end }}<a class="nowrap" href="{{ $l.Href }}" data-pos="{{ $i }}"><i class="dot {{ $l.State }}"></i>{{ if $l.Workflow }}<svg class="ic lead"><use href="#i-flow"/></svg>{{ else if $l.Summary }}<svg class="ic lead"><use href="#i-summary"/></svg>{{ else if $l.Prompt }}<svg class="ic lead"><use href="#i-prompt"/></svg>{{ end }}{{ $l.Text }}</a>{{ end }}</span></td>
				<td class="actions"><a class="btn btn-sm" href="/workflows/{{ $w.ID }}/runs/{{ .ID }}" data-workflow-view>View</a>{{ template "menu" (menu (print "Actions on the run of " $w.Name)
					(detail "Run" .ID "mono" "")
					(detail "Started" (stamp .StartedAt))
					(item "file" "Open the run's page" "href" (print "/workflows/" $w.ID "/runs/" .ID))
					(and $.Summarize (ne .State "running") (item "note" "Summary" "data-summary" (print "/api/workflow-runs/" .ID "/summary") "data-summary-page" (print "/workflows/" $w.ID "/runs/" .ID) "data-summary-title" $w.Name "title" "Claude's summary of every step's output"))
				) }}</td>
			</tr>
			{{- template "wfanswer" (list $ . 4) }}
		{{- else }}
			<tr class="empty"><td colspan="4">Not run yet.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- template "pager" $.Pager }}
</section>
{{- end }}
{{ end }}
