{{ define "content" }}
{{- $ws := .Workspace }}
<div class="page-head">
	<div>
		<h2>Generate key</h2>
		<p>Creates an SSH key of the workspace {{ $ws.Folder }} in {{ .KeysDir }}, named <span class="mono">workspace_{{ $ws.Folder }}_&lt;name&gt;_&lt;algorithm&gt;</span>. Its runs get it in their ssh agent. A name is unique across the folder.</p>
	</div>
</div>

<form class="form" data-api="/api/workspaces/{{ $ws.Folder }}/keys/generate" data-method="POST" data-next="/settings/workspaces/{{ $ws.Folder }}?notice=Key+generated">
	<div class="field">
		<label for="wk-name">Name <span class="faint">lower case letters, digits, dashes and underscores; empty is ssh</span></label>
		<input type="text" id="wk-name" name="name" value="" autocomplete="off" placeholder="deploy">
	</div>
	<div class="field">
		<label for="wk-algo">Algorithm</label>
		<select id="wk-algo" name="algo" data-key-algo>
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
