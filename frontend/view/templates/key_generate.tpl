{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Generate key</h2>
		<p>Creates a key in {{ .KeysDir }}, as <code>&lt;name&gt;_&lt;algorithm&gt;.key</code> and <code>.pub</code>. A name is unique across the folder.</p>
	</div>
</div>

<form class="form" data-api="/api/keys/generate" data-method="POST" data-next="/settings/keys?notice=Key+generated">
	{{- template "keyname" (index .Form "name") }}
	<div class="field">
		<label for="gen-algo">Algorithm</label>
		<select id="gen-algo" name="algo" data-key-algo>
			<option value="ed25519" selected>ed25519</option>
			<option value="ecdsa" data-bits="256">ECDSA P-256</option>
			<option value="ecdsa" data-bits="384">ECDSA P-384</option>
			<option value="ecdsa" data-bits="521">ECDSA P-521</option>
			<option value="rsa" data-bits="2048">RSA 2048</option>
			<option value="rsa" data-bits="3072">RSA 3072</option>
			<option value="rsa" data-bits="4096">RSA 4096</option>
		</select>
		<input type="hidden" name="bits" value="0" data-type="int">
	</div>
	{{- template "keypurpose" }}
	<div class="field">
		<label for="gen-comment">Comment</label>
		<input type="text" id="gen-comment" name="comment" value="task-ui">
	</div>
	{{- template "keypolicies" (list nil (list (index .Form "for")) .Users) }}
	<div class="form-foot">
		<a class="btn" href="/settings/keys">Cancel</a>
		<button class="btn btn-primary" type="submit">Generate</button>
	</div>
</form>
{{ end }}
