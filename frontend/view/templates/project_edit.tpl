{{ define "content" }}
{{- $p := .Folder }}
<div class="page-head">
	<div>
		<h2>Configure {{ $p.ID }}</h2>
		<p>What the instance adds to the folder: a title, who may run it, and which sources to read. The files in the folder stay as they are.</p>
	</div>
</div>

<form class="form" data-api="/api/projects/{{ $p.ID }}" data-method="PUT" data-next="/settings/projects?notice=Saved">
	<div class="field">
		<label for="pe-title">Title</label>
		<input type="text" id="pe-title" name="title" value="{{ index .Form "title" }}" autocomplete="off">
	</div>
	<div class="field">
		<label for="pe-description">Description</label>
		<input type="text" id="pe-description" name="description" value="{{ index .Form "description" }}" autocomplete="off">
	</div>
	<div class="field">
		<span class="label">Who may run <span class="faint">none checked is every runner</span></span>
		{{- template "whopicker" (list "allow" .AllowChoices) }}
	</div>
	{{- if .MoveFrom }}
	<div class="field">
		<label for="pe-workspace">Workspace <span class="faint">another one moves the checkout there, with the folder grants of the users; the runs stay under {{ $p.ID }}</span></label>
		<select id="pe-workspace" name="workspace" data-move="{{ $p.ID }}" data-move-from="{{ .MoveFrom }}" data-move-left="{{ join .MoveLeft "; " }}">
			<option value="_"{{ if eq .MoveFrom "_" }} selected{{ end }}>_ · the root</option>
			{{- range .WorkspaceList }}{{ if ne .Folder "_" }}
			<option value="{{ .Folder }}"{{ if eq .Folder $.MoveFrom }} selected{{ end }}>{{ .Folder }}{{ with .Title }} · {{ . }}{{ end }}</option>
			{{- end }}{{ end }}
		</select>
	</div>
	{{- end }}
	<div class="field">
		<label for="pe-sources">Sources <span class="faint">files to read, comma separated; empty detects them</span></label>
		<input type="text" id="pe-sources" name="sources" data-type="list" value="{{ index .Form "sources" }}" placeholder="atkins.yml, Makefile" spellcheck="false" autocomplete="off">
	</div>
	<div class="field">
		<label for="pe-hours">Hours <span class="faint">when the jobs may run, one rule per line; the last rule for a day wins; empty is always</span></label>
		<textarea id="pe-hours" name="hours" rows="3" spellcheck="false" autocomplete="off" placeholder="mon-fri 9-17&#10;fri 9-13">{{ index .Form "hours" }}</textarea>
	</div>
	<div class="field">
		<label class="check"><input type="checkbox" id="pe-recursive" name="recursive"{{ if or (not $.Project) (not $.Project.Shallow) }} checked{{ end }}> Recursive <span class="faint">Off: only the project and its top-level folders</span></label>
	</div>
	<div class="field">
		<label class="check"><input type="checkbox" id="pe-sidebar" name="sidebar"{{ if or (not $.Project) (not $.Project.SidebarHidden) }} checked{{ end }}> Show in sidebar <span class="faint">off keeps the project and its folders out of the sidebar; their pages stay</span></label>
	</div>
	{{- if $p.Jobs }}
	{{- $hours := $.Project.JobHoursMap }}
	<div class="field">
		<span class="label">Per job <span class="faint">who may run it and when, over the folder's; none checked and empty hours inherit</span></span>
		<div class="table-wrap">
		<table class="table cards">
			<thead><tr><th>Job</th><th>Who may run</th><th>Hours</th></tr></thead>
			<tbody>
			{{- range $p.Jobs }}
			<tr>
				<td class="key">{{ .Name }}</td>
				<td data-label="Who may run">{{ template "whopicker" (list (printf "job_allow.%s" .ID) (index $.JobAllowChoices .ID)) }}</td>
				<td data-label="Hours"><input type="text" name="job_hours.{{ .ID }}" value="{{ index $hours .ID }}" placeholder="mon-fri 9-17" spellcheck="false" autocomplete="off"></td>
			</tr>
			{{- end }}
			</tbody>
		</table>
		</div>
	</div>
	{{- end }}
	<div class="form-foot">
		<a class="btn" href="/settings/projects">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{ end }}
