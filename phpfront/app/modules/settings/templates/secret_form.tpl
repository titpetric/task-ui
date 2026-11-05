{eval $data_head = $data.head}
{include components/page-head.tpl}

<form class="form" data-api="/api/secrets" data-method="POST" data-next="{data.next}">{if $data.id}<input type="hidden" name="id" value="{data.id}">{/if}
	<div class="field">
		<label for="s-scope">Applies to</label>
		<select id="s-scope" name="scope">
{foreach $data.options as $option}			<option value="{option.value}"{if $option.selected} selected{/if}>{option.title}</option>
{/foreach}		</select>
	</div>
	<div class="field">
		<label for="s-name">Name</label>
		<input type="text" id="s-name" name="name" value="{data.name}" pattern="[A-Za-z_][A-Za-z0-9_]*" spellcheck="false" autocomplete="off" required>
	</div>
	<div class="field">
		<label for="s-content">Value <span class="faint">{data.hint}</span></label>
		{if $data.blank}<textarea id="s-content" name="content" rows="3" spellcheck="false" autocomplete="new-password" placeholder="unchanged"></textarea>{else}<textarea id="s-content" name="content" rows="3" spellcheck="false" autocomplete="off">{data.content}</textarea>{/if}
	</div>
	<label class="check"><input type="checkbox" name="is_secret"{if $data.secret} checked{/if}> Secret <span class="faint">the value is encrypted and never shown again; an env value is listed on the page</span></label>
	<div class="form-foot">
		<a class="btn" href="{data.cancel}">Cancel</a>
		<button class="btn btn-primary" type="submit">{data.submit}</button>
	</div>
</form>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
