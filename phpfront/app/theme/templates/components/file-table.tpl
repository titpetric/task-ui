<div class="table-wrap">
<table class="table file-table col-hover">
	<thead><tr><th>Name</th><th class="num">Size</th><th class="hide-sm">Changed</th><th></th></tr></thead>
	<tbody>
{if $data_files.up}		<tr class="up"><td class="key"><i class="ic ic-folder dir" aria-hidden="true"></i> <a href="{data_files.up}" title="Up a level">..</a></td><td></td><td class="hide-sm"></td><td></td></tr>
{/if}{foreach $data_files.rows as $row}{if $row.sealed}		<tr class="sealed" aria-disabled="true" title="task-ui lists {row.name} but does not open it">
			<td class="key"><i class="ic ic-secret" aria-hidden="true"></i> <span>{row.name}</span></td>
			<td class="num nowrap">{row.size}</td>
			<td class="hide-sm nowrap" title="{row.stamp}">{row.ago}</td>
			<td></td>
		</tr>
{else}		<tr class="{if $row.dir}dir{else}file{/if}">
			<td class="key">{if $row.exec}<i class="shebang" aria-hidden="true">#!</i>{else}<i class="ic ic-{if $row.dir}folder{else}file{/if}{if $row.dir} dir{/if}" aria-hidden="true"></i>{/if}{if $row.href} <a href="{row.href}"{if !$row.dir} data-file="{row.path}"{/if}>{row.name}</a>{else} <span>{row.name}</span>{/if}{if $row.exec} <span class="badge exec" title="Runs: executable, or a shebang on its first line">exec</span>{/if}</td>
			<td class="num nowrap">{row.size}</td>
			<td class="hide-sm nowrap" title="{row.stamp}">{row.ago}</td>
			<td class="actions">{if $row.menu}{eval $data_menu = $row.menu}{include components/menu.tpl}{/if}</td>
		</tr>
{/if}{else}		<tr class="empty"><td colspan="4">This directory is empty.</td></tr>
{/foreach}	</tbody>
</table>
</div>
