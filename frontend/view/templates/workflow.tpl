{{ define "content" }}
{{- with $w := .Workflow }}
<div class="page-head">
	<div>
		<h2>{{ $w.Name }}</h2>
		<p class="faint">{{ len $w.Steps }} step{{ if ne (len $w.Steps) 1 }}s{{ end }}, by {{ $w.UserName }}</p>
	</div>
	<div class="actions" data-live="actions">
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

<section class="block" data-sse="workflow:{{ $w.ID }}">
	<h3>Pipeline</h3>
	<div data-live="pipeline">
		{{- if $.WorkflowRuns }}{{ $latest := index $.WorkflowRuns 0 }}
		<p class="faint">Latest run <span class="state"><i class="dot {{ $latest.State }}"></i>{{ $latest.State }}</span>{{ if $latest.Message }} {{ $latest.Message }},{{ end }} <span title="{{ stamp $latest.StartedAt }}">{{ ago $latest.StartedAt }}</span> by {{ $latest.UserName }}</p>
		{{- else }}
		<p class="faint">Not run yet.</p>
		{{- end }}
		<ol class="pipeline">
			{{- range $i, $l := $.Pipeline }}
			<li class="node{{ with $l.State }} {{ . }}{{ end }}">
				<span class="rail"><i class="dot{{ with $l.State }} {{ . }}{{ end }}"></i></span>
				<div class="what">
					<a href="{{ $l.Href }}"{{ if $l.RunID }} data-attach="{{ $l.RunID }}"{{ end }}>{{ if $l.Workflow }}<svg class="ic"><use href="#i-flow"/></svg>{{ end }}{{ $l.Text }}</a>
					<small>{{ if $l.State }}{{ $l.State }}{{ else }}step {{ add $i 1 }}{{ end }}</small>
				</div>
			</li>
			{{- end }}
		</ol>
	</div>
</section>

<section class="block">
	<h3>Runs</h3>
	<div class="table-wrap" data-live="runs">
	<table class="table cards">
		<thead><tr><th>Started</th><th>By</th><th>Result</th><th>Steps</th></tr></thead>
		<tbody>
		{{- range $.WorkflowRuns }}
			<tr>
				<td class="nowrap" data-label="Started" title="{{ stamp .StartedAt }}">{{ ago .StartedAt }}</td>
				<td data-label="By">{{ .UserName }}</td>
				<td data-label="Result"><span class="state"><i class="dot {{ .State }}"></i>{{ .State }}</span>{{ if .Message }} <span class="faint">{{ .Message }}</span>{{ end }}</td>
				<td data-label="Steps"><span class="chain">{{ template "chain" .Steps }}</span></td>
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
