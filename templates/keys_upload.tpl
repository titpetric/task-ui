{{ define "content" }}
{{ template "crumbs" (list (crumb "Settings" "/settings") (crumb "SSH keys" "/settings/keys") (crumb "Upload" "")) }}
<div class="page-head">
	<div>
		<h2>Upload key</h2>
		<p>Bring your own private key, without a passphrase. The public key is derived from it, and a pair of the same type is moved aside as a backup.</p>
	</div>
</div>

{{- if .Error }}<div class="alert bad" role="alert">{{ .Error }}</div>{{ end }}

<form method="post" action="/settings/keys/upload" enctype="multipart/form-data" class="form">
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
	<div class="form-foot">
		<a class="btn" href="/settings/keys">Cancel</a>
		<button class="btn btn-primary" type="submit">Upload</button>
	</div>
</form>
{{ end }}
