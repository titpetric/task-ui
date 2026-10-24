{{ define "content" }}
{{- if and .Workflow .WorkflowRuns }}
{{- $w := .Workflow }}{{ $r := index .WorkflowRuns 0 }}
{{- /* data-pipeline-view is what the workflow page's window shows. */}}
<section class="block" data-pipeline-view>
	<div class="page-head">
		<div>
			<h2>{{ $w.Name }}</h2>
			<p class="faint"><span class="state"><i class="dot {{ $r.State }}"></i>{{ $r.State }}</span>{{ if $r.Message }} {{ $r.Message }},{{ end }} started <span title="{{ stamp $r.StartedAt }}">{{ ago $r.StartedAt }}</span> by {{ $r.UserName }}{{ if $r.FinishedAt }}, took {{ duration $r.Took }}{{ end }}</p>
		</div>
	</div>
	{{- template "wfthread" (list $ $r) }}
	<h3 class="wf-steps-title">Steps</h3>
	<ol class="pipeline">
		{{- range $i, $l := $.Pipeline }}{{ template "pipenode" (list $l $i) }}{{ end }}
	</ol>
	{{- template "promptbox" $ }}
</section>
{{- end }}
{{ end }}
