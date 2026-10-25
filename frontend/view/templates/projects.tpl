{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Projects</h2>
		<p>The projects of each workspace, a block per workspace. task-ui pulls a git checkout before every run.</p>
	</div>
	{{- if .Can "project/edit" }}
	<div class="actions">
		<a class="btn btn-primary" href="/settings/projects/new"><svg class="ic"><use href="#i-plus"/></svg>Add project</a>
	</div>
	{{- end }}
</div>

{{- range .ProjectGroups }}
<section class="block">
	{{- template "blockhead" (head .Title (.Note $.Root) (and (not .Entries) .Empty)
		(and (not .Default) ($.Can "project/edit") (item "plus" "Add project" "href" (print "/settings/projects/new?workspace=" .Workspace)))) }}
	{{- with .Entries }}{{ template "projecttable" (list $ .) }}{{ end }}
</section>
{{- end }}
{{ end }}
