{{ define "content" }}
{{- $ws := .Workspace }}
{{- $nested := index .Nested $ws.Folder }}{{ $default := eq $ws.Folder "_" }}
<div class="page-head">
	<div>
		<h2>Configure {{ $ws.Folder }}</h2>
		<p>Change the workspace's title, description and Recursive. You can't change its folder, repository or branch; for another checkout, add another workspace.{{ if $default }} The default workspace is the root, and you can't remove it.{{ end }}</p>
	</div>
	{{- if not $default }}
	<div class="actions">
		{{- template "workspacedelete" (list $ws $nested "/settings/workspaces?notice=Workspace+removed") }}
	</div>
	{{- end }}
</div>

<form class="form" data-api="/api/workspaces/{{ $ws.Folder }}" data-method="PUT" data-next="/settings/workspaces?notice=Saved">
	<div class="field">
		<span class="label">Repository</span>
		<span class="mono-break">{{ if $default }}<span class="faint">the root, {{ .Root }}</span>{{ else }}{{ $ws.SourceURL }}{{ with $ws.Branch }} <span class="badge">{{ . }}</span>{{ end }}{{ end }}</span>
	</div>
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
		<span class="label">Projects <span class="faint">{{ if $default }}at the top of the root{{ else }}cloned inside it; removing the workspace removes them{{ end }}</span></span>
		<span>{{ range $i, $id := $nested }}{{ if $i }}, {{ end }}<a href="/settings/projects/{{ $id }}">{{ $id }}</a>{{ else }}<span class="faint">none yet</span>{{ end }}</span>
	</div>
	<div class="form-foot">
		<a class="btn" href="/settings/workspaces">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{{- if .KeysDir }}
{{ template "workspacekeys" . }}
{{ template "workspaceexport" . }}
{{- end }}
{{ end }}

{{- /* workspaceexport downloads a workspace as a bundle another task-ui
     imports, encrypted to a passphrase. */}}
{{- define "workspaceexport" }}
{{- $ws := .Workspace }}
<section class="block workspace-export">
	<div class="block-head">
		<h3>Export</h3>
	</div>
	<p class="dim">Downloads <span class="mono-break">{{ $ws.Folder }}.taskui-ws</span> for Import on another task-ui's Workspaces page. The file holds the workspace, its projects, its env files, its secrets (still encrypted) and its keys with their links. The passphrase encrypts the file: anyone with the file and the passphrase has the keys. Schedules, workflows, runs and artifacts stay here.</p>
	<form class="form" data-export="/api/workspaces/{{ $ws.Folder }}/export">
		<div class="field">
			<label for="wx-passphrase">Passphrase <span class="faint">you need it again to import</span></label>
			<input type="password" id="wx-passphrase" name="passphrase" autocomplete="new-password" required>
		</div>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-download"/></svg>Export</button>
		</div>
	</form>
</section>
{{- end }}

{{- /* workspacekeys are a workspace's SSH keys, which its runs get,
     each to every run in the workspace or to the roles and users
     linked. */}}
{{- define "workspacekeys" }}
{{- $ws := .Workspace }}{{ $base := printf "/settings/workspaces/%s/keys" $ws.Folder }}{{ $api := printf "/api/workspaces/%s/keys" $ws.Folder }}
<section class="block workspace-keys" id="keys">
	<div class="block-head">
		<h3>Keys</h3>
		<div class="actions">
			<a class="btn btn-sm" href="{{ $base }}/upload">Upload key</a>
			<a class="btn btn-sm btn-primary" href="{{ $base }}/generate">Generate key</a>
		</div>
	</div>
	<p class="dim">The workspace's own keys, named <span class="mono">workspace_{{ $ws.Folder }}_&lt;name&gt;_&lt;algorithm&gt;</span> in the key folder, <span class="mono-break">{{ .KeysDir }}</span>, out of the checkout. Its runs get its SSH keys in their ssh agent, next to the instance keys linked to them. A key with no link goes to every run in the workspace; links limit it to the checked roles and users. The instance's encryption key protects the workspace's secrets.</p>
	<div class="table-wrap">
	<table class="table keys">
		<thead><tr><th>Name</th><th>Fingerprint</th><th>Who gets it</th><th>Date</th><th>Actions</th></tr></thead>
		<tbody>
		{{- range .Keys }}
			<tr>
				<td class="key"><span class="mono">{{ .Name }}</span>{{ with .Comment }}<div class="faint">{{ . }}</div>{{ end }}</td>
				<td class="fingerprint"><span class="kind dim">{{ .Kind }}</span>{{ .Fingerprint }}</td>
				<td class="nowrap">{{ range .Roles }}<span class="badge">{{ . }}</span> {{ end }}{{ range .Users }}{{ template "avatar" (list $ .) }} {{ end }}{{ if not (or .Roles .Users) }}<span class="dim">every run</span>{{ end }}</td>
				<td class="nowrap" title="{{ stamp .CreatedAt }}">{{ ago .CreatedAt }}</td>
				<td class="actions">
					<button type="button" class="btn-icon" data-copy="{{ .PublicKey }}" title="Copy public key" aria-label="Copy public key"><svg class="ic"><use href="#i-copy"/></svg></button>
					<a class="btn-icon" href="{{ $base }}/{{ .Name }}" title="Links: whose runs get it" aria-label="Edit links"><svg class="ic"><use href="#i-edit"/></svg></a>
					<button type="button" class="btn-icon danger" data-delete="{{ $api }}/{{ .Name }}" data-confirm="Remove {{ .Name }}? The workspace's runs no longer get it. Its files stay in the key folder as a backup." data-next="reload" title="Remove key" aria-label="Remove key"><svg class="ic"><use href="#i-trash"/></svg></button>
				</td>
			</tr>
		{{- else }}
			<tr class="empty"><td colspan="5">No SSH key yet. Generate one, or upload the private key the workspace's runs use.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{- end }}
