<li class="thread-row{if $data_row.state} {data_row.state}{/if}{if $data_row.unread} unread{/if}{if $data_row.selected} selected{/if}">
	<a class="thread-open" href="{data_row.href}" data-thread-id="{data_row.id}"{if $data_row.selected} aria-current="true"{/if}>
		<span class="thread-said">{data_row.said}</span>
		<span class="thread-meta">
			<svg class="ic thread-kind" aria-label="{data_row.kind}"><use href="#i-{if $data_row.kind == 'voice'}mic{elseif $data_row.kind == 'prompt'}prompt{else}note{/if}"/></svg>
			<span class="faint" title="{data_row.stamp}">{data_row.ago}</span>
			{if $data_row.place}<span class="faint mono thread-place">{data_row.place}</span>{/if}
			{if $data_row.messages > 1 && !$data_row.can_end}<span class="faint thread-count" title="{data_row.messages} messages" aria-label="{data_row.messages} messages">{data_row.messages}</span>{/if}
			{if $data_row.heard}<svg class="ic" aria-label="Recorded"><use href="#i-play"/></svg>{/if}
			{if $data_row.unread}<span class="badge unread">{data_row.unread}</span>{/if}
			{if $data_row.state}<span class="badge note-state {data_row.state}">{data_row.state}</span>{/if}
		</span>
	</a>
{if $data_row.can_end}	<button class="btn btn-sm thread-stop" type="button" data-stop="{data_row.run}" title="Stop the run behind this thread">Stop</button>
{/if}	<span class="thread-who">{eval $data_user = $data_row.person}{include components/user.tpl}</span>
</li>
