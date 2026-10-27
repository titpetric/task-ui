{{ define "content" }}
{{- $p := .Folder }}{{ $default := eq $p.ID "internal" }}
<div class="page-head">
	<div>
		<h2><svg class="ic"><use href="#i-folder-open"/></svg> {{ $p.Title }}</h2>
		{{- if $p.Description }}<p>{{ $p.Description }}</p>{{ end }}
	</div>
	<div class="actions">
		{{- with $.FolderMenu }}
		{{- if and .Workspace (not .Projects) ($.Can "project/edit") }}
		<a class="btn btn-primary" href="{{ .AddProjectPage }}"><svg class="ic"><use href="#i-plus"/></svg>Add project</a>
		{{- end }}
		{{ template "menu" (menu (print "Actions on " $p.Title)
			(and .Workspace ($.Can "project/edit") (item "plus" "Add project" "href" .AddProjectPage))
			(and .Workspace ($.Can "workspace/view") (item "settings" "Edit workspace" "href" .Settings))
			(and .Workspace ($.Can "workspace/view") (item "key" "SSH keys" "href" .KeysPage))
			(and .Workspace ($.Can "secret/view") (item "secret" "Secrets" "href" .SecretsPage))
			(and .Project ($.Can "project/view") (item "settings" "Edit project" "href" .Settings))
			(and (not $default) $.Pinning ($.PinItem $p.ID "" "" .Kind))
			(and (not $default) .Project ($.Can "env/edit") (item "plus" "Add env file" "href" (print ($.Href $p.ID) "/env")))
			(and (not $default) (or .Workspace .Project) ($.Can "schedule/edit") (item "clock" "Add schedule" "href" (print "/schedules?folder=" (urlquery $p.ID))))
			(and .Pull ($.CanRun $p .Pull) (item "download" "Pull now" "data-run" "" "data-folder" $p.ID "data-dir" .Pull.Dir "data-job" .Pull.ID "data-title" (print $p.ID " / " .Pull.Name)))
			(and .Project $.Checkout ($.Can "project/delete") (danger "trash" "Remove checkout" "data-delete" (print "/api/projects/" .ID) "data-confirm" (removeConfirm .ID $.Checkout.Remote) "data-confirm-ok" "Remove checkout" "data-next" (print "/settings/projects?notice=" (urlquery "Removed " .ID))))
		) }}
		{{- end }}
	</div>
</div>
{{- template "pipelinealert" (list $ ($.PipelineAlert $p.ID)) }}
{{- template "actions" $ }}

{{- if and $p.Error (not $p.Errors) }}
<div class="alert bad"><b>Could not read a source.</b> {{ $p.Error }}</div>
{{- end }}

{{- if and (not $p.Jobs) (not $.SubFolders) (not $default) }}
<div class="alert">task-ui found nothing to run in this folder: no pipeline file, Makefile, package scripts or scripts folder.</div>
{{- end }}

{{- template "promptbox" $ }}

{{- if or (not $default) $.Tree }}
<section class="block">
	<h3>Files <small>{{ if $default }}the root{{ else }}{{ $p.ID }}/{{ end }}</small></h3>
	{{ template "tree" $ }}
</section>
{{- end }}

{{- with $.HomeFolder }}
<section class="block folder-home" id="home">
	<h3>Home <small>{{ . }}/, the $HOME of every run in the workspace</small></h3>
	{{ template "filetable" (list "" $.HomeTree) }}
</section>
{{- end }}

{{- if not $default }}
<div class="grid-2">
<section class="block">
	<h3>Folder</h3>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
			<tr><th>Path</th><td><svg class="ic"><use href="#i-folder"/></svg> {{ $p.ID }}/</td></tr>
			{{- range $p.Sources }}<tr><th>{{ .Kind }}</th><td><svg class="ic"><use href="#i-file"/></svg> {{ .File }}{{ if .Command }} <span class="dim">runs with</span> <code>{{ join .Command " " }} &lt;target&gt;</code>{{ end }}{{ if ne .Kind "script" }}{{ template "showsource" (list $p .File) }}{{ end }}</td></tr>{{ end }}
			{{- with $.Checkout }}<tr><th>Source</th><td><div class="mono-break">{{ .Remote }}</div><div class="dim">{{ .Branch }} <span class="faint">{{ .Commit }}</span> {{ .Subject }}{{ if not .When.IsZero }} <span class="faint">{{ ago .When }}</span>{{ end }}</div><div class="faint">pulled before every run</div></td></tr>{{ end }}
			{{- with $.Space }}<tr><th>Checkout size</th><td>{{ template "spacebar" . }}</td></tr>{{ end }}
			<tr><th>Who can run</th><td>{{ if $p.Allow }}admin, {{ join $p.Allow ", " }}{{ else }}admin, runner{{ end }}</td></tr>
			<tr><th>Hours</th><td>{{ template "hours" (list $p.Hours $) }}</td></tr>
			{{- if $p.Includes }}<tr><th>Includes</th><td>{{ range $p.Includes }}<div><svg class="ic"><use href="#i-file"/></svg> {{ . }}{{ template "showsource" (list $p .) }}</div>{{ end }}</td></tr>{{ end }}
			<tr><th>Scheduled</th><td>{{ if .Schedules }}{{ len .Schedules }} <a href="/schedules?folder={{ $p.ID }}">schedule{{ if ne (len .Schedules) 1 }}s{{ end }}</a>{{ else }}<span class="faint">none</span>, <a href="/schedules?folder={{ $p.ID }}">add one</a>{{ end }}</td></tr>
		</tbody>
	</table>
	</div>
</section>

{{- /* Env files are per project: a project draws the block while it has no env file, so Add env file is on the page. */}}
{{- $addEnv := and (not $default) ($.Can "env/edit") }}
{{- $ownsEnv := and $.FolderMenu $.FolderMenu.Project }}
{{- if or .EnvFiles (and $addEnv $ownsEnv) }}
<section class="block">
	{{- template "blockhead" (head "Environment" "env files written into the folder before each run, the secrets appended" (and (not .EnvFiles) "No env files.")
		(and $addEnv (item "plus" "Add env file" "href" (print ($.Href $p.ID) "/env")))) }}
	{{- with .EnvFiles }}
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>File</th>{{ if $.Can "env/view" }}<th></th>{{ end }}</tr></thead>
		<tbody>
		{{- range . }}
			<tr>
				<td class="key"><svg class="ic"><use href="#i-file"/></svg> {{ .Path }}</td>
				{{- if $.Can "env/view" }}
				<td class="actions">{{ template "menu" (menu (print "Actions on " .Path)
					(item "edit" "Edit" "href" (print ($.Href $p.ID) "/env/" .Path))) }}</td>
				{{- end }}
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>
{{- end }}
</div>
{{- end }}
{{ end }}
