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

{{- define "pipelineerror" }}
{{- $ := index . 0 }}{{ $r := index . 1 }}
<li><a href="{{ $.Href $r.Folder.ID }}">{{ $r.Folder.ID }}</a> <code>{{ $r.Finding.Path }}</code>{{ with $r.Finding.Job }} <span class="dim">job</span> <b>{{ . }}</b>{{ end }}: <span{{ with $r.Finding.Issue }} title="{{ . }}"{{ end }}>{{ $r.Finding.Detail }}</span></li>
{{- end }}

{{- define "pipelinenote" }}
{{- $ := index . 0 }}{{ $a := index . 1 }}
{{- with $a }}
<div class="alert bad" role="alert"><b>{{ .Pipelines }} {{ if eq .Pipelines 1 }}pipeline has{{ else }}pipelines have{{ end }} errors</b> in {{ range $i, $ws := .Workspaces }}{{ if $i }}, {{ end }}<a href="{{ $.Href $ws }}">{{ if eq $ws "internal" }}internal{{ else }}{{ $ws }}{{ end }}</a>{{ end }}. {{ if eq (len .Workspaces) 1 }}Its page lists{{ else }}Their pages list{{ end }} them.</div>
{{- end }}
{{- end }}

{{- define "sidebad" }}
{{- with .Errors }}<small class="bad" title="{{ (index . 0).Detail }}">!{{ len . }}</small>{{ else }}<small class="bad" title="{{ .Error }}">!</small>{{ end }}
{{- end }}
