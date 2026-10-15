{{ define "content" }}
<div class="crumbs"><a href="/">Overview</a><span>SSH keys</span></div>
<div class="page-head">
	<div>
		<h2>SSH keys</h2>
		<p>The key pairs runbooks use to reach other hosts, git remotes and services. Add the public key where it should be accepted.</p>
	</div>
	<div class="dim">{{ .KeysDir }}</div>
</div>

{{- if .Notice }}<div class="alert ok" role="status">{{ .Notice }}</div>{{ end }}
{{- if .Error }}<div class="alert bad" role="alert">{{ .Error }}</div>{{ end }}

<section class="block">
	<h3>Instance keys</h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Name</th><th>Type</th><th>Fingerprint</th><th class="hide-sm">Comment</th><th class="hide-sm">Modified</th><th></th></tr></thead>
		<tbody>
		{{- $any := false }}
		{{- range .Keys }}{{ if not .Backup }}{{ $any = true }}
			<tr>
				<td class="key">{{ .Name }}</td>
				<td class="nowrap">{{ .Type }}{{ if .Bits }} <span class="faint">{{ .Bits }}</span>{{ end }}</td>
				<td class="mono-break">{{ .Fingerprint }}</td>
				<td class="dim nowrap hide-sm">{{ .Comment }}</td>
				<td class="nowrap hide-sm" title="{{ stamp .Modified }}">{{ ago .Modified }}</td>
				<td class="actions">
					<button type="button" class="btn btn-sm" data-copy="{{ .PublicKey }}">Copy public key</button>
					{{- $kind := "" }}{{ if eq .Name "id_ed25519" }}{{ $kind = "ed25519" }}{{ else if eq .Name "id_ecdsa" }}{{ $kind = "ecdsa" }}{{ else if eq .Name "id_rsa" }}{{ $kind = "rsa" }}{{ end }}
					{{- if $kind }}
					<form method="post" action="/admin/keys/generate" data-confirm="Rotate {{ .Name }}? The current pair is kept as a backup, hosts that trust it need the new public key.">
						<input type="hidden" name="type" value="{{ $kind }}">
						<input type="hidden" name="comment" value="{{ .Comment }}">
						<button class="btn btn-sm" type="submit">Rotate</button>
					</form>
					{{- end }}
					<form method="post" action="/admin/keys/delete" data-confirm="Delete {{ .Name }}? Runbooks using it lose access.">
						<input type="hidden" name="name" value="{{ .Name }}">
						<button class="btn btn-danger btn-sm" type="submit">Delete</button>
					</form>
				</td>
			</tr>
		{{- end }}{{ end }}
		{{- if not $any }}
			<tr class="empty"><td colspan="6">No key pair yet. Generate one below, or upload the private key you already use.</td></tr>
		{{- end }}
		</tbody>
	</table>
	</div>
</section>

<div class="grid-2">
	<section class="panel">
		<div class="section-title">Generate</div>
		<p>Creates a new pair. An existing pair of the same type is moved aside as a backup.</p>
		<form method="post" action="/admin/keys/generate">
			<div class="field">
				<label for="gen-comment">Comment</label>
				<input type="text" id="gen-comment" name="comment" value="task-ui">
			</div>
			<div class="btn-group">
				<button class="btn btn-primary" type="submit" name="type" value="ed25519">Generate ed25519</button>
				<button class="btn" type="submit" name="type" value="ecdsa">ECDSA P-256</button>
				<button class="btn" type="submit" name="type" value="rsa">RSA 4096</button>
			</div>
		</form>
	</section>

	<section class="panel">
		<div class="section-title">Upload</div>
		<p>Bring your own private key, without a passphrase. The public key is derived from it.</p>
		<form method="post" action="/admin/keys/upload" enctype="multipart/form-data">
			<div class="field">
				<label for="up-key">Private key</label>
				<textarea id="up-key" name="private_key" spellcheck="false" autocomplete="off"></textarea>
			</div>
			<div class="field">
				<label for="up-file">Or a key file</label>
				<input type="file" id="up-file" name="key_file">
			</div>
			<div class="field">
				<label for="up-comment">Comment</label>
				<input type="text" id="up-comment" name="comment" value="task-ui">
			</div>
			<button class="btn btn-primary" type="submit">Upload key</button>
		</form>
	</section>
</div>

{{- $backups := false }}{{ range .Keys }}{{ if .Backup }}{{ $backups = true }}{{ end }}{{ end }}
{{- if $backups }}
<section class="block" style="margin-top:1.6rem">
	<h3>Backups <small>pairs moved aside by a rotation or upload</small></h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>Name</th><th>Type</th><th>Fingerprint</th><th class="hide-sm">Comment</th><th></th></tr></thead>
		<tbody>
		{{- range .Keys }}{{ if .Backup }}
			<tr>
				<td class="key dim">{{ .Name }}</td>
				<td>{{ .Type }}</td>
				<td class="mono-break dim">{{ .Fingerprint }}</td>
				<td class="dim hide-sm">{{ .Comment }}</td>
				<td class="actions">
					<button type="button" class="btn btn-sm" data-copy="{{ .PublicKey }}">Copy public key</button>
					<form method="post" action="/admin/keys/delete" data-confirm="Delete the backup {{ .Name }}?">
						<input type="hidden" name="name" value="{{ .Name }}">
						<button class="btn btn-danger btn-sm" type="submit">Delete</button>
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
