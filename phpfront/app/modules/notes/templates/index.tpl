{* The notes board. The scope is {data}; the chrome is {chrome}. *}
<div class="page-head">
	<div>
		<h2>Notes</h2>
		<p>Leave a note for someone with <code>@name</code>. If you can run jobs, the note is an instruction to the people it names. It stays open until they mark it working, and then done or failed, here or over MCP.{if $data.agent} Name <span class="mention">@{data.agent}</span> in a note or a reply to start the agent on the thread. It works on the request and replies here when it finishes, or explains why it can't.{/if}</p>
	</div>
</div>

{if $data.edits}
<form class="note-form" method="post" action="{base}/notes">
	<textarea name="body" rows="3" required data-mentions placeholder="{if $data.agent}@{data.agent} update the weather job to …{else}@name …{/if}" aria-label="A new note"></textarea>
	<div class="form-foot note"><span class="faint">Ctrl+Enter posts. Markdown: <code>**bold**</code>, <code>- lists</code>, <code>`code`</code> and <code>```</code> blocks.</span><button class="btn btn-primary" type="submit">Post note</button></div>
</form>
{/if}

<div class="notes" data-notes>
{foreach $data.threads as $thread}
	{eval $data_note = $thread}
	{include components/note.tpl}
{else}
	<p class="empty-state">No notes yet.</p>
{/foreach}
</div>
