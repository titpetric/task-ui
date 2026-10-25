{{ define "content" }}
{{- $ws := .Workspace }}{{ $here := print "/settings/workspaces/" $ws.Folder "/secrets" }}
{{- $keyed := false }}{{ range .SecretGroups }}{{ if .Key }}{{ $keyed = true }}{{ end }}{{ end }}
<div class="page-head">
	<div>
		<h2>Secrets of {{ $ws.Folder }}</h2>
		<p>Variables for the environment of the workspace's runs, for every project in {{ $ws.Folder }} or for one; a project secret replaces a workspace one of the same name. Add, replace and delete them here; a value never shows again once saved.</p>
	</div>
	<div class="actions">
		{{- if .Can "secret/edit" }}
		{{- if $keyed }}
		<button type="button" class="btn btn-primary" data-dialog-open="secret-add-dialog"><svg class="ic"><use href="#i-plus"/></svg>Add secret</button>
		{{- else }}
		<button class="btn btn-primary" type="button" disabled title="The instance has no encryption key yet"><svg class="ic"><use href="#i-plus"/></svg>Add secret</button>
		{{- end }}
		{{- end }}
		{{ template "menu" (menu (print "Actions on " $ws.Folder)
			(item "folder-open" "Open" "href" ($.Href $ws.Folder))
			(and (.Can "workspace/view") (item "settings" "Edit workspace" "href" (print "/settings/workspaces/" $ws.Folder)))
			(and (.Can "workspace/view") (item "key" "SSH keys" "href" (print "/settings/workspaces/" $ws.Folder "/keys")))
			(item "secret" "Every workspace's secrets" "href" "/settings/secrets")
		) }}
	</div>
</div>

{{- range .SecretGroups }}{{ template "secretgroup" (list $ . true) }}{{ end }}
{{- if and (.Can "secret/edit") $keyed }}
{{ template "secretdialogs" (list . $here) }}
{{- end }}
{{ end }}
