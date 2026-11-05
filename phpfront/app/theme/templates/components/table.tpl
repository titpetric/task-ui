<div class="table-wrap{if $data_table.wrap} {data_table.wrap}{/if}">
<table class="table{if $data_table.class} {data_table.class}{/if}">
{if $data_table.header}	<thead>
{foreach $data_table.header as $row}		<tr>{foreach $row.cells as $cell}<th{if $cell.class} class="{cell.class}"{/if}{if $cell.span} colspan="{cell.span}"{/if}>{cell.html|raw}</th>{/foreach}</tr>
{/foreach}	</thead>
{/if}	<tbody>
{foreach $data_table.body as $row}		{eval $data_row = $row}{include components/table-row.tpl}
{/foreach}	</tbody>
{if $data_table.footer}	<tfoot>
{foreach $data_table.footer as $row}		{eval $data_row = $row}{include components/table-row.tpl}
{/foreach}	</tfoot>
{/if}</table>
</div>
