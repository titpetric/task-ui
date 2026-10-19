{{ define "content" }}
<div class="page-head">
	<div>
		<h2>Generate key</h2>
		<p>Creates a new pair in {{ .KeysDir }}. An existing pair of the same type and name is moved aside as a backup; a name makes a key beside the others, for someone.</p>
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
	<div class="field">
		<label for="gen-name">Name <span class="faint">for a key made for someone: id_&lt;type&gt;_&lt;name&gt;, beside the type's own key; empty is the type's own key</span></label>
		<input type="text" id="gen-name" name="name" pattern="[a-z0-9][a-z0-9-]{0,31}" placeholder="alice" spellcheck="false" autocomplete="off">
	</div>
	{{- template "keypolicies" (list (list) (list) .Users) }}
	<div class="form-foot">
		<a class="btn" href="/settings/keys">Cancel</a>
		<button class="btn btn-primary" type="submit">Generate</button>
	</div>
</form>
{{ end }}
