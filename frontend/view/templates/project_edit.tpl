{{ define "content" }}
{{- $p := .Folder }}
<div class="page-head">
	<div>
		<h2>Project settings <span class="badge">{{ $p.ID }}</span></h2>
		<p>Settings task-ui keeps for the folder: a title, who can run it, and which sources to read. task-ui doesn't change the files in the folder.</p>
	</div>
	{{- if and $.Checkout ($.Can "project/delete") }}
	<div class="actions">
		{{ template "menu" (menu (print "Actions on " $p.ID)
			(danger "trash" "Remove checkout" "data-delete" (print "/api/projects/" $p.ID) "data-confirm" (removeConfirm $p.ID $.Checkout.Remote) "data-confirm-ok" "Remove checkout" "data-next" (print "/settings/projects?notice=" (urlquery "Removed " $p.ID)))
		) }}
	</div>
	{{- end }}
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
	<div class="field" role="group" aria-labelledby="pe-allow-label">
		<span class="label" id="pe-allow-label">Who can run</span>
		{{- template "whofield" (who "allow" "set" .Allow .Allow .Users "everyone who may run can") }}
		<small class="faint">With nothing checked, everyone who may run can.</small>
	</div>
	{{- if .MoveFrom }}
	<div class="field">
		<label for="pe-workspace">Workspace</label>
		<select id="pe-workspace" name="workspace" data-move="{{ $p.ID }}" data-move-from="{{ .MoveFrom }}" data-move-left="{{ join .MoveLeft "; " }}">
			<option value="_"{{ if eq .MoveFrom "_" }} selected{{ end }}>_ · the root</option>
			{{- range .WorkspaceList }}{{ if ne .Folder "_" }}
			<option value="{{ .Folder }}"{{ if eq .Folder $.MoveFrom }} selected{{ end }}>{{ .Folder }}{{ with .Title }} · {{ . }}{{ end }}</option>
			{{- end }}{{ end }}
		</select>
		<small class="faint">Picking another workspace moves the checkout there, with the users' folder access; earlier runs stay under {{ $p.ID }}.</small>
	</div>
	{{- end }}
	<div class="field">
		<label for="pe-sources">Sources</label>
		<input type="text" id="pe-sources" name="sources" data-type="list" value="{{ index .Form "sources" }}" placeholder="atkins.yml, Makefile" spellcheck="false" autocomplete="off">
		<small class="faint">Files to read, comma separated; leave empty to detect them.</small>
	</div>
	{{- template "recursivefield" (list "pe-recursive" (or (not $.Project) (not $.Project.Shallow))) }}
	<div class="field">
		<label class="check"><input type="checkbox" id="pe-sidebar" name="sidebar"{{ if or (not $.Project) (not $.Project.SidebarHidden) }} checked{{ end }}> Show in sidebar</label>
		<small class="faint">Off hides the project and its folders from the sidebar; their pages still work.</small>
	</div>
	{{- if $p.Jobs }}
	<div class="field">
		<span class="label">Per job</span>
		<div class="table-wrap">
		<table class="table cards">
			<thead><tr><th>Job</th><th>Who can run</th></tr></thead>
			<tbody>
			{{- range $p.Jobs }}
			<tr>
				<td class="key">{{ .Name }}</td>
				<td data-label="Who can run">{{ $allow := index $.JobAllow .ID }}{{ template "whofield" (who (printf "job_allow.%s" .ID) "set" $allow $allow $.Users "the folder's Who can run applies").AsCompact }}</td>
			</tr>
			{{- end }}
			</tbody>
		</table>
		</div>
		<small class="faint">Who can run each job and when, on top of the folder's settings; leave unchecked and empty to use the folder's.</small>
	</div>
	{{- end }}
	<div class="form-foot">
		<a class="btn" href="/settings/projects">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{ end }}
