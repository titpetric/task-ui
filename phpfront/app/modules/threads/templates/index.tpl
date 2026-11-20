<div class="chat-shell">
	<aside class="chat-list">
		<div class="chat-list-top">
			<button type="button" class="btn btn-sm" data-new-thread title="Start a new thread"><svg class="ic"><use href="#i-plus"/></svg> New thread</button>
		</div>
		<form method="get" action="{base}/threads" class="chat-filters">
			<input class="chat-search" type="search" name="q" value="{data.form.q}" placeholder="Search threads" aria-label="Words of a thread" aria-keyshortcuts="/" title="Press / to search" spellcheck="false" autocomplete="off">
			<details class="chat-filters-more"{if $data.detailed} open{/if}>
				<summary class="faint">Filters{if $data.detailed} <span class="badge signal">on</span>{/if}</summary>
				<div class="chat-filter-fields">
					<label>Place <input type="text" name="place" value="{data.form.place}" list="threads-places" spellcheck="false" autocomplete="off"></label>
					<label for="threads-state">State
						<select id="threads-state" name="state">
							<option value="">any state</option>
{foreach $data.states as $state}<option value="{state}"{if $data.form.state === $state} selected{/if}>{state}</option>
{/foreach}						</select>
					</label>
					<label>By <input type="text" name="user" value="{data.form.user}" spellcheck="false" autocomplete="off"></label>
					<div class="chat-filter-actions">
						<button class="btn btn-sm" type="submit">Apply</button>
{if $data.filtered}						<a class="btn btn-sm" href="{base}/threads">Clear</a>
{/if}					</div>
				</div>
			</details>
{if $data.places}			<datalist id="threads-places">
{foreach $data.places as $place}				<option value="{place}"></option>
{/foreach}			</datalist>
{/if}		</form>

{if $data.chats && !$data.filtered}
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
				<h2>What should we work on?</h2>
				<p>Ask Claude in a folder, or leave a note for the people here. Every prompt becomes a thread and stays here.</p>
				{eval $data_composer = $data.composer}{include components/composer.tpl}
			</div>
		</div>
	</section>
</div>
