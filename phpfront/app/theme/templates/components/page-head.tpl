{* input scope: data_head. See page-head.md. *}
<div class="page-head">
	<div>
		<h2>{data_head.title}</h2>
{if $data_head.lead}		<p>{data_head.lead|raw}</p>
{/if}	</div>
{if $data_head.actions}	{eval $data_menu = $data_head.actions}
	{include components/menu.tpl}
{/if}</div>
