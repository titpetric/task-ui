{eval $data_head = $data.head}
{include components/page-head.tpl}

<ol class="talk" data-notes>
{foreach $data.messages as $message}	<li class="said{if $message.mine} mine{/if}{if $message.working} working{/if}" id="note-{message.id}">
		<header>
			<span class="author">{eval $data_user = $message.author}{include components/user.tpl}</span>
			<span class="faint" title="{message.stamp}">{message.ago}</span>
{if $message.working}			<span class="badge note-state working">working</span>
{/if}{if $message.run}			<a class="faint mono" href="{base}/runs/{message.run}" title="The run this message comes from">run {message.short}</a>
{/if}		</header>
{if $message.audio}		<audio class="said-audio" controls preload="none" src="{message.audio}"></audio>
{/if}		<div class="said-body prose">{message.body|raw}</div>
	</li>
{/foreach}</ol>

{if $data.replies}<form class="note-form" data-note-form data-parent="{data.id}">
	{eval $data_markdown = $data.editor}{include components/editor.tpl}
	<div class="form-foot note"><span class="faint">Ctrl+Enter posts.</span><button class="btn btn-primary" type="submit">Reply</button></div>
</form>{/if}
