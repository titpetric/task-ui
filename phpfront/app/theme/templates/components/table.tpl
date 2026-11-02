{* input scope: data_table. See table.md. *}
<div class="table-wrap">
<table>
{if $data_table.columns}	<colgroup>{foreach $data_table.columns as $column}<col{if $column.class} class="{column.class}"{/if}>{/foreach}</colgroup>
	<thead><tr>{foreach $data_table.columns as $column}<th{if $column.numeric} class="num"{/if}>{column.title}</th>{/foreach}</tr></thead>
{/if}	<tbody>
{foreach $data_table.rows as $row}		{eval $data_row = $row}{include components/table-row.tpl}
{else}		<tr class="empty-row"><td colspan="{data_table.width}">{data_table.empty}</td></tr>
{/foreach}	</tbody>
</table>
</div>
