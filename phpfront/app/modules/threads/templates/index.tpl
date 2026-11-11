{eval $data_head = $data.head}
{include components/page-head.tpl}

<form class="note-form" data-note-form>
	{eval $data_markdown = $data.editor}{include components/editor.tpl}
	<div class="form-foot note"><span class="faint">Ctrl+Enter posts. Markdown: <code>**bold**</code>, <code>- lists</code>, <code>`code`</code> and <code>```</code> blocks.</span><button class="btn btn-primary" type="submit">Post note</button></div>
</form>

<ol class="timeline" data-notes>
{foreach $data.rows as $row}	<li class="thread-row{if $row.state} {row.state}{/if}">
		<a class="thread-open" href="{row.href}">
			<span class="thread-said">{row.said}</span>
			<span class="thread-meta">
				<svg class="ic thread-kind" aria-label="{row.kind}"><use href="#i-{if $row.kind == 'voice'}mic{elseif $row.kind == 'prompt'}prompt{else}note{/if}"/></svg>
				<span class="faint" title="{row.stamp}">{row.ago}</span>
				<span class="faint">{row.messages} messages</span>
{if $row.heard}				<svg class="ic" aria-label="Recorded"><use href="#i-play"/></svg>
{/if}{if $row.state}				<span class="badge note-state {row.state}">{row.state}</span>
{/if}			</span>
		</a>
		<span class="thread-who">{eval $data_user = $row.author}{include components/user.tpl}</span>
	</li>
{else}	<li class="empty-state">Nothing said yet.</li>
{/foreach}</ol>
