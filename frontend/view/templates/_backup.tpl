{{/* The label of a key whose rotation kept a backup that serves beside
     it: BACKUP ACTIVE and the backup's name. */}}
{{ define "backup-active" }}<span class="backup-active" title="The previous material serves beside the key until you delete the backup">BACKUP ACTIVE: {{ . }}</span>{{ end }}
