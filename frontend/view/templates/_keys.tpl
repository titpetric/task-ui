{{/* The label of a key whose rotation kept a backup that serves beside
     it: BACKUP ACTIVE and the backup's name. */}}
{{ define "backup-active" }}<span class="backup-active" title="The previous material serves beside the key until you delete the backup">BACKUP ACTIVE: {{ . }}</span>{{ end }}

{{/* The head of a keys table, the same wherever keys list: the Keys
     page, the Instance page, a workspace's page, Your SSH keys and a
     profile: Name, as wide as the names, Algorithm, Size, Public key,
     which takes the rest of the width with its copy button, and the
     row's ⋯ menu. The table sits in a table-wrap fit; a narrow one
     leaves out Algorithm and Size, hide-narrow. */}}
{{ define "keyhead" }}<colgroup><col class="c-name"><col class="c-algo"><col class="c-size"><col class="c-pub"><col class="c-actions"></colgroup>
		<thead><tr><th>Name</th><th class="hide-narrow">Algorithm</th><th class="hide-narrow">Size</th><th>Public key</th><th class="actions" aria-label="Actions"></th></tr></thead>{{ end }}

{{/* A key's row, (list page key next [workspace [inline]]): its name,
     the encryption key marked, its public key, as much as fits with the
     whole line in its tooltip, to copy, and the [...] menu, its details
     on top, the purpose among them, then
     Copy and Download, Edit links, Rotate Key or Delete backup, and
     Remove last. next is the page a change returns to, Edit links
     too. On a workspace's
     page, a key of the workspace links and removes through the
     workspace, which takes workspace/edit. inline opens Edit links in
     the page's dialog key-links-<name>, which wskeylinks fills. */}}
{{ define "keyrow" }}
{{- $p := index . 0 }}{{ $k := index . 1 }}{{ $next := index . 2 }}{{ $ws := "" }}{{ if gt (len .) 3 }}{{ $ws = index . 3 }}{{ end }}
{{- $edit := print "/settings/keys/" $k.Name }}{{ if ne $next "/settings/keys" }}{{ $edit = print $edit "?next=" (urlquery $next) }}{{ end }}{{ $remove := print "/api/keys/" $k.Name }}{{ $canEdit := $p.Can "key/edit" }}{{ $canRemove := $p.Can "key/delete" }}
{{- if $ws }}{{ $edit = print "/settings/workspaces/" $ws "/keys/" $k.Name }}{{ $remove = print "/api/workspaces/" $ws "/keys/" $k.Name }}{{ $canRemove = $p.Can "workspace/edit" }}{{ end }}
{{- $inline := false }}{{ if gt (len .) 4 }}{{ $inline = index . 4 }}{{ end }}
{{- $links := join $k.Roles ", " }}{{ with $k.Users }}{{ if $links }}{{ $links = print $links ", " }}{{ end }}{{ $links = print $links (join . ", ") }}{{ end }}
{{- if $k.Encrypts }}{{ $links = "never handed out" }}{{ else if not $links }}{{ if $k.Workspace }}{{ $links = "every run of the workspace" }}{{ else }}{{ $links = "none" }}{{ end }}{{ end }}
<tr id="key-{{ $k.Name }}">
	<td class="key">{{ if or $k.Encrypts (not $canEdit) $inline }}<span class="mono">{{ $k.Name }}</span>{{ else }}<a class="mono" href="{{ $edit }}">{{ $k.Name }}</a>{{ end }}{{ if $k.Encrypts }} <span class="badge signal">encryption</span>{{ end }}{{ with $k.Backup }}{{ template "backup-active" . }}{{ end }}{{ with $k.Comment }}<div class="faint">{{ . }}</div>{{ end }}</td>
	<td class="mono hide-narrow" data-label="Algorithm">{{ $k.Algo }}</td>
	<td class="mono nowrap hide-narrow" data-label="Size">{{ $k.Size }} bits</td>
	<td class="pubkey" data-label="Public key">{{ template "copyable" (list $k.PublicKey (print "the public key of " $k.Name) $k.PublicKey true) }}</td>
	<td class="actions">{{ template "menu" (menu (print "Actions on " $k.Name)
		(detail "Fingerprint" $k.Fingerprint "mono" "")
		(and (not $k.CreatedAt.IsZero) (detail "Created" (ago $k.CreatedAt) "title" (stamp $k.CreatedAt)))
		(and $k.RotatedAt (detail "Rotated" (ago $k.RotatedAt) "title" (stamp $k.RotatedAt)))
		(and $k.Backup (detail "Backup" (print $k.Backup " " $k.BackupFingerprint) "mono" ""))
		(detail "Scope" (or (and $k.Workspace (print "workspace " $k.Workspace)) "instance"))
		(detail "Purpose" (or (and $k.Encrypts "encryption") "git + ssh") "title" (or (and $k.Git "git pulls and clones with it") ""))
		(and (or ($p.Can "key/view") $ws) (detail "Links" $links))
		(item "copy" "Copy public key" "data-copy" $k.PublicKey)
		(item "download" "Download public key" "data-download" $k.PublicKey "data-filename" (print $k.Name ".pub"))
		(and $canEdit (not $k.Encrypts) (not $inline) (item "edit" "Edit links" "href" $edit))
		(and $canRemove (not $k.Encrypts) $inline (item "edit" "Edit links" "data-dialog-open" (print "key-links-" $k.Name)))
		sep
		(and $canEdit $k.Backup (item "rotate" "Rotate Key" "disabled" "Delete the backup first"))
		(and $canEdit $k.Backup (danger "trash" "Delete backup" "data-delete" (print "/api/keys/" $k.Name "/backup") "data-confirm" (print "Delete the backup " $k.Backup ", " $k.BackupFingerprint "? " (or (and $k.Encrypts "Secrets the new key has not re-encrypted stop opening.") "Hosts that know only its public key stop letting runs in.")) "data-confirm-ok" "Delete backup" "data-next" "reload"))
		(and $canEdit (not $k.Backup) (item "rotate" "Rotate Key" "data-api" (print "/api/keys/" $k.Name "/rotate") "data-method" "POST" "data-next" (print $next (or (and $k.Encrypts "?rekeyed={rekeyed}&what=secret") "?notice=Key+rotated") "#key-" $k.Name) "data-confirm" (print "Rotate " $k.Name ", " $k.Fingerprint "? task-ui makes a new " $k.Algo " key under its name and keeps this one as " (backupName $k.Name) ", which " (or (and $k.Encrypts "still decrypts while task-ui re-encrypts every secret") "runs and git keep using beside the new key until you delete it: add the new public key to the hosts and remotes first") ".") "data-confirm-ok" "Rotate Key"))
		(and $canRemove (not $k.Encrypts) (danger "trash" "Remove" "data-delete" $remove "data-confirm" (print "Remove " $k.Name "? " (or (and $ws "The workspace's runs lose it.") "Runs and hosts that use it lose access.") " Its files stay in the key folder as a backup.") "data-confirm-ok" "Remove" "data-next" "reload"))
	) }}</td>
</tr>
{{- end }}

{{/* wskeyfoot is the foot of a workspace key's form, (list next dialog
     label): Cancel closes the dialog the form is in, or goes back to
     next, beside the submit button. */}}
{{ define "wskeyfoot" }}
{{- $next := index . 0 }}{{ $dialog := index . 1 }}{{ $label := index . 2 }}
	<div class="form-foot">
		{{- if $dialog }}
		<button type="button" class="btn" data-dialog-close>Cancel</button>
		{{- else }}
		<a class="btn" href="{{ $next }}">Cancel</a>
		{{- end }}
		<button class="btn btn-primary" type="submit">{{ $label }}</button>
	</div>
{{- end }}

{{/* wskeygenerate is the form that generates an SSH key of a workspace,
     (list page next dialog): its name, algorithm, comment and whose runs
     get it; it returns to next. In a dialog, Cancel closes it. */}}
{{ define "wskeygenerate" }}
{{- $p := index . 0 }}{{ $next := index . 1 }}{{ $dialog := index . 2 }}{{ $ws := $p.Workspace }}
<form class="form" data-api="/api/workspaces/{{ $ws.Folder }}/keys/generate" data-method="POST" data-next="{{ $next }}?notice=Key+generated">
	{{- if $dialog }}
	<h3 id="key-generate-title">Generate key</h3>
	<p>Creates an SSH key of {{ $ws.Folder }} in {{ $p.KeysDir }}, named <span class="mono">workspace_{{ $ws.Folder }}_&lt;name&gt;_&lt;algorithm&gt;</span>.</p>
	{{- end }}
	<div class="field">
		<label for="wkg-name">Name</label>
		<input type="text" id="wkg-name" name="name" value="" autocomplete="off" placeholder="deploy">
		<small class="faint">Lower case letters, digits, dashes and underscores; empty is ssh.</small>
	</div>
	<div class="field">
		<label for="wkg-algo">Algorithm</label>
		<select id="wkg-algo" name="algo" data-key-algo>
			<option value="ed25519" selected>ed25519</option>
			<option value="ecdsa" data-bits="256">ECDSA P-256</option>
			<option value="ecdsa" data-bits="384">ECDSA P-384</option>
			<option value="ecdsa" data-bits="521">ECDSA P-521</option>
			<option value="rsa" data-bits="2048">RSA 2048</option>
			<option value="rsa" data-bits="3072">RSA 3072</option>
			<option value="rsa" data-bits="4096">RSA 4096</option>
		</select>
		<input type="hidden" name="bits" value="0" data-type="int">
	</div>
	<div class="field">
		<label for="wkg-comment">Comment</label>
		<input type="text" id="wkg-comment" name="comment" value="task-ui {{ $ws.Folder }}">
	</div>
	{{- template "keygrants" (list (list) nil $p.Users) }}
	{{- template "wskeyfoot" (list $next $dialog "Generate") }}
</form>
{{- end }}

{{/* wskeyupload is the form that uploads a private key as an SSH key
     of a workspace, (list page next dialog), laid out as wskeygenerate. */}}
{{ define "wskeyupload" }}
{{- $p := index . 0 }}{{ $next := index . 1 }}{{ $dialog := index . 2 }}{{ $ws := $p.Workspace }}
<form class="form" data-api="/api/workspaces/{{ $ws.Folder }}/keys/upload" data-method="POST" data-next="{{ $next }}?notice=Key+uploaded">
	{{- if $dialog }}
	<h3 id="key-upload-title">Upload key</h3>
	<p>Saves a private key you already use as an SSH key of {{ $ws.Folder }}. The folder refuses a key it holds, under any name.</p>
	{{- end }}
	<div class="field">
		<label for="wku-file">Private key file</label>
		<input type="file" id="wku-file" data-read-into="wku-content">
	</div>
	<div class="field">
		<label for="wku-content">Or paste the private key</label>
		<textarea id="wku-content" name="content" rows="{{ if $dialog }}6{{ else }}10{{ end }}" spellcheck="false" autocomplete="off" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----" required></textarea>
	</div>
	<div class="field">
		<label for="wku-name">Name</label>
		<input type="text" id="wku-name" name="name" value="" autocomplete="off" placeholder="deploy">
		<small class="faint">Lower case letters, digits, dashes and underscores; empty is ssh.</small>
	</div>
	<div class="field">
		<label for="wku-comment">Comment</label>
		<input type="text" id="wku-comment" name="comment" value="task-ui {{ $ws.Folder }}">
	</div>
	{{- template "keygrants" (list (list) nil $p.Users) }}
	{{- template "wskeyfoot" (list $next $dialog "Upload") }}
</form>
{{- end }}

{{/* wskeylinks is the form that sets whose runs get a workspace's SSH
     key, (list page key next dialog). */}}
{{ define "wskeylinks" }}
{{- $p := index . 0 }}{{ $k := index . 1 }}{{ $next := index . 2 }}{{ $dialog := index . 3 }}{{ $ws := $p.Workspace }}
<form class="form" data-api="/api/workspaces/{{ $ws.Folder }}/keys/{{ $k.Name }}" data-method="PUT" data-next="{{ $next }}?notice=Links+saved">
	{{- if $dialog }}
	<h3 id="key-links-{{ $k.Name }}-title">Edit links of {{ $k.Name }}</h3>
	<p>With nothing checked, every run in {{ $ws.Folder }} gets it.</p>
	{{- end }}
	{{- template "keygrants" (list (index $p.KeyPolicies $k.Name) (index $p.KeyUsers $k.Name) $p.Users) }}
	{{- template "wskeyfoot" (list $next $dialog "Save") }}
</form>
{{- end }}
