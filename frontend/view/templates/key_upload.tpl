{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Upload key</h2>
		<p>Saves a private key you already use in {{ .KeysDir }}. task-ui reads its algorithm and size and works out the public key; the name gets the algorithm, as <code>deploy_rsa4096</code>. The folder refuses a key it holds, under any name: a key has one purpose.</p>
	</div>
</div>

<form class="form" data-api="/api/keys/upload" data-method="POST" data-next="/settings/keys?notice=Key+uploaded">
	{{- template "keyname" (index .Form "name") }}
	<div class="field">
		<label for="up-file">Private key file</label>
		<input type="file" id="up-file" data-read-into="up-content">
	</div>
	<div class="field">
		<label for="up-content">Or paste the private key</label>
		<textarea id="up-content" name="content" rows="10" spellcheck="false" autocomplete="off" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----" required></textarea>
	</div>
	{{- template "keypurpose" }}
	<div class="field">
		<label for="up-comment">Comment</label>
		<input type="text" id="up-comment" name="comment" value="task-ui">
	</div>
	{{- template "keypolicies" (list nil (list (index .Form "for")) .Users) }}
	<div class="form-foot">
		<a class="btn" href="/settings/keys">Cancel</a>
		<button class="btn btn-primary" type="submit">Add key</button>
	</div>
</form>
{{ end }}
