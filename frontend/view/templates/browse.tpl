{{ define "content" }}
{{- $p := .Folder }}
<div class="page-head">
	<div>
		<h2><svg class="ic"><use href="#i-folder-open"/></svg> {{ .Title }}</h2>
		<p>A directory of <a href="{{ $.Href $p.ID }}">{{ $p.ID }}</a>.{{ if $.Place }} Its jobs run in {{ $p.ID }}/{{ .Browse }}, under the rules of {{ $p.ID }}.{{ else if and $.Actions $.Actions.Skills }} It has no pipeline file, Makefile, package scripts or scripts folder of its own; the skills below run here, under the rules of {{ $p.ID }}.{{ else }} Nothing to run here: no pipeline file, Makefile, package scripts or scripts folder.{{ end }}</p>
	</div>
	{{- with $.Place }}
	<div class="actions">
		<span class="dim nowrap">{{ range .Sources }}<span class="badge">{{ .Kind }}</span> {{ end }}</span>
	</div>
	{{- end }}
</div>
{{- template "actions" $ }}
{{- template "promptbox" $ }}

{{- with $.Place }}
{{- if .Error }}
<div class="alert bad"><b>Could not read a source.</b> {{ .Error }}</div>
{{- end }}
{{- template "sources" (list $ .) }}
{{- end }}

<section class="block">
	<h3>Files <small>{{ $p.ID }}/{{ .Browse }}</small></h3>
	{{ template "tree" $ }}
</section>
{{ end }}
