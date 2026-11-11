{if $data_sched.rows}	<div class="table-wrap fit">
	<table class="table fit-cards">
		<thead><tr><th>Expression</th><th>Next run</th><th class="hide-narrow">Last run</th><th>Runs as</th><th>State</th><th></th></tr></thead>
		<tbody>
{foreach $data_sched.rows as $row}			<tr>
				<td class="key" data-label="Expression"><code>{row.expression}</code></td>
				<td data-label="Next run">{if $row.on}{foreach $row.next as $part}<span class="nowrap">{part}</span> {/foreach}{else}<span class="faint">-</span>{/if}</td>
				<td class="nowrap hide-narrow" data-label="Last run">{if $row.last !== ""}{row.last}{else}<span class="faint">never</span>{/if}</td>
				<td data-label="Runs as">{if $row.pick}<form class="run-as" data-api="{row.api}" data-method="PUT" data-next="reload" data-notice="Schedule saved">{eval $data_avatar = $row.user}{include components/avatar.tpl} <select name="user" data-submit aria-label="Runs as">
{eval $data_options = $row.options}{include components/user-options.tpl}		</select></form>{else}{eval $data_user = $row.user}{include components/user.tpl}{/if}</td>
				<td data-label="State">{if $row.on}<span class="badge ok">enabled</span>{else}<span class="badge">disabled</span>{/if}{if $row.triage} <span class="badge signal" title="Claude triages a failed run">triage</span>{/if}</td>
				<td class="actions">{if $data_sched.can}{eval $data_menu = $row.menu}{include components/menu.tpl}{/if}</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
{/if}