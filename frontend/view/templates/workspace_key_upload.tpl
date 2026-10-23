{{ define "content" }}
{{- $ws := .Workspace }}
<div class="page-head">
	<div>
		<h2>Upload key</h2>
		<p>Saves a private key you already use as an SSH key of the workspace {{ $ws.Folder }}. Its runs get it in their ssh agent. task-ui reads its algorithm, works out the public key and stores both in {{ .KeysDir }}. The folder refuses a key it holds, under any name.</p>
	</div>
</div>

<form class="form" data-api="/api/workspaces/{{ $ws.Folder }}/keys/upload" data-method="POST" data-next="/settings/workspaces/{{ $ws.Folder }}?notice=Key+uploaded">
	<div class="field">
		<label for="wk-file">Private key file</label>
		<input type="file" id="wk-file" data-read-into="wk-content">
	</div>
	<div class="field">
		<label for="wk-content">Or paste the private key</label>
		<textarea id="wk-content" name="content" rows="10" spellcheck="false" autocomplete="off" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----" required></textarea>
	</div>
	<div class="field">
		<label for="wk-name">Name <span class="faint">lower case letters, digits, dashes and underscores; empty is ssh</span></label>
		<input type="text" id="wk-name" name="name" value="" autocomplete="off" placeholder="deploy">
	</div>
	<div class="field">
		<label for="wk-comment">Comment</label>
		<input type="text" id="wk-comment" name="comment" value="task-ui {{ $ws.Folder }}">
	</div>
	{{- template "keygrants" (list (list) nil .Users) }}
	<div class="form-foot">
		<a class="btn" href="/settings/workspaces/{{ $ws.Folder }}">Cancel</a>
		<button class="btn btn-primary" type="submit">Upload</button>
	</div>
</form>
{{ end }}
