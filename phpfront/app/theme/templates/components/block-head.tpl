{* input scope: data_block. See block-head.md. *}
<div class="block-head{if $data_block.empty} is-empty{/if}">
	<h3>{data_block.title}{if $data_block.note} <small>{data_block.note}</small>{/if}</h3>
{if $data_block.empty}	<p class="empty-line">{data_block.empty}</p>
{/if}{if $data_block.buttons.items}	<div class="actions">{eval $data_buttons = $data_block.buttons}{include components/buttons.tpl}	</div>
{/if}</div>
