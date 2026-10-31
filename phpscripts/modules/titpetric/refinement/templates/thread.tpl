<section class="refinement">
	<p class="refinement-back"><a href="{$base}">&larr; Every thread</a></p>

	{* What is being refined is the picture. The run's prose is how it got
	   there: kept, and behind a disclosure. A report that embeds nothing
	   has no picture, and then its prose is the output and stands open. *}
	{if $visual}
	<div class="refinement-output">{$output|raw}</div>
		{if $thinking}
		<details class="refinement-thinking">
			<summary>show thinking</summary>
			<div class="refinement-report">{$thinking|raw}</div>
		</details>
		{/if}
	{elseif $thinking}
	<div class="block refinement-report">
		<h3>The design <small>as the last run left it</small></h3>
		{$thinking|raw}
	</div>
	{else}
	<p class="faint">The first run has not written its report yet. It appears here once the run ends.</p>
	{/if}

	{* Links, not listings. The thread's files are a folder of the artifact
	   tree and the artifacts view browses one; the scratch pad is what the
	   runs keep for themselves, which is worth reading when a run
	   surprises you. *}
	<p class="refinement-files">
		<a class="btn btn-sm" href="/artifacts?path={$thread.id}">The files of this thread</a>
		{if $scratchpad}<a class="btn btn-sm" href="/artifacts/files/{$thread.id}/SCRATCHPAD.md">The scratch pad</a>{/if}
	</p>

	<ol class="refinement-thread">
	{foreach $entries as $entry}
		{if $entry.folded_count}
		<li class="refinement-folded">
			<details>
				<summary>{$entry.folded_count} folded into the request below</summary>
				<ol>
				{foreach $entry.folded as $comment}
					<li class="refinement-comment">
						<div class="refinement-comment-head">
							<strong>{$comment.user}</strong>
							<span class="faint">{$comment.created_at}</span>
							{if $comment.run}<span class="state" data-run-id="{$comment.run}"></span>{/if}{if $comment.note_id}<a class="badge" href="/notes#{$comment.note_id}">the run</a>{/if}
							{if $comment.correction}<span class="badge">rule</span>{/if}
						</div>
						<div class="refinement-body">{$comment.html|raw}</div>
					</li>
				{/foreach}
				</ol>
			</details>
		</li>
		{/if}
		<li class="refinement-comment">
			<div class="refinement-comment-head">
				<strong>{$entry.comment.user}</strong>
				<span class="faint">{$entry.comment.created_at}</span>
				{if $entry.comment.run}<span class="state" data-run-id="{$entry.comment.run}"></span>{/if}{if $entry.comment.note_id}<a class="badge" href="/notes#{$entry.comment.note_id}">the run</a>{/if}
				{if $entry.applied}<span class="badge signal">applied</span>{/if}
				{if $entry.comment.correction}<span class="badge">rule</span>{/if}
				{* Deleting the design deletes the thread: the rules correct it,
				   and without it they correct nothing. The artifacts stay. *}
				<form class="refinement-delete" method="post" action="{$base}/threads/{$thread.id}/comments/{$entry.comment.id}/delete"
					{if $entry.comment.id == $design}data-confirm="Delete the design, and the whole thread with it? Its rules and its runs go. The files stay in the artifacts, under {$thread.id}."{else}data-confirm="Delete this request? The next Apply all will not carry it. The files it produced stay."{/if}>
					<button class="btn-icon" type="submit" title="Delete" aria-label="Delete"><svg class="ic"><use href="#i-trash"/></svg></button>
				</form>
			</div>
			<div class="refinement-body">{$entry.comment.html|raw}</div>
		</li>
	{/foreach}
	</ol>

	{* One run over every rule at once, rather than one run per rule over
	   what the run before it left. *}
	<form class="form refinement-apply" method="post" action="{$base}/threads/{$thread.id}/apply">
		<p class="faint">Folds every request above into one and runs it over the files as they stand, so the result satisfies all of them together rather than one after another.</p>
		<button class="btn" type="submit"><svg class="ic"><use href="#i-collapse"/></svg>Apply all</button>
	</form>

	{* The rules generalise and the subject does not. A fork is the next
	   subject with this thread's rules already in it, on a folder of its
	   own: what travels is how the work is done. *}
	<form class="form refinement-fork" method="post" action="{$base}/threads/{$thread.id}/fork">
		<label for="refinement-fork">Fork</label>
		<p class="faint" id="refinement-fork-help">A new thread on another subject, carrying every rule above. Say what it is and what it should produce; the rules come along and the files do not.</p>
		<textarea id="refinement-fork" name="want" rows="2" aria-describedby="refinement-fork-help" placeholder="A deployment topology diagram: the hosts, what runs on each and the links between them." required></textarea>
		<div class="form-foot">
			<button class="btn" type="submit"><svg class="ic"><use href="#i-flow"/></svg>Fork</button>
		</div>
	</form>

	<form class="form refinement-comment-form" method="post" action="{$base}/threads/{$thread.id}/comments">
		<label for="refinement-comment">Refine it</label>
		<p class="faint" id="refinement-comment-help">One change per comment. Each one starts a run over the artifacts as they stand now.</p>
		<textarea id="refinement-comment" name="comment" rows="3" aria-describedby="refinement-comment-help" placeholder="use ULID as primary keys" required></textarea>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Send</button>
		</div>
	</form>
</section>
