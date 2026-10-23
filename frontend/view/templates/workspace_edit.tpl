{{ define "content" }}
{{- $ws := .Workspace }}
{{- $nested := index .Nested $ws.Folder }}{{ $default := eq $ws.Folder "_" }}
<div class="page-head">
	<div>
		<h2>Configure {{ $ws.Folder }}</h2>
		<p>Change the workspace's title, description and Recursive. You can't change its folder, repository or branch; for another checkout, add another workspace.{{ if $default }} The default workspace is the root, and you can't remove it.{{ end }}</p>
	</div>
	{{- if or .KeysDir (not $default) }}
	<div class="actions">
		{{- if .KeysDir }}
		<button type="button" class="btn" data-dialog-open="export-dialog"><svg class="ic"><use href="#i-download"/></svg>Export</button>
		{{- end }}
		{{- if not $default }}
		{{- template "workspacedelete" (list $ws $nested "/settings/workspaces?notice=Workspace+removed" "Delete") }}
		{{- end }}
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
     linked. */}}
{{- define "workspacekeys" }}
{{- $ws := .Workspace }}{{ $base := printf "/settings/workspaces/%s/keys" $ws.Folder }}
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
	<table class="table keys key-table cards">
		{{ template "keyhead" }}
		<tbody>
		{{- range .Keys }}{{ template "keyrow" (list $ . (print "/settings/workspaces/" $ws.Folder) $ws.Folder) }}
		{{- else }}
			<tr class="empty"><td colspan="6">No SSH key yet. Generate one, or upload the private key the workspace's runs use.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{- end }}
