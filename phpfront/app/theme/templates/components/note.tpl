	<article class="note{if $data_note.state} instruction {data_note.state}{/if}" id="note-{data_note.id}">
		<header>
			<span class="author">{eval $data_user = $data_note.author}{include components/user.tpl}</span>
			<span class="faint" title="{data_note.stamp}">{data_note.ago}</span>
{if $data_note.edited}			<span class="faint edited" title="Edited {data_note.editedStamp}">edited</span>
{/if}{if $data_note.state}			<span class="badge note-state {data_note.state}">{data_note.state}</span>
{/if}{if $data_note.run}			<a class="faint mono" href="{base}/runs/{data_note.run}" title="The run the agent worked in">run {data_note.short}</a>
{/if}			<span class="spacer"></span>
			{eval $data_menu = $data_note.menu}
			{include components/menu.tpl}
		</header>
		<div class="note-body"{if $data_note.edits} data-edit="{data_note.id}" data-body="{data_note.raw}" title="Click to edit"{/if}>{data_note.body|raw}</div>
{if $data_note.replies}		<ol class="replies">
{foreach $data_note.replies as $reply}			<li{if $reply.working} class="rerun"{/if}>
				<header><span class="author">{eval $data_user = $reply.author}{include components/user.tpl}</span> <span class="faint" title="{reply.stamp}">{reply.ago}</span>{if $reply.edited} <span class="faint edited" title="Edited {reply.editedStamp}">edited</span>{/if}{if $reply.working} <span class="badge note-state working">working</span>{/if}{if $reply.run} <a class="faint mono" href="{base}/runs/{reply.run}" title="The run this reply comes from">run {reply.short}</a>{/if}<span class="spacer"></span>{eval $data_menu = $reply.menu}{include components/menu.tpl}</header>
				<div class="note-body"{if $reply.edits} data-edit="{reply.id}" data-body="{reply.raw}" title="Click to edit"{/if}>{reply.body|raw}</div>
			</li>
{/foreach}		</ol>
{/if}		<details class="reply">
			<summary>Reply</summary>
			<form class="note-form" data-note-form data-parent="{data_note.id}">
				<textarea name="body" rows="2" required data-mentions aria-label="A reply"></textarea>
				<div class="form-foot note"><span></span><button class="btn" type="submit">Reply</button></div>
			</form>
		</details>
	</article>
