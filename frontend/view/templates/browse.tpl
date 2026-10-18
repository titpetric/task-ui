{{ define "content" }}
{{- $p := .Folder }}
<div class="page-head">
	<div>
		<h2><svg class="ic"><use href="#i-folder-open"/></svg> {{ .Title }}</h2>
		<p>A directory of <a href="/f/{{ $p.ID }}">{{ $p.ID }}</a>. Nothing to run here: no pipeline file, Makefile, package scripts or scripts folder.</p>
	</div>
</div>

<section class="block">
	<h3>Files <small>{{ $p.ID }}/{{ .Browse }}</small></h3>
	{{ template "tree" $ }}
</section>
{{ end }}
