<div class="notes wf-thread" data-workflow-thread="{data_thread.id}"{if $data_thread.running} data-follow="{data_thread.workflowId}" data-page="{base}{data_thread.page}"{/if}>
	<article class="note instruction {data_thread.noteState}">
		<header>
			<span class="author">{eval $data_user = $data_thread.user}{include components/user.tpl}</span>
			<span class="faint" title="{data_thread.stamp}">{data_thread.ago}</span>
			<span class="spacer"></span>
		</header>
{if $data_thread.question}		<div class="note-body">{data_thread.question|raw}</div>
{else}		<div class="note-body"><p class="faint">Ran {data_thread.name}, with no input.</p></div>
{/if}		<ol class="replies">
			<li>
				<header>
{if $data_thread.agent}					<span class="author">{eval $data_user = $data_thread.agent}{include components/user.tpl}</span>
{else}					<span class="author">{data_thread.name}</span>
{/if}{if $data_thread.running} <span class="badge note-state working">working</span>
{else} <span class="faint" title="{data_thread.doneStamp}">{data_thread.doneAgo}</span> <span class="state"><i class="dot {data_thread.state}"></i>{data_thread.state}</span>
{/if}				</header>
{if $data_thread.answer}				<div class="note-body">{data_thread.answer|raw}</div>
{elseif $data_thread.running}				<div class="note-body"><p class="faint">{data_thread.waiting}</p></div>
{else}				<div class="note-body"><p>{data_thread.ended}</p></div>
{/if}			</li>
		</ol>
	</article>
</div>
