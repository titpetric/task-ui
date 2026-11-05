<div class="table-wrap fit">
<table class="table projects cards">
	<thead><tr><th>Folder</th><th>Sources</th><th>Origin</th><th class="hide-sm hide-narrow">Last commit</th>{if $data_projects.sized}<th class="hide-sm hide-narrow">Size</th>{/if}<th class="actions" aria-label="Actions"></th></tr></thead>
	<tbody>
{foreach $data_projects.rows as $row}		<tr>
			<td class="key nowrap"><a href="{row.href}"><svg class="ic"><use href="#i-folder"/></svg> {row.id}</a></td>
			<td class="nowrap" data-label="Sources">{if $row.bad}<span class="badge bad">error</span> {/if}{if $row.sources}{row.sources} <span class="faint">{row.jobs} jobs</span>{else}<span class="faint">nothing to run</span>{/if}</td>
			<td class="origin" data-label="Origin">{if $row.checkout}<span class="remote mono" title="{row.remote}">{row.remote}</span> <span class="badge">{row.branch}</span>{else}<span class="faint">local folder</span>{/if}</td>
			<td class="nowrap hide-sm hide-narrow" data-label="Last commit">{if $row.checkout}<span class="mono" title="{row.subject}">{row.commit}</span>{if $row.ago} <span class="faint">{row.ago}</span>{/if}{/if}</td>{if $data_projects.sized}<td class="hide-sm hide-narrow" data-label="Size">{eval $data_mini = $row.space}{include components/space-mini.tpl}</td>{/if}
			<td class="actions">{eval $data_menu = $row.menu}{include components/menu.tpl}</td>
		</tr>
{/foreach}	</tbody>
</table>
</div>