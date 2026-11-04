{* input scope: data_pick. See folder-pick.md. *}	<div class="field" data-folder-pick{if $data_pick.label} role="group" aria-labelledby="folders-label"{/if}>
{if $data_pick.label}		<span class="label" id="folders-label">Folders</span>
{/if}		<label class="check"><input type="radio" name="scope" value="all"{if $data_pick.all} checked{/if}> All folders</label>
		<label class="check"><input type="radio" name="scope" value="custom"{if !$data_pick.all} checked{/if}> Custom</label>
		<div class="folder-choices">
{foreach $data_pick.choices as $choice}			<label class="check"><input type="checkbox" name="folders" value="{choice.id}" data-type="set"{if $choice.checked} checked{/if}{if $data_pick.all} disabled{/if}> <span class="mono">{choice.id}</span></label>
{/foreach}{if !$data_pick.choices}			<span class="faint">No folders yet.</span>
{/if}		</div>
	</div>
