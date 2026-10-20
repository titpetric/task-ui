{{ define "content" }}
<div class="page-head">
	<div>
		<h2>SSH keys</h2>
		<p>The key pairs jobs use to reach other hosts, git remotes and services, and what the secrets are encrypted with{{ if .Encryption }}: <b>{{ join .Encryption ", " }}</b>{{ end }}. Add the public key where it should be accepted. Rotating a key writes the secrets again for the new one; the last key that opens them cannot be deleted. A key's policies are the roles and the users whose runs get it in their ssh agent; a key made for someone carries their name, id_ed25519_alice.</p>
	</div>
	<div class="actions">
		<a class="btn" href="/settings/keys/upload">Upload key</a>
		<a class="btn btn-primary" href="/settings/keys/generate">Generate key</a>
	</div>
</div>

<section class="block">
	<h3>Keys <small>{{ .KeysDir }}</small></h3>
	<div class="table-wrap">
	<table class="table keys">
		<thead><tr><th>Comment</th><th>Fingerprint</th><th>Policies</th><th>Date</th><th>Actions</th></tr></thead>
		<tbody>
		{{- $any := false }}
		{{- range .Keys }}{{ if not .Backup }}{{ $any = true }}
			<tr>
				{{- $name := .Name }}
				<td class="key">{{ .Comment }} <span class="badge">{{ .Name }}</span>{{ range $.Encryption }}{{ if eq . $name }} <span class="badge signal" title="The secrets are encrypted with this key"><svg class="ic"><use href="#i-secret"/></svg> secrets</span>{{ end }}{{ end }}</td>
				<td class="fingerprint"><span class="kind dim">{{ .Type }}</span>{{ .Fingerprint }}</td>
				{{- $roles := index $.KeyPolicies .Name }}{{ $users := index $.KeyUsers .Name }}
				<td class="nowrap">{{ range $roles }}<span class="badge">{{ . }}</span> {{ end }}{{ range $users }}{{ template "avatar" (list $ .) }} {{ end }}{{ if not (or $roles $users) }}<span class="faint">none</span>{{ end }}</td>
				<td class="nowrap" title="{{ stamp .Modified }}">{{ ago .Modified }}</td>
				<td class="actions">
					<button type="button" class="btn-icon" data-copy="{{ .PublicKey }}" title="Copy public key" aria-label="Copy public key"><svg class="ic"><use href="#i-copy"/></svg></button>
					<a class="btn-icon" href="/settings/keys/{{ .Name }}" title="Policies: the roles whose runs get it" aria-label="Edit policies"><svg class="ic"><use href="#i-edit"/></svg></a>
					{{- $kind := .Kind }}
					{{- if $kind }}
					<form data-api="/api/keys/generate" data-method="POST" data-next="reload" data-confirm="Rotate {{ .Name }}? The current pair is kept as a backup. Hosts that trust it need the new public key." data-confirm-ok="Rotate">
						<input type="hidden" name="type" value="{{ $kind }}">
						<input type="hidden" name="name" value="{{ .Label }}">
						<input type="hidden" name="comment" value="{{ .Comment }}">
						<button class="btn-icon" type="submit" title="Rotate: generate a new pair" aria-label="Rotate key"><svg class="ic"><use href="#i-rotate"/></svg></button>
					</form>
					{{- end }}
					<button type="button" class="btn-icon danger" data-delete="/api/keys/{{ .Name }}" data-confirm="Delete {{ .Name }}? Jobs using it lose access." data-next="reload" title="Delete key" aria-label="Delete key"><svg class="ic"><use href="#i-trash"/></svg></button>
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
				<td class="nowrap">{{ template "avatar" (list $ .User.Name) }} {{ .User.Name }} <span class="faint">{{ .Role }}</span></td>
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
	<p class="dim">The keys of the user task-ui runs as. ssh and git take them by default: in every run, for git remotes, and for the inventory's logins. They come with the image, so they change there, not here.</p>
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
<div class="alert signal"><b>{{ .Project }}</b> has private keys in <b>{{ .Dir }}</b>: {{ range $i, $k := .Keys }}{{ if $i }}, {{ end }}{{ $k.Name }}{{ end }}. They are not loaded into its runs. Upload a key the project needs as an instance key and give it to the roles or users by policy.</div>
{{- end }}

{{- $backups := false }}{{ range .Keys }}{{ if .Backup }}{{ $backups = true }}{{ end }}{{ end }}
{{- if $backups }}
<section class="block">
	<h3>Backups <small>pairs moved aside by a rotation or upload</small></h3>
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
