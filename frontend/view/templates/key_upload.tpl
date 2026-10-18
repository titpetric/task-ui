{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Upload key</h2>
		<p>Stores a private key you already use as the pair of its type, in {{ .KeysDir }}. The public key is derived from it. A pair of the same type that is there is moved aside as a backup.</p>
	</div>
</div>

<form class="form" data-api="/api/keys/upload" data-method="POST" data-next="/settings/keys?notice=Key+uploaded">
	<div class="field">
		<label for="up-file">Private key file</label>
		<input type="file" id="up-file" data-read-into="up-content">
	</div>
	<div class="field">
		<label for="up-content">Or paste the private key</label>
		<textarea id="up-content" name="content" rows="10" spellcheck="false" autocomplete="off" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----" required></textarea>
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
