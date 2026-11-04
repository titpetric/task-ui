{* The edit key links form. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}
<form class="form" data-api="{data.api}" data-method="PUT" data-next="{data.saved}">{eval $data_whofield = $data.who}{include components/key-policies.tpl}
	<div class="field">
		<label class="check"><input type="checkbox" name="git"{if $data.git} checked{/if}> git pulls and clones with this key</label>
		<small class="faint">One key at a time; the remotes trust its public half.</small>
	</div>
	<div class="form-foot">
		<a class="btn" href="{data.cancel}">Cancel</a>
		<button class="btn btn-primary" type="submit">Save</button>
	</div>
</form>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
