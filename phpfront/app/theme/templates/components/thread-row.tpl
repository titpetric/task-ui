<li class="{data_row.className}">
	<a class="thread-open" href="{data_row.href}" data-thread-id="{data_row.id}"{if $data_row.selected} aria-current="true"{/if}>
		<span class="thread-said">{data_row.said}</span>
		<span class="thread-meta">
			<i class="ic ic-{if $data_row.kind == 'voice'}mic{elseif $data_row.kind == 'prompt'}prompt{else}note{/if} thread-kind" role="img" aria-label="{data_row.kind}"></i>
			<span class="faint" title="{data_row.stamp}">{data_row.ago}</span>
			{if $data_row.place}<span class="faint mono thread-place">{data_row.place}</span>{/if}
			{if $data_row.replies > 0}<span class="faint thread-count" title="{data_row.replies} replies" aria-label="{data_row.replies} replies">{data_row.replies}</span>{/if}
			{if $data_row.context}<span class="faint thread-context" title="Context read across this thread">{data_row.context}</span>{/if}
			{if $data_row.heard}<i class="ic ic-play" role="img" aria-label="Recorded"></i>{/if}
			{if $data_row.unread}<span class="badge unread">{data_row.unread}</span>{/if}
			{if $data_row.state}<span class="badge note-state {data_row.state}">{data_row.state}</span>{/if}
		</span>
	</a>
{if $data_row.can_end}	<button class="btn-icon thread-stop" type="button" data-stop="{data_row.run}" title="Stop the run behind this thread" aria-label="Stop the run behind this thread"><i class="ic ic-stop" aria-hidden="true"></i></button>
{/if}	<span class="thread-who">{eval $data_user = $data_row.person}{include components/user.tpl}</span>
</li>
