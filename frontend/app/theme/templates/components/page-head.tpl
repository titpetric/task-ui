<div class="page-head">
	<div>
		<h2>{if $data_head.icon}<svg class="ic"><use href="#i-{data_head.icon}"/></svg> {/if}{data_head.title}{if $data_head.badge} <span class="badge{if $data_head.tone} {data_head.tone}{/if}">{data_head.badge}</span>{/if}{foreach $data_head.tags as $tag} <span class="badge">{tag}</span>{/foreach}{if $data_head.note} <span class="term-title" title="{data_head.note}">{data_head.note}</span>{/if}</h2>
{if $data_head.lead || $data_head.leads}		<p>{data_head.lead|raw}</p>
{/if}{if $data_head.faint}		<p class="faint">{data_head.faint}</p>
{/if}	</div>
{if $data_head.show}	<div class="actions"{if $data_head.live} data-actions{/if}>
{if $data_head.run}		{eval $data_run = $data_head.run}{include components/run-buttons.tpl}
{/if}{if $data_head.buttons.items}		{eval $data_buttons = $data_head.buttons}{include components/buttons.tpl}
{/if}{if $data_head.actions}		{eval $data_menu = $data_head.actions}
		{include components/menu.tpl}
{/if}	</div>
{/if}</div>
