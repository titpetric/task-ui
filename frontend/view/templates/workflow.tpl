{{ define "content" }}
{{- with $w := .Workflow }}
<div class="page-head">
	<div>
		<h2>{{ $w.Name }}</h2>
		<p class="faint">{{ len $w.Steps }} step{{ if ne (len $w.Steps) 1 }}s{{ end }}, by {{ $w.UserName }}</p>
	</div>
	<div class="actions" data-actions>
		{{- if $w.CanRun }}{{ if $w.Running }}
		<button type="button" class="btn" data-workflow-stop="{{ $w.Running }}"><svg class="ic"><use href="#i-stop"/></svg>Stop</button>
		{{- else }}
		<button type="button" class="btn btn-primary" data-workflow-run="{{ $w.ID }}"><svg class="ic"><use href="#i-play"/></svg>Run</button>
		{{- end }}{{ end }}
		{{- if $w.CanChange }}
		<a class="btn" href="/workflows/{{ $w.ID }}/edit"><svg class="ic"><use href="#i-edit"/></svg>Edit</a>
		{{- end }}
	</div>
</div>

{{- /* The page follows the workflow's topic: data-sse names it, and the
     parts below are updated in place from the events, no fetch. */}}
<section class="block" data-sse="workflow:{{ $w.ID }}" data-user="{{ $.User.Name }}">
	<h3>Pipeline</h3>
	{{- if $.WorkflowRuns }}{{ $latest := index $.WorkflowRuns 0 }}
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
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Started</th><th>By</th><th>Result</th><th>Steps</th></tr></thead>
		<tbody data-runs>
		{{- range $.WorkflowRuns }}
			<tr data-run="{{ .ID }}">
				<td class="nowrap" data-label="Started" title="{{ stamp .StartedAt }}">{{ ago .StartedAt }}</td>
				<td data-label="By">{{ .UserName }}</td>
				<td data-label="Result" data-result><span class="state"><i class="dot {{ .State }}"></i>{{ .State }}</span>{{ if .Message }} <span class="faint">{{ .Message }}</span>{{ end }}</td>
				<td data-label="Steps"><span class="chain">{{ range $i, $l := chain .Steps }}{{ if $i }}<span class="faint"> → </span>{{ end }}<a class="nowrap" href="{{ $l.Href }}" data-pos="{{ $i }}"{{ if $l.RunID }} data-attach="{{ $l.RunID }}"{{ end }}><i class="dot {{ $l.State }}"></i>{{ if $l.Workflow }}<svg class="ic"><use href="#i-flow"/></svg>{{ end }}{{ $l.Text }}</a>{{ end }}</span></td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="4">Not run yet.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{- end }}
{{ end }}

{{/* pipenode is a step of the pipeline: its name, state and, for a run's
     step, its terminal's last screen, polled while it runs (data-frame).
     A step that ran a workflow holds that run's chain. */}}
{{- define "pipenode" }}
{{- $l := index . 0 }}{{ $i := index . 1 }}
<li class="node{{ with $l.State }} {{ . }}{{ end }}" data-step="{{ $i }}">
	<span class="rail"><i class="dot{{ with $l.State }} {{ . }}{{ end }}"></i></span>
	<div class="what">
		<a href="{{ $l.Href }}"{{ if $l.RunID }} data-attach="{{ $l.RunID }}"{{ end }}>{{ if $l.Workflow }}<svg class="ic"><use href="#i-flow"/></svg>{{ end }}{{ $l.Text }}</a>
		<small>{{ if $l.State }}{{ $l.State }}{{ else }}step {{ add $i 1 }}{{ end }}</small>
		{{- if $l.RunID }}<div class="frame-tools"><button type="button" class="btn btn-sm" data-expand="{{ $l.RunID }}" aria-expanded="false">Expand</button></div>{{ end }}
		{{- if $l.RunID }}<a class="frame" href="/runs/{{ $l.RunID }}" data-attach="{{ $l.RunID }}"{{ if eq $l.State "running" }} data-frame="{{ $l.RunID }}"{{ end }} role="img" aria-label="Terminal of {{ $l.Text }}">{{ $l.FrameSVG }}</a>{{ end }}
		{{- if $l.WorkflowRunID }}
		<ol class="pipeline sub" data-sub="{{ $l.WorkflowRunID }}" data-workflow="{{ $l.WorkflowID }}">
			{{- range $j, $s := $l.Sub }}{{ template "pipenode" (list $s $j) }}{{ end }}
		</ol>
		{{- end }}
	</div>
</li>
{{- end }}
