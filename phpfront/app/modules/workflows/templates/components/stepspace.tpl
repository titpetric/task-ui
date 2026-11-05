<div class="field">
	<span class="label">Workspace</span>
	<select data-step-space aria-label="Workspace">
{if $data_space.none}		<option value="" data-workspace="" data-rel="."{if $data_space.noneSelected} selected{/if}>none (the workspace root)</option>
{/if}{foreach $data_space.spaces as $one}		<option value="{one.id}"{if $one.selected} selected{/if}>{one.title}</option>
{/foreach}{if $data_space.flows}		<option value="workflow:"{if $data_space.flowsSelected} selected{/if}>Workflows</option>
{/if}	</select>
</div>
