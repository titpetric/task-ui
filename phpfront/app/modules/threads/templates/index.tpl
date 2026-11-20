<div class="chat-shell">
	<aside class="chat-list">
		<form method="get" action="{base}/threads" class="filters">
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
		<ol class="timeline">
{foreach $data.chats as $row}			{eval $data_row = $row}{include components/thread-row.tpl}
{/foreach}		</ol>
{/if}

		<h3 class="section-title">Threads</h3>
		<ol class="timeline" data-notes>
{foreach $data.rows as $row}			{eval $data_row = $row}{include components/thread-row.tpl}
{else}			{eval $data_empty = $data.empty}{include components/empty.tpl}
{/foreach}		</ol>
	</aside>

	<section class="chat-pane" data-thread-pane>
		<div class="conversation">
			<div class="chat-blank">
				<h2>{data.head.title}</h2>
				<p>{data.head.lead|raw}</p>
				{eval $data_composer = $data.composer}{include components/composer.tpl}
			</div>
		</div>
	</section>
</div>
