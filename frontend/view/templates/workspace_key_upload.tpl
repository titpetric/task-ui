{{ define "content" }}
{{- $ws := .Workspace }}
<div class="page-head">
	<div>
		<h2>Upload key</h2>
		<p>Saves a private key you already use as an SSH key of the workspace {{ $ws.Folder }}. Its runs get it in their ssh agent. task-ui reads its algorithm, works out the public key and stores both in {{ .KeysDir }}. The folder refuses a key it holds, under any name.</p>
	</div>
</div>

{{ template "wskeyupload" (list . (print "/settings/workspaces/" $ws.Folder) false) }}
{{ end }}
