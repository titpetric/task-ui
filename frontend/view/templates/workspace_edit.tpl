{{ define "content" }}
{{- $ws := .Workspace }}
{{- $nested := index .Nested $ws.Folder }}{{ $default := eq $ws.Folder "_" }}
<div class="page-head">
	<div>
		<h2>Configure {{ $ws.Folder }}</h2>
		<p>The workspace's settings, its projects and the SSH keys its runs get. Its folder, repository and branch stay as cloned; for another checkout, add another workspace.{{ if $default }} The default workspace is the root, and it stays.{{ end }}</p>
	</div>
	<div class="actions">
		{{- if .Can "project/edit" }}
		<a class="btn btn-primary" href="/settings/projects/new?workspace={{ $ws.Folder }}"><svg class="ic"><use href="#i-plus"/></svg>Add project</a>
		{{- end }}
		{{- if and .KeysDir (.Can "workspace/edit") }}
		<button type="button" class="btn" data-dialog-open="export-dialog"><svg class="ic"><use href="#i-download"/></svg>Export</button>
		{{- end }}
		{{ template "menu" (menu (print "Actions on " $ws.Folder)
			(item "folder-open" "Open" "href" ($.Href $ws.Folder))
			(and (not $default) (.Can "workspace/delete") (danger "trash" "Delete" "data-delete" (print "/api/workspaces/" $ws.Folder) "data-confirm" (deleteWorkspaceConfirm $ws.Folder $ws.SourceURL $nested) "data-confirm-ok" "Delete" "data-next" "/settings/workspaces?notice=Workspace+deleted"))
		) }}
	</div>
</div>

<section class="block">
	<h3>Settings <small class="mono-break">{{ if $default }}the root, {{ .Root }}{{ else }}{{ $ws.SourceURL }}{{ with $ws.Branch }}, {{ . }}{{ end }}{{ end }}</small></h3>
	<form class="form" data-api="/api/workspaces/{{ $ws.Folder }}" data-method="PUT" data-next="/settings/workspaces/{{ $ws.Folder }}?notice=Saved">
		<div class="field">
			<label for="we-title">Title</label>
			<input type="text" id="we-title" name="title" value="{{ index .Form "title" }}" autocomplete="off">
		</div>
		<div class="field">
			<label for="we-description">Description</label>
			<input type="text" id="we-description" name="description" value="{{ index .Form "description" }}" autocomplete="off">
		</div>
		{{- template "recursivefield" (list "we-recursive" (not $ws.Shallow)) }}
		<div class="form-foot">
			<a class="btn" href="/settings/workspaces">Cancel</a>
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
</section>

<section class="block" id="projects">
	{{- template "blockhead" (head "Projects" (or (and $default "at the top of the root") (print "inside " $ws.Folder "/")) (and (not .Entries) (print "No projects in " $ws.Folder " yet."))) }}
	{{- with .Entries }}{{ template "projecttable" (list $ .) }}{{ end }}
</section>
{{- if .KeysDir }}
{{ template "workspacekeys" . }}
{{ template "workspaceexport" . }}
{{- end }}
{{ end }}

{{- /* workspaceexport is the dialog that downloads a workspace as a
     bundle another task-ui imports, an age file encrypted to a
     passphrase typed twice or generated. */}}
{{- define "workspaceexport" }}
{{- $ws := .Workspace }}
<dialog class="confirm workspace-export" id="export-dialog" aria-labelledby="export-title">
	<form class="form" data-export="/api/workspaces/{{ $ws.Folder }}/export">
		<h3 id="export-title">Export {{ $ws.Folder }}</h3>
		<p>Downloads <span class="mono-break">{{ $ws.Folder }}.tui.age</span> for Import on another task-ui's Workspaces page: the workspace, its projects, env files and grants, its secrets and the SSH keys its runs get, encrypted to the passphrase. Anyone with the file and the passphrase has the secrets and the keys. No instance key and no encryption key goes; schedules, workflows, runs and artifacts stay here.</p>
		<div class="field">
			<label for="wx-passphrase">Passphrase <span class="faint">at least 12 characters</span></label>
			<div class="passphrase">
				<input type="password" id="wx-passphrase" name="passphrase" minlength="12" autocomplete="new-password" spellcheck="false" required>
				<button type="button" class="btn-icon" data-copy="" data-passphrase-copy title="Copy passphrase" aria-label="Copy passphrase" hidden><svg class="ic"><use href="#i-copy"/></svg></button>
				<button type="button" class="btn" data-passphrase-generate>Generate</button>
			</div>
		</div>
		<div class="field" data-passphrase-confirm>
			<label for="wx-confirm">Confirm passphrase</label>
			<input type="password" id="wx-confirm" name="confirm" autocomplete="new-password" spellcheck="false">
		</div>
		<p class="passphrase-note" data-passphrase-note hidden>Store this passphrase before you export: task-ui keeps no copy, and nothing recovers it.</p>
		<div class="confirm-foot">
			<button type="button" class="btn" data-dialog-close>Cancel</button>
			<button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-download"/></svg>Export</button>
		</div>
	</form>
</dialog>
{{- end }}

{{- /* workspacekeys are a workspace's SSH keys, which its runs get,
     each to every run in the workspace or to the roles and users
     linked: the keys table every page shares, or one line with Upload
     key and Generate key while there is none. */}}
{{- define "workspacekeys" }}
{{- $ws := .Workspace }}{{ $base := printf "/settings/workspaces/%s/keys" $ws.Folder }}{{ $edit := .Can "workspace/edit" }}
<section class="block workspace-keys" id="keys">
	{{- template "blockhead" (head "Keys" "SSH keys the workspace's runs get" (and (not .Keys) "No keys added.")
		(and $edit (item "upload" "Upload key" "href" (print $base "/upload")))
		(and $edit (item "key" "Generate key" "href" (print $base "/generate")))) }}
	{{- with .Keys }}
	<p class="dim">A key goes to every run in the workspace, or to the roles and users its links name, through the run's ssh agent.</p>
	<div class="table-wrap">
	<table class="table keys key-table cards">
		{{ template "keyhead" }}
		<tbody>
		{{- range . }}{{ template "keyrow" (list $ . (print "/settings/workspaces/" $ws.Folder) $ws.Folder) }}{{ end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>
{{- end }}
