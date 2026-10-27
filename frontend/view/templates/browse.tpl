{{ define "content" }}
{{- $p := .Folder }}
<div class="page-head">
	<div>
		<h2><svg class="ic"><use href="#i-folder-open"/></svg> {{ .Title }}</h2>
		<p>A directory of <a href="{{ $.Href $p.ID }}">{{ $p.ID }}</a>.{{ if $.Place }} Its jobs run in {{ $p.ID }}/{{ .Browse }}, under the rules of {{ $p.ID }}.{{ else if $.JobGroups }} It has no pipeline file, Makefile, package scripts or scripts folder of its own; the jobs below run here, under the rules of {{ $p.ID }}.{{ else }} Nothing to run here: no pipeline file, Makefile, package scripts or scripts folder.{{ end }}</p>
	</div>
	<div class="actions">
		{{- if $.Pinning }}
		{{ template "menu" (menu (print "Actions on " $.BrowseID) ($.PinItem $.BrowseID "" "" "folder")) }}
		{{- end }}
	</div>
</div>
{{- template "actions" $ }}
{{- with $.Place }}
{{- if .Error }}
<div class="alert bad"><b>Could not read a source.</b> {{ .Error }}</div>
{{- end }}
{{- end }}
{{- template "jobtable" $ }}
{{- template "promptbox" $ }}

<section class="block">
	<h3>Files <small>{{ $p.ID }}/{{ .Browse }}</small></h3>
	{{ template "tree" $ }}
</section>
{{ end }}
