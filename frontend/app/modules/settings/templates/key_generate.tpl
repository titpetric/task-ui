{eval $data_head = $data.head}
{include components/page-head.tpl}

<form class="form" data-api="/api/keys/generate" data-method="POST" data-next="/settings/keys?notice=Key+generated">{eval $data_keyname = $data.name}{include components/key-name.tpl}
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
	</div>{include components/key-purpose.tpl}
	<div class="field">
		<label for="gen-comment">Comment</label>
		<input type="text" id="gen-comment" name="comment" value="task-ui">
	</div>{eval $data_whofield = $data.who}{include components/key-policies.tpl}
	<div class="form-foot">
		<a class="btn" href="{data.cancel}">Cancel</a>
		<button class="btn btn-primary" type="submit">Add key</button>
	</div>
</form>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
