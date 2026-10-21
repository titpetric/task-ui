{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Workspaces</h2>
		<p>Checkouts under the root that hold projects: an organisation's tooling, with the projects cloned inside it. The folder is the namespace of its projects; any number of workspaces may come from one repository. The default workspace, _, is the root itself: it holds the projects cloned at the top, comes first, and stays.</p>
	</div>
	<div class="actions">
		<a class="btn btn-primary" href="/settings/workspaces/new">Add workspace</a>
	</div>
</div>

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
			<td class="actions">
				{{- if not $default }}
				<button type="button" class="btn-icon" data-workspace-move="up" data-folder="{{ .Folder }}" title="Move up" aria-label="Move {{ .Folder }} up"{{ if eq $i $first }} disabled{{ end }}>↑</button>
				<button type="button" class="btn-icon" data-workspace-move="down" data-folder="{{ .Folder }}" title="Move down" aria-label="Move {{ .Folder }} down"{{ if eq (add $i 1) $last }} disabled{{ end }}>↓</button>
				{{- end }}
				<a class="btn-icon" href="/settings/workspaces/{{ .Folder }}" title="Configure {{ .Folder }}" aria-label="Configure {{ .Folder }}"><svg class="ic"><use href="#i-edit"/></svg></a>
				{{- if not $default }}{{ template "workspacedelete" (list . $nested "reload") }}{{ end }}
			</td>
		</tr>
	{{- else }}
		<tr class="empty"><td colspan="5">No workspaces yet. Add one from a git repository, then clone projects into it from Add project.</td></tr>
	{{- end }}
	</tbody>
</table>
</div>
<p class="faint" style="margin-top:0.6rem">Root: {{ .Root }}</p>

<section class="block workspace-import">
	<div class="block-head">
		<h3>Import</h3>
	</div>
	<p class="dim">Adds a workspace from a bundle another task-ui exported. It clones the workspace's repository under the folder, writes its keys, adds its projects, env files, secrets and grants, then clones each project; a project that does not clone stays, with git's output shown. Up to the projects, a failure keeps nothing. A bundle of the default workspace, _, imports as a named one: name its folder.</p>
	<form class="form" data-import="/api/workspaces/import">
		<div class="field">
			<label for="wi-bundle">Bundle <span class="faint">a .taskui-ws file</span></label>
			<input type="file" id="wi-bundle" name="bundle" accept=".taskui-ws" required>
		</div>
		<div class="field">
			<label for="wi-passphrase">Passphrase</label>
			<input type="password" id="wi-passphrase" name="passphrase" autocomplete="off" required>
		</div>
		<div class="field">
			<label for="wi-folder">Folder <span class="faint">optional, the bundle's own otherwise; one name, no slashes</span></label>
			<input type="text" id="wi-folder" name="folder" pattern="[A-Za-z0-9._-]+" spellcheck="false" autocomplete="off">
		</div>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Import</button>
		</div>
	</form>
</section>
{{ end }}
