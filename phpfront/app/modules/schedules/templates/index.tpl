{eval $data_head = $data.head}
{include components/page-head.tpl}

<section class="block">
{eval $data_block = $data.block}{include components/block-head.tpl}
{if $data.schedules}	<div class="table-wrap fit">
	<table class="table fit-cards">
		<thead><tr><th>Folder</th><th>Job</th><th>Expression</th><th>Next run</th><th class="hide-narrow">Last run</th><th>Runs as</th><th>State</th><th></th></tr></thead>
		<tbody>
{foreach $data.schedules as $row}			<tr>
				<td data-label="Folder">{if $row.workflow}<svg class="ic lead"><use href="#i-flow"/></svg>workflow{else}<a href="{base}{row.folderHref}">{row.folder}</a>{/if}</td>
				<td class="key" data-label="Job">{if $row.jobHref !== ""}<a href="{base}{row.jobHref}">{row.job}</a>{else}{row.job}{/if}{if $row.args !== ""} <code>{row.args}</code>{/if}</td>
				<td data-label="Expression"><code class="nowrap">{row.expression}</code></td>
				<td data-label="Next run">{if $row.on}{foreach $row.next as $part}<span class="nowrap">{part}</span> {/foreach}{else}<span class="faint">-</span>{/if}</td>
				<td class="nowrap hide-narrow" data-label="Last run">{if $row.last !== ""}{row.last}{else}<span class="faint">never</span>{/if}</td>
				<td data-label="Runs as">{if $row.pick}<form class="run-as" data-api="{row.api}" data-method="PUT" data-next="reload" data-notice="Schedule saved">{eval $data_avatar = $row.user}{include components/avatar.tpl} <select name="user" data-submit aria-label="Runs as">
{eval $data_options = $row.options}{include components/user-options.tpl}		</select></form>{else}{eval $data_user = $row.user}{include components/user.tpl}{/if}</td>
				<td data-label="State">{if $row.on}<span class="badge ok">enabled</span>{else}<span class="badge">disabled</span>{/if}{if $row.triage} <span class="badge signal" title="Claude triages a failed run">triage</span>{/if}{if $row.source} <span class="badge" title="Declared in source; an admin sets Runs as and enables it">{row.source}</span>{/if}</td>
				<td class="actions">{eval $data_menu = $row.menu}{include components/menu.tpl}</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
{/if}</section>
