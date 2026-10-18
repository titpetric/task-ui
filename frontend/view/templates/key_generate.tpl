{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Generate key</h2>
		<p>Creates a new pair in {{ .KeysDir }}. An existing pair of the same type is moved aside as a backup.</p>
	</div>
</div>

<form class="form" data-api="/api/keys/generate" data-method="POST" data-next="/settings/keys?notice=Key+generated">
	<div class="field">
		<label for="gen-type">Type</label>
		<select id="gen-type" name="type">
			<option value="ed25519">ed25519</option>
			<option value="ecdsa">ECDSA P-256</option>
			<option value="rsa">RSA 4096</option>
		</select>
	</div>
	<div class="field">
		<label for="gen-comment">Comment</label>
		<input type="text" id="gen-comment" name="comment" value="task-ui">
	</div>
	<div class="form-foot">
		<a class="btn" href="/settings/keys">Cancel</a>
		<button class="btn btn-primary" type="submit">Generate</button>
	</div>
</form>
{{ end }}
