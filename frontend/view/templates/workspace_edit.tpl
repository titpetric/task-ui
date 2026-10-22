{{ define "content" }}
{{- $ws := .Workspace }}
{{- $nested := index .Nested $ws.Folder }}{{ $default := eq $ws.Folder "_" }}
<div class="page-head">
	<div>
		<h2>Configure {{ $ws.Folder }}</h2>
		<p>A workspace's title, description and Recursive. Its folder, repository and branch stay: another checkout is another workspace.{{ if $default }} The default workspace is the root, and stays.{{ end }}</p>
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
	<p class="dim">Downloads <span class="mono-break">{{ $ws.Folder }}.taskui-ws</span>, which Import on another task-ui's Workspaces page reads: the workspace, its projects, its env files, its secrets as they are encrypted, and its keys with their grants. The file is encrypted to the passphrase; anyone with both holds the keys. Schedules, workflows, runs and artifacts stay here.</p>
	<form class="form" data-export="/api/workspaces/{{ $ws.Folder }}/export">
		<div class="field">
			<label for="wx-passphrase">Passphrase <span class="faint">the import asks for it</span></label>
			<input type="password" id="wx-passphrase" name="passphrase" autocomplete="new-password" required>
		</div>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-download"/></svg>Export</button>
		</div>
	</form>
</section>
{{- end }}

{{- /* workspacekeys are a workspace's keys: the encryption key its
     secrets use, which no run gets, and the SSH keys its runs get, each
     to every run in the workspace or to the roles and users granted. */}}
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
	<p class="dim">The workspace's own keys, kept apart from the checkout. Its runs get its SSH keys in their ssh agent, as <b>ws:</b>&lt;name&gt;, beside the instance keys their policies give them: a key without a grant goes to every run in the workspace, a grant narrows it to the roles and the users checked. The encryption key encrypts the workspace's secrets; no run gets it. They live in <span class="mono-break">{{ .KeysDir }}</span>.</p>
	<div class="table-wrap">
	<table class="table keys">
		<thead><tr><th>Comment</th><th>Fingerprint</th><th>Who gets it</th><th>Date</th><th>Actions</th></tr></thead>
		<tbody>
		{{- with .EncryptKey }}
			<tr class="encrypt-key">
				<td class="key">{{ .Comment }} <span class="badge">{{ .Name }}</span> <span class="badge signal" title="The workspace's secrets are encrypted with this key"><svg class="ic"><use href="#i-secret"/></svg> encryption</span></td>
				<td class="fingerprint"><span class="kind dim">{{ .Type }}</span>{{ .Fingerprint }}</td>
				<td class="nowrap"><span class="faint">no run</span></td>
				<td class="nowrap" title="{{ stamp .Modified }}">{{ ago .Modified }}</td>
				<td class="actions">
					<button type="button" class="btn-icon" data-copy="{{ .PublicKey }}" title="Copy public key" aria-label="Copy public key"><svg class="ic"><use href="#i-copy"/></svg></button>
					<a class="btn-icon" href="{{ $base }}/upload?encrypt=1" title="Upload: replace it with a key you have" aria-label="Upload an encryption key"><svg class="ic"><use href="#i-download"/></svg></a>
					<form data-api="{{ $api }}/generate" data-method="POST" data-next="/settings/workspaces/{{ $ws.Folder }}?rekeyed={rekeyed}#keys" data-confirm="Rotate the encryption key of {{ $ws.Folder }}? The current key is kept as a backup, and the workspace's secrets are written again for the new one." data-confirm-ok="Rotate">
						<input type="hidden" name="encrypt" value="true" data-type="bool">
						<input type="hidden" name="comment" value="{{ .Comment }}">
						<button class="btn-icon" type="submit" title="Rotate: generate a new encryption key" aria-label="Rotate the encryption key"><svg class="ic"><use href="#i-rotate"/></svg></button>
					</form>
				</td>
			</tr>
		{{- else }}
			<tr class="encrypt-key">
				<td class="key"><span class="badge bad">missing</span> <span class="badge signal"><svg class="ic"><use href="#i-secret"/></svg> encryption</span></td>
				<td colspan="3" class="dim">No encryption key: the workspace stores no secrets until it has one.</td>
				<td class="actions">
					<a class="btn-icon" href="{{ $base }}/upload?encrypt=1" title="Upload an encryption key you have" aria-label="Upload an encryption key"><svg class="ic"><use href="#i-download"/></svg></a>
					<form data-api="{{ $api }}/generate" data-method="POST" data-next="/settings/workspaces/{{ $ws.Folder }}?notice=Encryption+key+added#keys">
						<input type="hidden" name="encrypt" value="true" data-type="bool">
						<input type="hidden" name="comment" value="task-ui {{ $ws.Folder }}">
						<button class="btn-icon" type="submit" title="Generate an encryption key" aria-label="Generate an encryption key"><svg class="ic"><use href="#i-key"/></svg></button>
					</form>
				</td>
			</tr>
		{{- end }}
		{{- $any := false }}
		{{- range .Keys }}{{ if not .Backup }}{{ $any = true }}
			<tr>
				<td class="key">{{ .Comment }} <span class="badge">{{ .Name }}</span></td>
				<td class="fingerprint"><span class="kind dim">{{ .Type }}</span>{{ .Fingerprint }}</td>
				{{- $roles := index $.KeyPolicies .Name }}{{ $users := index $.KeyUsers .Name }}
				<td class="nowrap">{{ range $roles }}<span class="badge">{{ . }}</span> {{ end }}{{ range $users }}{{ template "avatar" (list $ .) }} {{ end }}{{ if not (or $roles $users) }}<span class="dim">every run</span>{{ end }}</td>
				<td class="nowrap" title="{{ stamp .Modified }}">{{ ago .Modified }}</td>
				<td class="actions">
					<button type="button" class="btn-icon" data-copy="{{ .PublicKey }}" title="Copy public key" aria-label="Copy public key"><svg class="ic"><use href="#i-copy"/></svg></button>
					<a class="btn-icon" href="{{ $base }}/{{ .Name }}" title="Grants: whose runs get it" aria-label="Edit grants"><svg class="ic"><use href="#i-edit"/></svg></a>
					{{- if .Kind }}
					<form data-api="{{ $api }}/generate" data-method="POST" data-next="reload" data-confirm="Rotate {{ .Name }}? The current pair is kept as a backup. Hosts that trust it need the new public key." data-confirm-ok="Rotate">
						<input type="hidden" name="type" value="{{ .Kind }}">
						<input type="hidden" name="name" value="{{ .Label }}">
						<input type="hidden" name="comment" value="{{ .Comment }}">
						<button class="btn-icon" type="submit" title="Rotate: generate a new pair" aria-label="Rotate key"><svg class="ic"><use href="#i-rotate"/></svg></button>
					</form>
					{{- end }}
					<button type="button" class="btn-icon danger" data-delete="{{ $api }}/{{ .Name }}" data-confirm="Delete {{ .Name }}? The workspace's runs lose it." data-next="reload" title="Delete key" aria-label="Delete key"><svg class="ic"><use href="#i-trash"/></svg></button>
				</td>
			</tr>
		{{- end }}{{ end }}
		{{- if not $any }}
			<tr class="empty"><td colspan="5">No SSH key yet. Generate one, or upload the private key the workspace's runs use.</td></tr>
		{{- end }}
		{{- range .Keys }}{{ if .Backup }}
			<tr>
				<td class="key dim">{{ .Comment }} <span class="badge">{{ .Name }}</span></td>
				<td class="fingerprint dim"><span class="kind faint">{{ .Type }}</span>{{ .Fingerprint }}</td>
				<td class="nowrap"><span class="faint">backup, no run</span></td>
				<td class="nowrap dim" title="{{ stamp .Modified }}">{{ ago .Modified }}</td>
				<td class="actions">
					<button type="button" class="btn-icon" data-copy="{{ .PublicKey }}" title="Copy public key" aria-label="Copy public key"><svg class="ic"><use href="#i-copy"/></svg></button>
					<button type="button" class="btn-icon danger" data-delete="{{ $api }}/{{ .Name }}" data-confirm="Delete the backup {{ .Name }}?" data-next="reload" title="Delete backup" aria-label="Delete backup"><svg class="ic"><use href="#i-trash"/></svg></button>
				</td>
			</tr>
		{{- end }}{{ end }}
		</tbody>
	</table>
	</div>
</section>
{{- end }}
