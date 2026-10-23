{{/* The label of a key whose rotation kept a backup that serves beside
     it: BACKUP ACTIVE and the backup's name. */}}
{{ define "backup-active" }}<span class="backup-active" title="The previous material serves beside the key until you delete the backup">BACKUP ACTIVE: {{ . }}</span>{{ end }}

{{/* The head of a keys table, the same on the Keys page and the
     Instance page: Name, Algorithm, Size, Public key with its copy
     button, Purpose and the row's ⋯ menu. */}}
{{ define "keyhead" }}<colgroup><col class="c-name"><col class="c-algo"><col class="c-size"><col class="c-pub"><col class="c-purpose"><col class="c-actions"></colgroup>
		<thead><tr><th>Name</th><th>Algorithm</th><th>Size</th><th>Public key</th><th>Purpose</th><th class="actions" aria-label="Actions"></th></tr></thead>{{ end }}

{{/* A key's row, (list page key next): its kind, its public key to
     copy, and the ⋯ menu, its details on top, then Copy and Download,
     Edit links, Rotate Key or Delete backup, and Remove last. next is
     the page a change returns to. */}}
{{ define "keyrow" }}
{{- $p := index . 0 }}{{ $k := index . 1 }}{{ $next := index . 2 }}
<tr id="key-{{ $k.Name }}">
	<td class="key">{{ if $k.Encrypts }}<span class="mono">{{ $k.Name }}</span>{{ else }}<a class="mono" href="/settings/keys/{{ $k.Name }}">{{ $k.Name }}</a>{{ end }}{{ with $k.Backup }}{{ template "backup-active" . }}{{ end }}{{ with $k.Comment }}<div class="faint">{{ . }}</div>{{ end }}</td>
	<td class="mono" data-label="Algorithm">{{ $k.Algo }}</td>
	<td class="mono nowrap" data-label="Size">{{ $k.Size }} bits</td>
	<td class="pubkey" data-label="Public key"><div class="pub-line"><span class="mono" title="{{ $k.PublicKey }}">{{ $k.ShortPublic }}</span><button type="button" class="btn-icon" data-copy="{{ $k.PublicKey }}" title="Copy public key" aria-label="Copy the public key of {{ $k.Name }}"><svg class="ic"><use href="#i-copy"/></svg></button></div></td>
	<td data-label="Purpose"><span class="badge{{ if $k.Encrypts }} signal{{ end }}"{{ if $k.Git }} title="git pulls and clones with it"{{ end }}>{{ if $k.Encrypts }}encryption{{ else }}git + ssh{{ end }}</span></td>
	<td class="actions">
		<details class="more-menu row-more">
			<summary class="btn-icon" aria-haspopup="menu" title="Actions on {{ $k.Name }}" aria-label="Actions on {{ $k.Name }}"><svg class="ic"><use href="#i-more"/></svg></summary>
			<div class="more-list" role="menu">
				<dl class="more-details">
					<dt>Fingerprint</dt><dd class="mono">{{ $k.Fingerprint }}</dd>
					{{- if not $k.CreatedAt.IsZero }}<dt>Created</dt><dd title="{{ stamp $k.CreatedAt }}">{{ ago $k.CreatedAt }}</dd>{{ end }}
					{{- with $k.RotatedAt }}<dt>Rotated</dt><dd title="{{ stamp . }}">{{ ago . }}</dd>{{ end }}
					{{- with $k.Backup }}<dt>Backup</dt><dd class="mono">{{ . }}<br>{{ $k.BackupFingerprint }}</dd>{{ end }}
					<dt>Scope</dt><dd>{{ with $k.Workspace }}workspace{{ else }}instance{{ end }}</dd>
					{{- if $k.Encrypts }}<dt>Links</dt><dd class="faint">never handed out</dd>
					{{- else }}<dt>Links</dt><dd>{{ range $k.Roles }}<span class="badge">{{ . }}</span> {{ end }}{{ range $k.Users }}{{ . }} {{ end }}{{ if not (or $k.Roles $k.Users) }}<span class="faint">{{ if $k.Workspace }}every run of the workspace{{ else }}none{{ end }}</span>{{ end }}</dd>{{ end }}
					{{- with $p.KeyReads }}{{ with index . $k.Name }}<dt>Reads</dt><dd>{{ . }}</dd>{{ end }}{{ end }}
				</dl>
				<button type="button" class="more-item" role="menuitem" data-copy="{{ $k.PublicKey }}"><svg class="ic"><use href="#i-copy"/></svg> Copy public key</button>
				<button type="button" class="more-item" role="menuitem" data-download="{{ $k.PublicKey }}" data-filename="{{ $k.Name }}.pub"><svg class="ic"><use href="#i-download"/></svg> Download public key</button>
				{{- if and ($p.Can "key/edit") (not $k.Encrypts) }}
				<a class="more-item" role="menuitem" href="/settings/keys/{{ $k.Name }}"><svg class="ic"><use href="#i-edit"/></svg> Edit links</a>
				{{- end }}
				{{- if $p.Can "key/edit" }}
				<hr class="more-sep">
				{{- if $k.Backup }}
				<button class="more-item" type="button" role="menuitem" disabled title="Delete the backup first"><svg class="ic"><use href="#i-rotate"/></svg> Rotate Key <small class="faint">Delete the backup first</small></button>
				<button type="button" class="more-item danger" role="menuitem" data-delete="/api/keys/{{ $k.Name }}/backup" data-confirm="Delete the backup {{ $k.Backup }}, {{ $k.BackupFingerprint }}? {{ if $k.Encrypts }}Secrets the new key has not re-encrypted stop opening.{{ else }}Hosts that know only its public key stop letting runs in.{{ end }}" data-next="reload"><svg class="ic"><use href="#i-trash"/></svg> Delete backup</button>
				{{- else }}
				<form data-api="/api/keys/{{ $k.Name }}/rotate" data-method="POST" data-next="{{ $next }}{{ if $k.Encrypts }}?rekeyed={rekeyed}&what=secret{{ else }}?notice=Key+rotated{{ end }}#key-{{ $k.Name }}" data-confirm="Rotate {{ $k.Name }}, {{ $k.Fingerprint }}? task-ui makes a new {{ $k.Algo }} key under its name and keeps this one as {{ backupName $k.Name }}, which {{ if $k.Encrypts }}still decrypts while task-ui re-encrypts every secret{{ else }}runs and git keep using beside the new key until you delete it: add the new public key to the hosts and remotes first{{ end }}." data-confirm-ok="Rotate Key">
					<button class="more-item" type="submit" role="menuitem"><svg class="ic"><use href="#i-rotate"/></svg> Rotate Key</button>
				</form>
				{{- end }}
				{{- end }}
				{{- if and ($p.Can "key/delete") (not $k.Encrypts) }}
				<button type="button" class="more-item danger" role="menuitem" data-delete="/api/keys/{{ $k.Name }}" data-confirm="Remove {{ $k.Name }}? Runs and hosts that use it lose access. Its files stay in the key folder as a backup." data-next="reload"><svg class="ic"><use href="#i-trash"/></svg> Remove</button>
				{{- end }}
			</div>
		</details>
	</td>
</tr>
{{- end }}
