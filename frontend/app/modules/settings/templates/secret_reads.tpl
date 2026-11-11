{eval $data_head = $data.head}
{include components/page-head.tpl}

<section class="block">
	<h3>{data.heading} <small>{data.order}</small></h3>
	<div class="table-wrap">
	<table class="table secret-reads">
		<thead><tr><th>When</th><th>Run</th><th>Job</th><th>By</th></tr></thead>
		<tbody>
{foreach $data.rows as $row}			<tr>
				<td class="nowrap" title="{row.stamp}">{row.ago}</td>
				<td><a class="faint mono" href="{row.runHref}" title="{row.run}">{row.short}</a></td>
				<td class="key">{row.key}</td>
				<td>{eval $data_user = $row.user}{include components/user.tpl}</td>
			</tr>
{else}			<tr class="empty"><td colspan="4">No run got {data.name} yet.</td></tr>
{/foreach}		</tbody>
	</table>
	</div>
{eval $data_pager = $data.pager}
{include components/pager.tpl}
</section>
{if $data.trail}{eval $data_trail = $data.trail}{include components/setup-trail.tpl}{/if}
