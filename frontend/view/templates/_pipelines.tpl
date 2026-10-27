{{- /* pipelinealert is the alert of the broken atkins pipelines in a
     place, from Page.PipelineAlert: a row per finding with its folder,
     file and job, the rows past five folded under Show more. atkins lists
     no job of a pipeline its linter refuses. (list $ alert) */}}
{{- define "pipelinealert" }}
{{- $ := index . 0 }}{{ $a := index . 1 }}
{{- with $a }}
<div class="alert bad pipeline-alert" role="alert">
	<b>{{ .Pipelines }} {{ if eq .Pipelines 1 }}pipeline has{{ else }}pipelines have{{ end }} errors.</b> atkins lists no jobs of a pipeline its linter refuses; fix the file and they return.
	<ul class="pipeline-errors">
		{{- range .Shown }}{{ template "pipelineerror" (list $ .) }}{{ end }}
	</ul>
	{{- if .More }}
	<details class="pipeline-more">
		<summary><svg class="ic chev"><use href="#i-chev"/></svg>Show {{ len .More }} more</summary>
		<ul class="pipeline-errors">
			{{- range .More }}{{ template "pipelineerror" (list $ .) }}{{ end }}
		</ul>
	</details>
	{{- end }}
</div>
{{- end }}
{{- end }}

{{- /* pipelineerror is one finding of a pipeline alert. (list $ row) */}}
{{- define "pipelineerror" }}
{{- $ := index . 0 }}{{ $r := index . 1 }}
<li><a href="{{ $.Href $r.Folder.ID }}">{{ $r.Folder.ID }}</a> <code>{{ $r.Finding.Path }}</code>{{ with $r.Finding.Job }} <span class="dim">job</span> <b>{{ . }}</b>{{ end }}: <span{{ with $r.Finding.Issue }} title="{{ . }}"{{ end }}>{{ $r.Finding.Detail }}</span></li>
{{- end }}

{{- /* pipelinenote is the dashboard's line for the broken pipelines the
     user sees, each workspace linked to the page that lists them. (list
     $ alert) */}}
{{- define "pipelinenote" }}
{{- $ := index . 0 }}{{ $a := index . 1 }}
{{- with $a }}
<div class="alert bad" role="alert"><b>{{ .Pipelines }} {{ if eq .Pipelines 1 }}pipeline has{{ else }}pipelines have{{ end }} errors</b> in {{ range $i, $ws := .Workspaces }}{{ if $i }}, {{ end }}<a href="{{ $.Href $ws }}">{{ if eq $ws "internal" }}the default workspace{{ else }}{{ $ws }}{{ end }}</a>{{ end }}. {{ if eq (len .Workspaces) 1 }}Its page lists{{ else }}Their pages list{{ end }} them.</div>
{{- end }}
{{- end }}

{{- /* sidebad is the warning badge of a folder in the sidebar that
     failed to read: the count of atkins' findings, titled with the
     first, or ! with the error. */}}
{{- define "sidebad" }}
{{- with .Errors }}<small class="bad" title="{{ (index . 0).Detail }}">!{{ len . }}</small>{{ else }}<small class="bad" title="{{ .Error }}">!</small>{{ end }}
{{- end }}
