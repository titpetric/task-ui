{{ define "content" }}
{{- $ws := .Workspace }}
<div class="page-head">
	<div>
		<h2>Generate key</h2>
		<p>Creates a new SSH key pair for the workspace {{ $ws.Folder }}, in {{ .KeysDir }}. Its runs get it in their ssh agent as ws:&lt;name&gt;. If a pair with the same name exists, task-ui keeps it as a backup.</p>
	</div>
</div>

<form class="form" data-api="/api/workspaces/{{ $ws.Folder }}/keys/generate" data-method="POST" data-next="/settings/workspaces/{{ $ws.Folder }}?notice=Key+generated">
	<div class="field">
		<label for="wk-type">Type</label>
		<select id="wk-type" name="type">
			<option value="ed25519">ed25519</option>
			<option value="ecdsa">ECDSA P-256</option>
			<option value="rsa">RSA 4096</option>
		</select>
	</div>
	<div class="field">
		<label for="wk-name">Name <span class="faint">lower case letters, digits and dashes: id_&lt;type&gt;_&lt;name&gt;</span></label>
		<input type="text" id="wk-name" name="name" value="" autocomplete="off" placeholder="deploy">
	</div>
	<div class="field">
		<label for="wk-comment">Comment</label>
		<input type="text" id="wk-comment" name="comment" value="task-ui {{ $ws.Folder }}">
	</div>
	{{- template "keygrants" (list (list) nil .Users) }}
	<div class="form-foot">
		<a class="btn" href="/settings/workspaces/{{ $ws.Folder }}">Cancel</a>
		<button class="btn btn-primary" type="submit">Generate</button>
	</div>
</form>
{{ end }}
