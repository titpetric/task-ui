{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Projects</h2>
		<p>The folders under the root. task-ui pulls a git checkout before every run.</p>
	</div>
	{{- if .Can "project/edit" }}
	<div class="actions">
		<a class="btn btn-primary" href="/settings/projects/new"><svg class="ic"><use href="#i-plus"/></svg>Add project</a>
	</div>
	{{- end }}
</div>

<section class="block">
	{{- template "blockhead" (head "Projects" (print "under " .Root) (and (not .Entries) "No projects yet: add one from a git repository, or put a folder with jobs under the root.")) }}
	{{- with .Entries }}{{ template "projecttable" (list $ .) }}{{ end }}
</section>
{{ end }}
