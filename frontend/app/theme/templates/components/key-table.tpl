{if $data_keys}	<div class="table-wrap fit">
	<table class="table keys key-table cards">
		{include components/key-head.tpl}
		<tbody>
{foreach $data_keys as $row}{eval $data_key = $row}{include components/key-row.tpl}
{/foreach}		</tbody>
	</table>
	</div>
{/if}
