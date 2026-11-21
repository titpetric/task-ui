{eval $data_head = $data.head}
{include components/page-head.tpl}
<section class="block">
<p class="faint">{data.count} threads, {data.replies} replies, {data.tools} tool calls, {data.tokens} context, {data.cost}</p>
<div class="table-wrap">
<table class="table">
	<thead><tr><th>Thread</th><th class="num">Replies</th><th class="num">Tools</th><th class="num">Context</th><th class="num">Cost</th><th>State</th><th>Last</th></tr></thead>
	<tbody>
{foreach $data.rows as $row}		<tr>
			<td class="key"><a href="{row.href}">{row.said}</a></td>
			<td class="num">{row.replies}</td>
			<td class="num">{row.tools}</td>
			<td class="num">{row.context}</td>
			<td class="num">{row.cost}</td>
			<td>{if $row.state}<span class="badge note-state {row.state}">{row.state}</span>{/if}</td>
			<td class="faint" title="{row.stamp}">{row.ago}</td>
		</tr>
{/foreach}	</tbody>
</table>
</div>
</section>
