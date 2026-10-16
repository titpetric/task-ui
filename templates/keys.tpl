{{ define "content" }}
{{ template "crumbs" (list (crumb "Settings" "/settings") (crumb "SSH keys" "")) }}
<div class="page-head">
	<div>
		<h2>SSH keys</h2>
		<p>The key pairs jobs use to reach other hosts, git remotes and services. Add the public key where it should be accepted.</p>
	</div>
	<div class="actions">
		<a class="btn" href="/settings/keys/upload">Upload key</a>
		<a class="btn btn-primary" href="/settings/keys/generate">Generate key</a>
	</div>
</div>

{{- if .Notice }}<div class="alert ok" role="status">{{ .Notice }}</div>{{ end }}
{{- if .Error }}<div class="alert bad" role="alert">{{ .Error }}</div>{{ end }}

<section class="block">
	<h3>Keys <small>{{ .KeysDir }}</small></h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Comment</th><th>Fingerprint</th><th>Date</th><th>Actions</th></tr></thead>
		<tbody>
		{{- $any := false }}
		{{- range .Keys }}{{ if not .Backup }}{{ $any = true }}
			<tr>
				<td class="key nowrap">{{ .Comment }} <span class="badge">{{ .Name }}</span></td>
				<td class="mono-break"><span class="dim">{{ .Type }}</span> {{ .Fingerprint }}</td>
				<td class="nowrap" title="{{ ago .Modified }}">{{ stamp .Modified }}</td>
				<td class="actions">
					<button type="button" class="btn-icon" data-copy="{{ .PublicKey }}" title="Copy public key" aria-label="Copy public key"><svg class="ic"><use href="#i-copy"/></svg></button>
					{{- $kind := "" }}{{ if eq .Name "id_ed25519" }}{{ $kind = "ed25519" }}{{ else if eq .Name "id_ecdsa" }}{{ $kind = "ecdsa" }}{{ else if eq .Name "id_rsa" }}{{ $kind = "rsa" }}{{ end }}
					{{- if $kind }}
					<form method="post" action="/settings/keys/generate" data-confirm="Rotate {{ .Name }}? The current pair is kept as a backup. Hosts that trust it need the new public key." data-confirm-ok="Rotate">
						<input type="hidden" name="type" value="{{ $kind }}">
						<input type="hidden" name="comment" value="{{ .Comment }}">
						<button class="btn-icon" type="submit" title="Rotate: generate a new pair" aria-label="Rotate key"><svg class="ic"><use href="#i-rotate"/></svg></button>
					</form>
					{{- end }}
					<form method="post" action="/settings/keys/delete" data-confirm="Delete {{ .Name }}? Jobs using it lose access." data-confirm-ok="Delete">
						<input type="hidden" name="name" value="{{ .Name }}">
						<button class="btn-icon danger" type="submit" title="Delete key" aria-label="Delete key"><svg class="ic"><use href="#i-trash"/></svg></button>
					</form>
				</td>
			</tr>
		{{- end }}{{ end }}
		{{- if not $any }}
			<tr class="empty"><td colspan="4">No key pair yet. Generate one, or upload the private key you already use.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>

{{- $backups := false }}{{ range .Keys }}{{ if .Backup }}{{ $backups = true }}{{ end }}{{ end }}
{{- if $backups }}
<section class="block">
	<h3>Backups <small>pairs moved aside by a rotation or upload</small></h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Comment</th><th>Fingerprint</th><th>Date</th><th>Actions</th></tr></thead>
		<tbody>
		{{- range .Keys }}{{ if .Backup }}
			<tr>
				<td class="key nowrap dim">{{ .Comment }} <span class="badge">{{ .Name }}</span></td>
				<td class="mono-break dim"><span class="faint">{{ .Type }}</span> {{ .Fingerprint }}</td>
				<td class="nowrap dim">{{ stamp .Modified }}</td>
				<td class="actions">
					<button type="button" class="btn-icon" data-copy="{{ .PublicKey }}" title="Copy public key" aria-label="Copy public key"><svg class="ic"><use href="#i-copy"/></svg></button>
					<form method="post" action="/settings/keys/delete" data-confirm="Delete the backup {{ .Name }}?" data-confirm-ok="Delete">
						<input type="hidden" name="name" value="{{ .Name }}">
						<button class="btn-icon danger" type="submit" title="Delete backup" aria-label="Delete backup"><svg class="ic"><use href="#i-trash"/></svg></button>
					</form>
				</td>
			</tr>
		{{- end }}{{ end }}
		</tbody>
	</table>
	</div>
</section>
{{- end }}
{{ end }}
