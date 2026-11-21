{eval $data_head = $data.head}
{include components/page-head.tpl}

{if $data.create}<form class="form orb-new" data-orb-create data-api="{base}/api" data-base="{base}">
	<label class="field">
		<span>New meeting</span>
		<input type="text" name="title" maxlength="120" placeholder="What it's about" autocomplete="off">
	</label>
	<button class="btn btn-primary" type="submit"><i class="ic ic-mic" aria-hidden="true"></i>Start meeting</button>
</form>
{/if}
<section class="block">
{eval $data_block = $data.block}{include components/block-head.tpl}
{if $data.rows}	<div class="table-wrap">
	<table class="table cards" data-api="{base}/api">
		<thead><tr><th>Meeting</th><th>State</th><th>In the room</th><th>Started</th><th>By</th><th></th></tr></thead>
		<tbody>
{foreach $data.rows as $row}			<tr>
				<td class="key"><a href="{base}/orb/{row.id}">{row.title}</a></td>
				<td data-label="State"><span class="badge {row.state}">{row.state}</span></td>
				<td data-label="In the room">{if $row.open}{row.here}{else}-{/if}</td>
				<td class="nowrap" data-label="Started">{row.ago}</td>
				<td data-label="By">{eval $data_user = $row.user}{include components/user.tpl}</td>
				<td class="actions"><a class="btn btn-sm" href="{base}/orb/{row.id}">{row.action}</a>{if $row.delete}
					<button class="btn btn-danger btn-sm" type="button" data-orb-delete="{row.id}" data-confirm="Delete this meeting, its transcript and its audio?">Delete</button>{/if}</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
{else}	<p class="empty">{data_block.empty}</p>
{/if}</section>
