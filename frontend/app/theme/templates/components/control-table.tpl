<div class="table-wrap">
<table class="table control col-hover">
	<tbody>
{foreach $data_control as $row}		<tr class="named" data-span>
			<td class="key nowrap"{if $row.title} title="{row.title}"{/if}>{if $row.href}<a class="badge" href="{row.href}">{row.label}</a>{else}<span class="badge">{row.label}</span>{/if}</td>
			<td colspan="1">{if $row.empty}<span class="faint">{row.empty}</span>{else}<div class="cc-row">{foreach $row.jobs as $job}{eval $data_cc = $job}{include components/control-job.tpl}{/foreach}</div>{/if}</td>
			<td class="actions">{eval $data_menu = $row.menu}{include components/menu.tpl}</td>
		</tr>
{/foreach}	</tbody>
</table>
</div>
