{if $data_sched.rows}	<div data-flags>
	<div class="table-wrap fit">
	<table class="table fit-cards">
		<thead><tr><th>Expression</th><th>Next run</th><th class="hide-narrow">Last run</th><th>Runs as</th><th>Enabled</th><th>Triage</th><th></th></tr></thead>
		<tbody>
{foreach $data_sched.rows as $row}			<tr data-flag-row data-flag-api="/api/schedules/{row.id}">
				<td class="key" data-label="Expression"><code>{row.expression}</code></td>
				<td data-label="Next run">{if $row.on}{foreach $row.next as $part}<span class="nowrap">{part}</span> {/foreach}{else}<span class="faint">-</span>{/if}</td>
				<td class="nowrap hide-narrow" data-label="Last run">{if $row.last !== ""}{row.last}{else}<span class="faint">never</span>{/if}</td>
				<td data-label="Runs as">{if $row.pick}<form class="run-as" data-api="{row.api}" data-method="PUT" data-next="reload" data-notice="Schedule saved">{eval $data_avatar = $row.user}{include components/avatar.tpl} <select name="user" data-submit aria-label="Runs as">
{eval $data_options = $row.options}{include components/user-options.tpl}		</select></form>{else}{eval $data_user = $row.user}{include components/user.tpl}{/if}</td>
				<td class="flags" data-label="Enabled"><input type="checkbox" name="enabled" data-flag="enabled" data-was="{if $row.on}1{else}0{/if}"{if $row.on} checked{/if}{if !$row.can} disabled{/if} aria-label="Enabled"></td>
				<td class="flags" data-label="Triage"><input type="checkbox" name="triage" data-flag="triage" data-was="{if $row.triage}1{else}0{/if}"{if $row.triage} checked{/if}{if !$row.can} disabled{/if} aria-label="Triage failures with Claude"></td>
				<td class="actions">{if $data_sched.can}{eval $data_menu = $row.menu}{include components/menu.tpl}{/if}</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
{if $data_sched.can}	<div class="table-actions"><button class="btn btn-primary" type="button" data-flag-save data-notice="Schedules saved" disabled>Save</button></div>
{/if}	</div>
{/if}