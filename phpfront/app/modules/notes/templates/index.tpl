{eval $data_head = $data.head}
{include components/page-head.tpl}

<form class="note-form" data-note-form>
	<textarea name="body" rows="3" required data-mentions aria-label="A new note"></textarea>
	<div class="form-foot note"><span class="faint">Ctrl+Enter posts. Markdown: <code>**bold**</code>, <code>- lists</code>, <code>`code`</code> and <code>```</code> blocks.</span><button class="btn btn-primary" type="submit">Post note</button></div>
</form>

<div class="notes" data-notes>
{foreach $data.threads as $thread}
	{eval $data_note = $thread}
	{include components/note.tpl}
{else}
	<p class="empty-state">No notes yet.</p>
{/foreach}
</div>
