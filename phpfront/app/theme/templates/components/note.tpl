{* input scope: data_note. See note.md. *}
	<article class="note{if $data_note.state} instruction {data_note.state}{/if}" id="note-{data_note.id}">
		<header>
			<span class="author">{eval $data_user = $data_note.author}{include components/user.tpl}</span>
			<span class="faint">{data_note.ago}</span>
{if $data_note.edited}			<span class="faint edited">edited</span>
{/if}{if $data_note.state}			<span class="badge note-state {data_note.state}">{data_note.state}</span>
{/if}{if $data_note.run}			<a class="faint mono" href="{base}/runs/{data_note.run}" title="The run the agent worked in">run {data_note.short}</a>
{/if}			<span class="spacer"></span>
			{eval $data_menu = $data_note.menu}
			{include components/menu.tpl}
		</header>
		<div class="note-body">{data_note.body|raw}</div>
{if $data_note.replies}		<ol class="replies">
{foreach $data_note.replies as $reply}			<li>
				<header><span class="author">{eval $data_user = $reply.author}{include components/user.tpl}</span> <span class="faint">{reply.ago}</span>{if $reply.edited} <span class="faint edited">edited</span>{/if}<span class="spacer"></span>{eval $data_menu = $reply.menu}{include components/menu.tpl}</header>
				<div class="note-body">{reply.body|raw}</div>
			</li>
{/foreach}		</ol>
{/if}{if $data_note.replyable}		<details class="reply">
			<summary>Reply</summary>
			<form class="note-form" method="post" action="{base}/notes">
				<input type="hidden" name="parent" value="{data_note.id}">
				<textarea name="body" rows="2" required data-mentions aria-label="A reply"></textarea>
				<div class="form-foot note"><span></span><button class="btn" type="submit">Reply</button></div>
			</form>
		</details>
{/if}	</article>
