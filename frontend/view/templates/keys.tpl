{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Keys</h2>
		<p>There are two kinds of key, and a key can't be both. An <b>encryption key</b> protects secrets. The instance has one for the integrations, and each workspace has one for its secrets; without it, a workspace can't store secrets. <b>SSH keys</b> let runs and git connect to hosts and remotes. Instance SSH keys go to the runs their policies name; a workspace's SSH keys go to that workspace's runs.</p>
	</div>
</div>

{{- range .KeyOverlaps }}
<div class="alert bad"><svg class="ic"><use href="#i-secret"/></svg> The key <b>{{ .Fingerprint }}</b> is both an encryption key, <span class="mono-break">{{ join .Encrypt ", " }}</span>, and an SSH key, <span class="mono-break">{{ join .Access ", " }}</span>. Runs and git skip it, so this page leaves it out. Remove the SSH copy, or use a separate key for SSH.</div>
{{- end }}

<section class="block key-overview">
	<div class="table-wrap">
	<table class="table key-purposes">
		<thead><tr><th>Scope</th><th>Secrets</th><th>Encryption key</th><th>SSH keys</th></tr></thead>
		<tbody>
			{{- $instance := 0 }}{{ range .Keys }}{{ if not .Backup }}{{ $instance = add $instance 1 }}{{ end }}{{ end }}
			<tr class="key-instance" id="instance">
				<td><b>Instance</b><div class="faint">every run, by policy</div></td>
				<td><a href="/settings/integrations">integrations</a><div class="faint">their settings</div></td>
				<td>
					{{- with .InstanceKey }}
					<span class="badge ok">set</span>
					<form class="inline" data-api="/api/keys/encrypt/rotate" data-method="POST" data-next="/settings/keys?rekeyed={rekeyed}&what=integration+setting" data-confirm="Rotate the instance's encryption key? task-ui generates a new key, keeps the current one as a backup, and re-encrypts the integrations' settings." data-confirm-ok="Rotate">
						<button class="btn btn-sm" type="submit" title="Generate a new encryption key"><svg class="ic"><use href="#i-rotate"/></svg> Rotate</button>
					</form>
					<div class="key-fp faint" title="{{ $.InstanceKeyDir }}">{{ .Fingerprint }}</div>
					{{- else }}
					<span class="badge bad">missing</span> <span class="faint">task-ui creates it on start</span>
					{{- end }}
				</td>
				<td>
					{{- if $instance }}
					<a href="#instance-ssh">{{ $instance }} key{{ if ne $instance 1 }}s{{ end }}</a>
					<div class="key-names">{{ range .Keys }}{{ if not .Backup }}<a class="key-chip" href="/settings/keys/{{ .Name }}" title="Policies of {{ .Name }}">{{ .Name }}</a>{{ end }}{{ end }}</div>
					{{- else }}
					<span class="faint">none: runs get no SSH key from the instance</span> <a href="/settings/keys/generate">Generate</a>
					{{- end }}
				</td>
			</tr>
			{{- range .KeyWorkspaces }}
			{{- $ws := .Workspace }}{{ $api := printf "/api/workspaces/%s/keys" $ws.Folder }}{{ $base := printf "/settings/workspaces/%s" $ws.Folder }}
			<tr id="ws-{{ $ws.Folder }}"{{ if not .Encrypt }} class="key-missing"{{ end }}>
				<td>
					<a href="{{ $base }}"><b>{{ or $ws.Title $ws.Folder }}</b></a>{{ if and $ws.Title (ne $ws.Title $ws.Folder) }} <span class="badge">{{ $ws.Folder }}</span>{{ end }}
					<div class="faint key-projects">{{ if .Projects }}{{ join .Projects ", " }}{{ else }}no projects{{ end }}</div>
				</td>
				<td>{{ if .Secrets }}<a href="/settings/secrets?workspace={{ $ws.Folder }}">{{ .Secrets }} secret{{ if ne .Secrets 1 }}s{{ end }}</a>{{ else if .Encrypt }}<a class="faint" href="/settings/secrets?workspace={{ $ws.Folder }}">none</a>{{ else }}<span class="faint">none</span>{{ end }}</td>
				<td>
					{{- with .Encrypt }}
					<span class="badge ok">set</span>
					<form class="inline" data-api="{{ $api }}/generate" data-method="POST" data-next="/settings/keys?rekeyed={rekeyed}#ws-{{ $ws.Folder }}" data-confirm="Rotate the encryption key of {{ $ws.Folder }}? task-ui generates a new key, keeps the current one as a backup, and re-encrypts the workspace's secrets." data-confirm-ok="Rotate">
						<input type="hidden" name="encrypt" value="true" data-type="bool">
						<input type="hidden" name="comment" value="{{ .Comment }}">
						<button class="btn btn-sm" type="submit" title="Generate a new encryption key"><svg class="ic"><use href="#i-rotate"/></svg> Rotate</button>
					</form>
					<div class="key-fp faint">{{ .Fingerprint }}</div>
					{{- else }}
					<span class="badge bad">missing</span>
					<form class="inline" data-api="{{ $api }}/generate" data-method="POST" data-next="/settings/keys?notice=Encryption+key+added+to+{{ $ws.Folder }}#ws-{{ $ws.Folder }}">
						<input type="hidden" name="encrypt" value="true" data-type="bool">
						<input type="hidden" name="comment" value="task-ui {{ $ws.Folder }}">
						<button class="btn btn-sm btn-primary" type="submit">Generate</button>
					</form>
					<a class="btn btn-sm" href="{{ $base }}/keys/upload?encrypt=1">Upload</a>
					<div class="warn">You can't add secrets until it has one.</div>
					{{- end }}
				</td>
				<td>
					{{- if .SSH }}
					<a href="{{ $base }}#keys">{{ len .SSH }} key{{ if ne (len .SSH) 1 }}s{{ end }}</a>
					<div class="key-names">{{ range .SSH }}<a class="key-chip" href="{{ $base }}/keys/{{ .Name }}" title="Grants of {{ .Name }}">{{ .Name }}</a>{{ end }}</div>
					{{- else }}
					<span class="faint">none</span> <a href="{{ $base }}/keys/generate">Generate</a> · <a href="{{ $base }}/keys/upload">Upload</a>
					<div class="faint">This workspace has no SSH key of its own. Its runs get only the instance SSH keys their policies grant.</div>
					{{- end }}
				</td>
			</tr>
			{{- end }}
		</tbody>
	</table>
	</div>
</section>

<section class="block" id="instance-ssh">
	<div class="block-head">
		<h3>Instance SSH keys <small>{{ .KeysDir }}</small></h3>
		<div class="actions">
			<a class="btn btn-sm" href="/settings/keys/upload">Upload key</a>
			<a class="btn btn-sm btn-primary" href="/settings/keys/generate">Generate key</a>
		</div>
	</div>
	<p class="dim">Runs and git use these keys to reach hosts, git remotes and services. A key's policies list the roles and users whose runs get it in their ssh agent. A key made for one user carries their name, such as id_ed25519_alice. A key with no policy stays with the instance: git and the inventory's logins use it, runs don't.</p>
	<div class="table-wrap">
	<table class="table keys">
		<thead><tr><th>Comment</th><th>Fingerprint</th><th>Policies</th><th>Date</th><th>Actions</th></tr></thead>
		<tbody>
		{{- $any := false }}
		{{- range .Keys }}{{ if not .Backup }}{{ $any = true }}
			<tr>
				<td class="key">{{ .Comment }} <span class="badge">{{ .Name }}</span></td>
				<td class="fingerprint"><span class="kind dim">{{ .Type }}</span>{{ .Fingerprint }}</td>
				{{- $roles := index $.KeyPolicies .Name }}{{ $users := index $.KeyUsers .Name }}
				<td class="nowrap">{{ range $roles }}<span class="badge">{{ . }}</span> {{ end }}{{ range $users }}{{ template "avatar" (list $ .) }} {{ end }}{{ if not (or $roles $users) }}<span class="faint">none</span>{{ end }}</td>
				<td class="nowrap" title="{{ stamp .Modified }}">{{ ago .Modified }}</td>
				<td class="actions">
					<button type="button" class="btn-icon" data-copy="{{ .PublicKey }}" title="Copy public key" aria-label="Copy public key"><svg class="ic"><use href="#i-copy"/></svg></button>
					<a class="btn-icon" href="/settings/keys/{{ .Name }}" title="Policies: the roles whose runs get it" aria-label="Edit policies"><svg class="ic"><use href="#i-edit"/></svg></a>
					{{- $kind := .Kind }}
					{{- if $kind }}
					<form data-api="/api/keys/generate" data-method="POST" data-next="reload" data-confirm="Rotate {{ .Name }}? task-ui keeps the current pair as a backup. Hosts that trust it need the new public key." data-confirm-ok="Rotate">
						<input type="hidden" name="type" value="{{ $kind }}">
						<input type="hidden" name="name" value="{{ .Label }}">
						<input type="hidden" name="comment" value="{{ .Comment }}">
						<button class="btn-icon" type="submit" title="Rotate: generate a new pair" aria-label="Rotate key"><svg class="ic"><use href="#i-rotate"/></svg></button>
					</form>
					{{- end }}
					<button type="button" class="btn-icon danger" data-delete="/api/keys/{{ .Name }}" data-confirm="Delete {{ .Name }}? Jobs that use it lose access." data-next="reload" title="Delete key" aria-label="Delete key"><svg class="ic"><use href="#i-trash"/></svg></button>
				</td>
			</tr>
		{{- end }}{{ end }}
		{{- if not $any }}
			<tr class="empty"><td colspan="5">No key pair yet. Generate one, or upload the private key you already use.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>

<section class="block">
	<h3>Who gets which key <small>through the ssh agent of their runs</small></h3>
	<div class="table-wrap">
	<table class="table key-holders">
		<thead><tr><th>Policy or user</th><th>Keys</th></tr></thead>
		<tbody>
		{{- range .KeyHolders }}
			<tr>
				{{- if .User }}
				<td class="nowrap">{{ template "user" (list $ .User.Name) }} <span class="faint">{{ .Role }}</span></td>
				{{- else }}
				<td class="nowrap"><span class="badge">{{ .Role }}</span> <span class="faint">every {{ .Role }}</span></td>
				{{- end }}
				<td>
					{{- $role := .Role }}
					{{- range .Direct }}<a class="key-chip" href="/settings/keys/{{ .Name }}" title="{{ .Name }}, {{ .Fingerprint }}">{{ or .Comment .Name }}</a>{{ end }}
					{{- range .ViaRole }}<a class="key-chip faint" href="/settings/keys/{{ .Name }}" title="{{ .Name }}, from the {{ $role }} role">{{ or .Comment .Name }} <small>via role</small></a>{{ end }}
					{{- if not (or .Direct .ViaRole) }}<span class="faint">none</span>{{ end }}
				</td>
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>

{{- if .HomeKeys }}
<section class="block">
	<h3>Server user's keys <small>{{ .HomeKeysDir }}</small></h3>
	<p class="dim">The keys of the system user task-ui runs as. ssh and git use them by default: in every run, for git remotes and for the inventory's logins. They come with the image; change them there, not here.</p>
	<div class="table-wrap">
	<table class="table keys">
		<thead><tr><th>Comment</th><th>Fingerprint</th><th>Date</th><th>Actions</th></tr></thead>
		<tbody>
		{{- range .HomeKeys }}
			<tr>
				<td class="key">{{ .Comment }} <span class="badge">{{ .Name }}</span></td>
				<td class="fingerprint"><span class="kind dim">{{ .Type }}</span>{{ .Fingerprint }}</td>
				<td class="nowrap" title="{{ stamp .Modified }}">{{ ago .Modified }}</td>
				<td class="actions">
					<button type="button" class="btn-icon" data-copy="{{ .PublicKey }}" title="Copy public key" aria-label="Copy public key"><svg class="ic"><use href="#i-copy"/></svg></button>
				</td>
			</tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>
{{- end }}

{{- range .KeyFolders }}
<div class="alert signal"><b>{{ .Project }}</b> has private keys in <b>{{ .Dir }}</b>: {{ range $i, $k := .Keys }}{{ if $i }}, {{ end }}{{ $k.Name }}{{ end }}. Its runs don't get them. Upload a key the project needs as an instance key, then grant it to roles or users with a policy.</div>
{{- end }}

{{- $backups := false }}{{ range .Keys }}{{ if .Backup }}{{ $backups = true }}{{ end }}{{ end }}
{{- if $backups }}
<section class="block">
	<h3>Backups <small>pairs a rotation or upload set aside</small></h3>
	<div class="table-wrap">
	<table class="table keys">
		<thead><tr><th>Comment</th><th>Fingerprint</th><th>Date</th><th>Actions</th></tr></thead>
		<tbody>
		{{- range .Keys }}{{ if .Backup }}
			<tr>
				<td class="key nowrap dim">{{ .Comment }} <span class="badge">{{ .Name }}</span></td>
				<td class="fingerprint dim"><span class="kind faint">{{ .Type }}</span>{{ .Fingerprint }}</td>
				<td class="nowrap dim" title="{{ stamp .Modified }}">{{ ago .Modified }}</td>
				<td class="actions">
					<button type="button" class="btn-icon" data-copy="{{ .PublicKey }}" title="Copy public key" aria-label="Copy public key"><svg class="ic"><use href="#i-copy"/></svg></button>
					<button type="button" class="btn-icon danger" data-delete="/api/keys/{{ .Name }}" data-confirm="Delete the backup {{ .Name }}?" data-next="reload" title="Delete backup" aria-label="Delete backup"><svg class="ic"><use href="#i-trash"/></svg></button>
				</td>
			</tr>
		{{- end }}{{ end }}
		</tbody>
	</table>
	</div>
</section>
{{- end }}
{{ end }}
