{{/* The label of a key whose rotation kept a backup that serves beside
     it: BACKUP ACTIVE and the backup's name. */}}
{{ define "backup-active" }}<span class="backup-active" title="The previous material serves beside the key until you delete the backup">BACKUP ACTIVE: {{ . }}</span>{{ end }}

{{/* The head of a keys table, the same wherever keys list: the Keys
     page, the Instance page, a workspace's page, Your SSH keys and a
     profile: Name, Algorithm, Size, Public key with its copy button,
     Purpose and the row's ⋯ menu. */}}
{{ define "keyhead" }}<colgroup><col class="c-name"><col class="c-algo"><col class="c-size"><col class="c-pub"><col class="c-purpose"><col class="c-actions"></colgroup>
		<thead><tr><th>Name</th><th>Algorithm</th><th>Size</th><th>Public key</th><th>Purpose</th><th class="actions" aria-label="Actions"></th></tr></thead>{{ end }}

{{/* A key's row, (list page key next [workspace]): its kind, its public
     key to copy, and the [...] menu, its details on top, then Copy and
     Download, Edit links, Rotate Key or Delete backup, and Remove last.
     next is the page a change returns to. On a workspace's page, a key of
     the workspace links and removes through the workspace, which takes
     workspace/edit. */}}
{{ define "keyrow" }}
{{- $p := index . 0 }}{{ $k := index . 1 }}{{ $next := index . 2 }}{{ $ws := "" }}{{ if gt (len .) 3 }}{{ $ws = index . 3 }}{{ end }}
{{- $edit := print "/settings/keys/" $k.Name }}{{ $remove := print "/api/keys/" $k.Name }}{{ $canEdit := $p.Can "key/edit" }}{{ $canRemove := $p.Can "key/delete" }}
{{- if $ws }}{{ $edit = print "/settings/workspaces/" $ws "/keys/" $k.Name }}{{ $remove = print "/api/workspaces/" $ws "/keys/" $k.Name }}{{ $canRemove = $p.Can "workspace/edit" }}{{ end }}
{{- $links := join $k.Roles ", " }}{{ with $k.Users }}{{ if $links }}{{ $links = print $links ", " }}{{ end }}{{ $links = print $links (join . ", ") }}{{ end }}
{{- if $k.Encrypts }}{{ $links = "never handed out" }}{{ else if not $links }}{{ if $k.Workspace }}{{ $links = "every run of the workspace" }}{{ else }}{{ $links = "none" }}{{ end }}{{ end }}
<tr id="key-{{ $k.Name }}">
	<td class="key">{{ if or $k.Encrypts (not $canEdit) }}<span class="mono">{{ $k.Name }}</span>{{ else }}<a class="mono" href="{{ $edit }}">{{ $k.Name }}</a>{{ end }}{{ with $k.Backup }}{{ template "backup-active" . }}{{ end }}{{ with $k.Comment }}<div class="faint">{{ . }}</div>{{ end }}</td>
	<td class="mono" data-label="Algorithm">{{ $k.Algo }}</td>
	<td class="mono nowrap" data-label="Size">{{ $k.Size }} bits</td>
	<td class="pubkey" data-label="Public key">{{ template "copyable" (list $k.PublicKey (print "the public key of " $k.Name) $k.ShortPublic) }}</td>
	<td data-label="Purpose"><span class="badge{{ if $k.Encrypts }} signal{{ end }}"{{ if $k.Git }} title="git pulls and clones with it"{{ end }}>{{ if $k.Encrypts }}encryption{{ else }}git + ssh{{ end }}</span></td>
	<td class="actions">{{ template "menu" (menu (print "Actions on " $k.Name)
		(detail "Fingerprint" $k.Fingerprint "mono" "")
		(and (not $k.CreatedAt.IsZero) (detail "Created" (ago $k.CreatedAt) "title" (stamp $k.CreatedAt)))
		(and $k.RotatedAt (detail "Rotated" (ago $k.RotatedAt) "title" (stamp $k.RotatedAt)))
		(and $k.Backup (detail "Backup" (print $k.Backup " " $k.BackupFingerprint) "mono" ""))
		(detail "Scope" (or (and $k.Workspace (print "workspace " $k.Workspace)) "instance"))
		(and (or ($p.Can "key/view") $ws) (detail "Links" $links))
		(and $p.KeyReads (index $p.KeyReads $k.Name) (detail "Reads" (print (index $p.KeyReads $k.Name))))
		(item "copy" "Copy public key" "data-copy" $k.PublicKey)
		(item "download" "Download public key" "data-download" $k.PublicKey "data-filename" (print $k.Name ".pub"))
		(and $canEdit (not $k.Encrypts) (item "edit" "Edit links" "href" $edit))
		sep
		(and $canEdit $k.Backup (item "rotate" "Rotate Key" "disabled" "Delete the backup first"))
		(and $canEdit $k.Backup (danger "trash" "Delete backup" "data-delete" (print "/api/keys/" $k.Name "/backup") "data-confirm" (print "Delete the backup " $k.Backup ", " $k.BackupFingerprint "? " (or (and $k.Encrypts "Secrets the new key has not re-encrypted stop opening.") "Hosts that know only its public key stop letting runs in.")) "data-confirm-ok" "Delete backup" "data-next" "reload"))
		(and $canEdit (not $k.Backup) (item "rotate" "Rotate Key" "data-api" (print "/api/keys/" $k.Name "/rotate") "data-method" "POST" "data-next" (print $next (or (and $k.Encrypts "?rekeyed={rekeyed}&what=secret") "?notice=Key+rotated") "#key-" $k.Name) "data-confirm" (print "Rotate " $k.Name ", " $k.Fingerprint "? task-ui makes a new " $k.Algo " key under its name and keeps this one as " (backupName $k.Name) ", which " (or (and $k.Encrypts "still decrypts while task-ui re-encrypts every secret") "runs and git keep using beside the new key until you delete it: add the new public key to the hosts and remotes first") ".") "data-confirm-ok" "Rotate Key"))
		(and $canRemove (not $k.Encrypts) (danger "trash" "Remove" "data-delete" $remove "data-confirm" (print "Remove " $k.Name "? " (or (and $ws "The workspace's runs lose it.") "Runs and hosts that use it lose access.") " Its files stay in the key folder as a backup.") "data-confirm-ok" "Remove" "data-next" "reload"))
	) }}</td>
</tr>
{{- end }}
