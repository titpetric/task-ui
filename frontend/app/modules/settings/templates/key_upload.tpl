{eval $data_head = $data.head}
{include components/page-head.tpl}

<form class="form" data-api="/api/keys/upload" data-method="POST" data-next="/settings/keys?notice=Key+uploaded">{eval $data_keyname = $data.name}{include components/key-name.tpl}
	<div class="field">
		<label for="up-file">Private key file</label>
		<input type="file" id="up-file" data-read-into="up-content">
	</div>
	<div class="field">
		<label for="up-content">Or paste the private key</label>
		<textarea id="up-content" name="content" rows="10" spellcheck="false" autocomplete="off" placeholder="-----BEGIN OPENSSH PRIVATE KEY-----" required></textarea>
	</div>{include components/key-purpose.tpl}
	<div class="field">
		<label for="up-comment">Comment</label>
		<input type="text" id="up-comment" name="comment" value="task-ui">
	</div>{eval $data_whofield = $data.who}{include components/key-policies.tpl}
	<div class="form-foot">
		<a class="btn" href="{data.cancel}">Cancel</a>
		<button class="btn btn-primary" type="submit">Add key</button>
	</div>
</form>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
