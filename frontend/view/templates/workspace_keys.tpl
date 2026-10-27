{{ define "content" }}
{{- $ws := .Workspace }}{{ $here := print "/settings/workspaces/" $ws.Folder "/keys" }}{{ $edit := .Can "workspace/edit" }}
<div class="page-head">
	<div>
		<h2>Keys of {{ $ws.Folder }}</h2>
		<p>The keys the workspace's runs get in their ssh agent, each to every run in {{ $ws.Folder }} or to the roles and users its links name. Generate, upload, link and remove them here.</p>
	</div>
	<div class="actions">
		{{- if and $edit .KeysDir }}
		<button type="button" class="btn" data-dialog-open="key-upload-dialog"><svg class="ic"><use href="#i-upload"/></svg>Upload key</button>
		<button type="button" class="btn btn-primary" data-dialog-open="key-generate-dialog"><svg class="ic"><use href="#i-key"/></svg>Generate key</button>
		{{- end }}
		{{ template "menu" (menu (print "Actions on " $ws.Folder)
			(item "folder-open" "Open" "href" ($.Href $ws.Folder))
			(item "settings" "Edit workspace" "href" (print "/settings/workspaces/" $ws.Folder))
			(and (.Can "secret/view") (item "secret" "Secrets" "href" (workspaceSecretsHref $ws.Folder)))
		) }}
	</div>
</div>

{{- if not .KeysDir }}
<div class="alert">The instance has no key folder, so it keeps no SSH keys.</div>
{{- else }}
<section class="block workspace-keys" id="keys">
	{{- template "blockhead" (head "Keys" (print (len .Keys) " key" (or (and (ne (len .Keys) 1) "s") "")) (and (not .Keys) "No keys added.")) }}
	{{- with .Keys }}
	<div class="table-wrap fit">
	<table class="table keys key-table cards">
		{{ template "keyhead" }}
		<tbody>
		{{- range . }}{{ template "keyrow" (list $ . $here $ws.Folder true) }}{{ end }}
		</tbody>
	</table>
	</div>
	{{- end }}
</section>
{{- if $edit }}
{{ template "wskeydialogs" (list . $here) }}
{{- range .Keys }}{{ if not .Encrypts }}
<dialog class="confirm" id="key-links-{{ .Name }}" aria-labelledby="key-links-{{ .Name }}-title">
	{{ template "wskeylinks" (list $ . $here true) }}
</dialog>
{{- end }}{{ end }}
{{- end }}
{{- end }}
{{ end }}
