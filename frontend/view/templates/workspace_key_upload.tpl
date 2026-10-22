{{ define "content" }}
{{- $ws := .Workspace }}{{ $encrypt := index .Form "encrypt" }}
<div class="page-head">
	<div>
		<h2>{{ if $encrypt }}Upload encryption key{{ else }}Upload key{{ end }}</h2>
		<p>{{ if $encrypt }}Saves an ed25519 private key as the key that encrypts the secrets of {{ $ws.Folder }}. No run gets it.{{ else }}Saves a private key you already use as an SSH key of the workspace {{ $ws.Folder }}. Its runs get it as ws:&lt;name&gt;.{{ end }} task-ui works out the public key from it and stores both in {{ .KeysDir }}. If a pair with the same name exists, task-ui keeps it as a backup.</p>
	</div>
</div>

<form class="form" data-api="/api/workspaces/{{ $ws.Folder }}/keys/upload" data-method="POST" data-next="/settings/workspaces/{{ $ws.Folder }}?notice=Key+uploaded">
	{{- if $encrypt }}
	<input type="hidden" name="encrypt" value="true" data-type="bool">
	{{- end }}
	<div class="field">
		<label for="wk-file">Private key file</label>
		<input type="file" id="wk-file" data-read-into="wk-content">
	</div>
	<div class="field">
		<label for="wk-content">Or paste the private key</label>
		<textarea id="wk-content" name="content" rows="10" spellcheck="false" autocomplete="off" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----" required></textarea>
	</div>
	{{- if not $encrypt }}
	<div class="field">
		<label for="wk-name">Name <span class="faint">lower case letters, digits and dashes: id_&lt;type&gt;_&lt;name&gt;</span></label>
		<input type="text" id="wk-name" name="name" value="" autocomplete="off" placeholder="deploy">
	</div>
	{{- end }}
	<div class="field">
		<label for="wk-comment">Comment</label>
		<input type="text" id="wk-comment" name="comment" value="task-ui {{ $ws.Folder }}">
	</div>
	{{- if not $encrypt }}
	{{- template "keygrants" (list (list) nil .Users) }}
	{{- end }}
	<div class="form-foot">
		<a class="btn" href="/settings/workspaces/{{ $ws.Folder }}">Cancel</a>
		<button class="btn btn-primary" type="submit">Upload</button>
	</div>
</form>
{{ end }}
