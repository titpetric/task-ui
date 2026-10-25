{{ define "content" }}
{{- $ws := .Workspace }}
{{- $nested := index .Nested $ws.Folder }}{{ $default := eq $ws.Folder "_" }}{{ $secrets := index .SecretCounts $ws.Folder }}
<div class="page-head">
	<div>
		<h2>Workspace settings <span class="badge">{{ $ws.Folder }}</span></h2>
		<p>The workspace's settings, its home repository, its projects and the SSH keys its runs get. The workspace holds its home, the <code>HOME</code> of its runs, and its projects side by side under <code>src/</code>; its folder stays as added.{{ if $default }} The default workspace is the root itself: deleting it removes the root's home and projects, and the instance runs without one.{{ end }}</p>
	</div>
	<div class="actions">
		{{- if .Can "project/edit" }}
		<a class="btn btn-primary" href="/settings/projects/new?workspace={{ $ws.Folder }}"><svg class="ic"><use href="#i-plus"/></svg>Add project</a>
		{{- end }}
		{{- if and .KeysDir .User .User.CanAdmin }}
		<button type="button" class="btn" data-dialog-open="export-dialog"><svg class="ic"><use href="#i-download"/></svg>Export</button>
		{{- end }}
		{{ template "menu" (menu (print "Actions on " $ws.Folder)
			(item "folder-open" "Open" "href" ($.Href $ws.Folder))
			(and (.Can "workspace/delete") $secrets (item "secret" (deleteBlocked $secrets) "href" (print "/settings/workspaces/" $ws.Folder "/secrets")))
			(and (.Can "workspace/delete") (not $secrets) (danger "trash" "Delete" "data-delete" (print "/api/workspaces/" $ws.Folder) "data-confirm" (deleteWorkspaceConfirm $ws.Folder $nested) "data-confirm-ok" "Delete" "data-next" "/settings/workspaces?notice=Workspace+deleted"))
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
		<div class="field">
			<label for="we-sidebar">Sidebar</label>
			<input type="number" id="we-sidebar" name="sidebar_limit" data-type="int" min="0" step="1" placeholder="all" value="{{ if gt $ws.SidebarLimit 0 }}{{ $ws.SidebarLimit }}{{ end }}" autocomplete="off">
			<small class="faint">How many projects the sidebar lists for this workspace, the most recently committed first. Empty or 0 lists them all, by name.</small>
		</div>
		<div class="form-foot">
			<a class="btn" href="/settings/workspaces">Cancel</a>
			<button class="btn btn-primary" type="submit">Save</button>
		</div>
	</form>
</section>

{{ template "workspacehome" . }}

<section class="block" id="projects">
	{{- template "blockhead" (head "Projects" (or (and $default "in src/ at the root") (print "inside " $ws.Folder "/src/")) (and (not .Entries) (print "No projects in " $ws.Folder " yet."))) }}
	{{- with .Entries }}{{ template "projecttable" (list $ .) }}{{ end }}
</section>
{{- if .KeysDir }}
{{ template "workspacekeys" . }}
{{ template "workspaceexport" . }}
{{- end }}
{{ end }}

{{- /* workspacehome is a workspace's home, the HOME of its runs: the
     repository of its checkout with its branch and commit, or
     "The home is an empty directory.", with Pull now, which runs the
     home's @git:pull, Replace, a dialog that clones a repository in its
     place, and Remove, which empties it. */}}
{{- define "workspacehome" }}
{{- $ws := .Workspace }}{{ $edit := .Can "workspace/edit" }}
{{- with .Home }}{{ $home := . }}
<section class="block workspace-home" id="home">
	{{- template "blockhead" (head "Home repository" (print $home.Folder "/, the HOME of the workspace's runs") (and $home.Empty "The home is an empty directory.")
		(and $home.Pull (item "download" "Pull now" "data-run" "" "data-folder" $home.Folder "data-dir" $home.Pull.Dir "data-job" $home.Pull.ID "data-title" (print $home.Folder " / " $home.Pull.Name)))
		(and $edit (item "plus" (or (and (not $home.Empty) "Replace") "Clone") "data-dialog-open" "home-replace-dialog"))
		(and $edit (not $home.Empty) (danger "" "Remove" "data-api" (print "/api/workspaces/" $ws.Folder "/home") "data-method" "DELETE" "data-confirm" (print "Remove the home repository of " $ws.Folder "? Its checkout goes, and the home becomes an empty directory. Runs keep HOME there.") "data-confirm-ok" "Remove" "data-next" (print "/settings/workspaces/" $ws.Folder "?notice=Home+repository+removed")))) }}
	{{- if not $home.Empty }}
	<p><a class="mono-break" href="{{ $.Href $home.Folder }}"><svg class="ic"><use href="#i-folder-open"/></svg> {{ $home.Folder }}</a> <span class="faint mono-break">{{ $home.URL }}</span>{{ with $home.Branch }} <span class="badge">{{ . }}</span>{{ end }}</p>
	{{- with $home.Checkout }}
	<p class="dim">{{ .Branch }} <span class="faint">{{ .Commit }}</span> {{ .Subject }}{{ if not .When.IsZero }} <span class="faint">{{ ago .When }}</span>{{ end }}; pulled before every run.</p>
	{{- else }}
	<p class="dim">The home holds no checkout: the clone failed, or someone replaced the directory by hand. Replace clones it again.</p>
	{{- end }}
	{{- end }}
	<p class="dim">Every run in {{ $ws.Folder }} gets <code>HOME</code> here: <code>bash -l</code> reads its <code>.bash_profile</code>, atkins its <code>.atkins/skills</code> and <code>.atkins/atkins.db</code>, Claude its <code>CLAUDE.md</code> and <code>.claude</code> skills. Registry logins, Claude Code's settings and the build caches stay in the server's home.</p>
</section>
{{- if $edit }}
<dialog class="confirm" id="home-replace-dialog" aria-labelledby="home-replace-title">
	<form class="form" data-api="/api/workspaces/{{ $ws.Folder }}/home" data-method="PUT" data-next="/settings/workspaces/{{ $ws.Folder }}?notice=Home+repository+cloned">
		<h3 id="home-replace-title">{{ if $home.Empty }}Clone{{ else }}Replace{{ end }} the home repository</h3>
		<p>Removes <span class="mono-break">{{ $home.Folder }}/</span> and clones the repository in its place. task-ui refuses while a run of {{ $ws.Folder }} is active. If the clone fails, you see git's output and the home stays an empty directory.</p>
		<div class="field">
			<label for="hr-url">Repository URL</label>
			<input type="text" id="hr-url" name="url" value="{{ $home.URL }}" placeholder="git@github.com:org/workspace.git" spellcheck="false" autocomplete="off" required>
		</div>
		<div class="field">
			<label for="hr-branch">Branch <span class="faint">optional</span></label>
			<input type="text" id="hr-branch" name="branch" value="{{ $home.Branch }}" spellcheck="false" autocomplete="off">
		</div>
		<div class="confirm-foot">
			<button type="button" class="btn" data-dialog-close>Cancel</button>
			<button class="btn btn-primary" type="submit">Clone</button>
		</div>
	</form>
</dialog>
{{- end }}
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
	<div class="table-wrap fit">
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
