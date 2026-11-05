{eval $data_head = $data.head}
{include components/page-head.tpl}

<form class="form" data-api="/api/users" data-method="POST" data-next="/settings/users?notice=User+added">
	<div class="field">
		<label for="u-name">Name</label>
		<input type="text" id="u-name" name="name" required pattern="{data.pattern}" spellcheck="false" autocomplete="off" autofocus>
		<small class="faint">Lowercase letters, digits and dashes; it also names their SSH keys.</small>
	</div>
	<div class="field">
		<label for="u-password">Password <span class="faint">at least 8 characters</span></label>
		<input type="password" id="u-password" name="password" required minlength="8" autocomplete="new-password">
		<small class="faint">Stored as a bcrypt hash.</small>
	</div>{eval $data_pick = $data.pick}{include components/folder-pick.tpl}
	<div class="form-foot">
		<a class="btn" href="{data.cancel}">Cancel</a>
		<button class="btn btn-primary" type="submit">Add user</button>
	</div>
</form>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
