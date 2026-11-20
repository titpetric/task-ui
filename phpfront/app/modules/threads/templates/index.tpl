{eval $data_head = $data.head}
{include components/page-head.tpl}

<div class="threads-shell">
	<aside class="threads-list">
		<form method="get" action="{base}/threads" class="filters">
			<input type="hidden" name="id" value="{data.selected_id}">
			<label>Words <input type="search" name="q" value="{data.form.q}" spellcheck="false" autocomplete="off"></label>
			<label>Place <input type="text" name="place" value="{data.form.place}" spellcheck="false" autocomplete="off"></label>
			<label for="threads-state">State
				<select id="threads-state" name="state">
					<option value="">any state</option>
{foreach $data.states as $state}<option value="{state}"{if $data.form.state === $state} selected{/if}>{state}</option>
{/foreach}				</select>
			</label>
			<label>By <input type="text" name="user" value="{data.form.user}" spellcheck="false" autocomplete="off"></label>
			<button class="btn" type="submit">Filter</button>
		</form>

{if $data.chats}
		<h3 class="section-title">Chats</h3>
		<ol class="timeline" data-chats>
{foreach $data.chats as $row}			<li class="thread-row{if $row.unread} unread{/if}{if $row.id === $data.selected_id} selected{/if}">
				<a class="thread-open" href="{row.href}" data-thread-id="{row.id}">
					<span class="thread-said">{row.said}</span>
					<span class="thread-meta">
						<svg class="ic thread-kind" aria-label="chat"><use href="#i-note"/></svg>
						<span class="faint" title="{row.stamp}">{row.ago}</span>
{if $row.unread}						<span class="badge unread">{row.unread}</span>
{/if}{if $row.heard}						<svg class="ic" aria-label="Recorded"><use href="#i-play"/></svg>
{/if}					</span>
				</a>
				<span class="thread-who">{eval $data_user = $row.peer}{include components/user.tpl}</span>
			</li>
{/foreach}		</ol>
{/if}

		<ol class="timeline" data-notes>
{foreach $data.rows as $row}			<li class="thread-row{if $row.state} {row.state}{/if}{if $row.id === $data.selected_id} selected{/if}">
				<a class="thread-open" href="{row.href}" data-thread-id="{row.id}">
					<span class="thread-said">{row.said}</span>
					<span class="thread-meta">
						<svg class="ic thread-kind" aria-label="{row.kind}"><use href="#i-{if $row.kind == 'voice'}mic{elseif $row.kind == 'prompt'}prompt{else}note{/if}"/></svg>
						<span class="faint" title="{row.stamp}">{row.ago}</span>
						{if $row.place}<span class="faint mono">{row.place}</span>{/if}
						<span class="faint">{row.messages} messages</span>
{if $row.heard}						<svg class="ic" aria-label="Recorded"><use href="#i-play"/></svg>
{/if}{if $row.state}						<span class="badge note-state {row.state}">{row.state}</span>
{/if}					</span>
				</a>
				<span class="thread-who">{eval $data_user = $row.author}{include components/user.tpl}</span>
			</li>
{else}			<li class="empty-state">Nothing said yet.</li>
{/foreach}		</ol>
	</aside>

	<section class="threads-pane" data-thread-pane>
{if $data.selected}
		{eval $data_pane = $data.selected}{include pane.tpl}
{elseif $data.pane_error}
		<p class="empty-state">{data.pane_error}</p>
{else}
		<div class="threads-new">
			<p class="faint">A new thread. Send a note, or pick one on the left.</p>
			<form class="note-form" data-note-form data-open-thread>
				{eval $data_markdown = $data.editor}{include components/editor.tpl}
				<div class="form-foot note"><span class="faint">Ctrl+Enter sends. Markdown: <code>**bold**</code>, <code>- lists</code>, <code>`code`</code> and <code>```</code> blocks.</span><button class="btn btn-primary" type="submit">Send</button></div>
			</form>
		</div>
{/if}	</section>
</div>
