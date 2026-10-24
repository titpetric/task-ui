{{ define "content" }}
{{- $ws := .Workspace }}
{{- $nested := index .Nested $ws.Folder }}{{ $default := eq $ws.Folder "_" }}
<div class="page-head">
	<div>
		<h2>Workspace settings <span class="badge">{{ $ws.Folder }}</span></h2>
		<p>The workspace's settings, its home project, its projects and the SSH keys its runs get. The workspace is a directory of projects side by side; its folder stays as added.{{ if $default }} The default workspace is the root, and it stays.{{ end }}</p>
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
			(and (not $default) (.Can "workspace/delete") (danger "trash" "Delete" "data-delete" (print "/api/workspaces/" $ws.Folder) "data-confirm" (deleteWorkspaceConfirm $ws.Folder $nested) "data-confirm-ok" "Delete" "data-next" "/settings/workspaces?notice=Workspace+deleted"))
		) }}
	</div>
</div>

<section class="block">
	<h3>Settings <small class="mono-break">{{ if $default }}the root, {{ .Root }}{{ else }}{{ $ws.Dir .Root }}{{ end }}</small></h3>
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

{{ template "workspacehome" . }}

<section class="block" id="projects">
	{{- template "blockhead" (head "Projects" (or (and $default "at the top of the root") (print "inside " $ws.Folder "/")) (and (not .Entries) (print "No projects in " $ws.Folder " yet."))) }}
	{{- with .Entries }}{{ template "projecttable" (list $ .) }}{{ end }}
</section>
{{- if .KeysDir }}
{{ template "workspacekeys" . }}
{{ template "workspaceexport" . }}
{{- end }}
{{ end }}

{{- /* workspacehome is a workspace's home project, the HOME of its
     runs: the project and its repository, or "No home project.", with
     Change, a dialog that picks a project at the workspace's top,
     Add or Replace, a dialog that clones one and makes it home, and
     Remove, which unsets it and keeps the project. */}}
{{- define "workspacehome" }}
{{- $ws := .Workspace }}{{ $home := .HomeProject }}{{ $edit := .Can "workspace/edit" }}
<section class="block workspace-home" id="home">
	{{- template "blockhead" (head "Home project" "the HOME of the workspace's runs" (and (not $home) "No home project.")
		(and $edit .HomeChoices (item "edit" "Change" "data-dialog-open" "home-change-dialog"))
		(and $edit (.Can "project/edit") (item "plus" (or (and $home "Replace") "Add") "data-dialog-open" "home-replace-dialog"))
		(and $edit $home (danger "" "Remove" "data-api" (print "/api/workspaces/" $ws.Folder) "data-method" "PUT" "data-body" (homeBody $ws "") "data-confirm" (print "Unset " $home.Folder " as the home project of " $ws.Folder "? The project stays; runs get the server's HOME.") "data-confirm-ok" "Remove" "data-next" (print "/settings/workspaces/" $ws.Folder "?notice=Home+project+removed")))) }}
	{{- with $home }}
	<p><a class="mono-break" href="{{ $.Href .Folder }}"><svg class="ic"><use href="#i-folder-open"/></svg> {{ .Folder }}</a>{{ with .SourceURL }} <span class="faint mono-break">{{ . }}</span>{{ end }}{{ with .Branch }} <span class="badge">{{ . }}</span>{{ end }}</p>
	<p class="dim">Runs in {{ $ws.Folder }} get <code>HOME</code> here: atkins reads its <code>.atkins/skills</code> and Claude its <code>CLAUDE.md</code> and <code>.claude</code> skills. Stats, registry logins, the git identity and caches stay in the server's home.</p>
	{{- end }}
</section>
{{- if and $edit .HomeChoices }}
<dialog class="confirm" id="home-change-dialog" aria-labelledby="home-change-title">
	<form class="form" data-api="/api/workspaces/{{ $ws.Folder }}" data-method="PUT" data-next="/settings/workspaces/{{ $ws.Folder }}?notice=Home+project+changed">
		<h3 id="home-change-title">Change the home project</h3>
		<p>Runs in {{ $ws.Folder }} get the project's directory as <code>HOME</code>.</p>
		<input type="hidden" name="title" value="{{ index .Form "title" }}">
		<input type="hidden" name="description" value="{{ index .Form "description" }}">
		<div class="field">
			<label for="hc-project">Project</label>
			<select id="hc-project" name="home_project">
				{{- range .HomeChoices }}
				<option value="{{ . }}"{{ if and $home (eq . $home.Folder) }} selected{{ end }}>{{ . }}</option>
				{{- end }}
			</select>
		</div>
		<div class="confirm-foot">
			<button type="button" class="btn" data-dialog-close>Cancel</button>
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
</dialog>
{{- end }}
{{- if and $edit (.Can "project/edit") }}
<dialog class="confirm" id="home-replace-dialog" aria-labelledby="home-replace-title">
	<form class="form" data-api="/api/projects" data-method="POST" data-next="/settings/workspaces/{{ $ws.Folder }}?notice=Home+project+cloned">
		<h3 id="home-replace-title">{{ if $home }}Replace{{ else }}Add{{ end }} the home project</h3>
		<p>Clones the repository into {{ $ws.Folder }} as a project and makes it the home project.{{ with $home }} {{ .Folder }} stays as a project.{{ end }} If the clone fails, you see git's output and task-ui keeps nothing.</p>
		<input type="hidden" name="workspace" value="{{ $ws.Folder }}">
		<input type="hidden" name="home" value="true" data-type="bool">
		<div class="field">
			<label for="hr-url">Repository URL</label>
			<input type="text" id="hr-url" name="url" placeholder="git@github.com:org/workspace.git" spellcheck="false" autocomplete="off" required>
		</div>
		<div class="field">
			<label for="hr-id">Folder <span class="faint">optional, defaults to the repository's name</span></label>
			<input type="text" id="hr-id" name="id" pattern="[A-Za-z0-9._-]+" spellcheck="false" autocomplete="off">
		</div>
		<div class="field">
			<label for="hr-branch">Branch <span class="faint">optional</span></label>
			<input type="text" id="hr-branch" name="branch" spellcheck="false" autocomplete="off">
		</div>
		<div class="confirm-foot">
			<button type="button" class="btn" data-dialog-close>Cancel</button>
			<button class="btn btn-primary" type="submit">Clone</button>
		</div>
	</form>
</dialog>
{{- end }}
{{- end }}

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
