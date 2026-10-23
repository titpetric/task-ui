{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Workspaces</h2>
		<p>Checkouts under the root that hold projects. The folder name is the projects' namespace; <code>_</code> is the root itself.</p>
	</div>
	<div class="actions">
		<button type="button" class="btn" data-dialog-open="import-dialog"><svg class="ic"><use href="#i-upload"/></svg>Import</button>
		<a class="btn btn-primary" href="/settings/workspaces/new">Add workspace</a>
	</div>
</div>

<section class="block">
	<h3>Workspaces</h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Folder</th><th>Title</th><th>Repository</th><th class="hide-sm">Projects</th><th></th></tr></thead>
		<tbody>
		{{- $first := 0 }}{{ $last := len .WorkspaceList }}
		{{- with .WorkspaceList }}{{ if eq (index . 0).Folder "_" }}{{ $first = 1 }}{{ end }}{{ end }}
		{{- range $i, $ws := .WorkspaceList }}
			{{- $nested := index $.Nested .Folder }}{{ $default := eq .Folder "_" }}
			<tr>
				<td class="key nowrap"><a href="{{ $.Href .Folder }}"><svg class="ic"><use href="#i-folder-open"/></svg> {{ .Folder }}</a></td>
				<td class="desc">{{ .Title }}{{ with .Description }} <span class="faint">{{ . }}</span>{{ end }}</td>
				<td class="mono-break">{{ if $default }}<span class="faint">the root</span>{{ else }}{{ .SourceURL }}{{ with .Branch }} <span class="badge">{{ . }}</span>{{ end }}{{ end }}</td>
				<td class="hide-sm">{{ range $i, $id := $nested }}{{ if $i }}, {{ end }}<a href="{{ $.Href $id }}">{{ $id }}</a>{{ else }}<span class="faint">none yet</span>{{ end }}</td>
				<td class="actions">{{ template "menu" (menu (print "Actions on " .Folder)
					(item "folder-open" "Open" "href" ($.Href .Folder))
					(item "edit" "Configure" "href" (print "/settings/workspaces/" .Folder))
					sep
					(and (not $default) (ne $i $first) (item "up" "Move up" "data-workspace-move" "up" "data-folder" .Folder))
					(and (not $default) (ne (add $i 1) $last) (item "down" "Move down" "data-workspace-move" "down" "data-folder" .Folder))
					(and (not $default) (danger "trash" "Remove" "data-delete" (print "/api/workspaces/" .Folder) "data-confirm" (print "Remove the workspace " .Folder " with its checkout, keys and secrets" (or (and $nested (print ", and the projects inside it: " (join $nested ", "))) "") "? You can clone it again from " .SourceURL ".") "data-confirm-ok" "Remove" "data-next" "reload"))
				) }}</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="5">No workspaces yet. Add one from a git repository, then clone projects into it with Add project.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
<p class="faint" style="margin-top:0.6rem">Root: {{ .Root }}</p>

<dialog class="confirm workspace-import" id="import-dialog" aria-labelledby="import-title">
	<form class="form" data-import="/api/workspaces/import">
		<h3 id="import-title">Import a workspace</h3>
		<p>Adds a workspace from a bundle another task-ui exported, opened with its passphrase. Import clones the workspace's repository into the folder, adds its SSH keys to the key folder, encrypts its secrets with this instance's encryption key, adds its projects, env files and grants, and then clones each project. If a project fails to clone, it stays listed with git's output. If anything before the projects fails, task-ui keeps nothing. To import a bundle of the default workspace, _, enter a folder name.</p>
		<div class="field">
			<label for="wi-bundle">Bundle <span class="faint">a .tui.age file</span></label>
			<input type="file" id="wi-bundle" name="bundle" accept=".age" required>
		</div>
		<div class="field">
			<label for="wi-passphrase">Passphrase</label>
			<input type="password" id="wi-passphrase" name="passphrase" autocomplete="off" spellcheck="false" required>
		</div>
		<div class="field">
			<label for="wi-folder">Folder <span class="faint">optional, defaults to the bundle's folder; one name, no slashes</span></label>
			<input type="text" id="wi-folder" name="folder" pattern="[A-Za-z0-9._-]+" spellcheck="false" autocomplete="off">
		</div>
		<div class="confirm-foot">
			<button type="button" class="btn" data-dialog-close>Cancel</button>
			<button class="btn btn-primary" type="submit">Import</button>
		</div>
	</form>
</dialog>
{{ end }}
