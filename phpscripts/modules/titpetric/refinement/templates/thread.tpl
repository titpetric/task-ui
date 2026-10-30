<section class="refinement">
	<p class="refinement-back"><a href="{$base}">&larr; Every thread</a></p>

	{if $report}
	<div class="block refinement-report">
		<h3>The design <small>as the last run left it</small></h3>
		{$report|raw}
	</div>
	{else}
	<p class="faint">The first run has not written its report yet. It appears here once the run ends.</p>
	{/if}

	{* Links, not listings. The thread's files are a folder of the artifact
	   tree and the artifacts view browses one; CONTEXT.md is what the runs
	   keep for themselves, which is worth reading when a run surprises
	   you. *}
	<p class="refinement-files">
		<a class="btn btn-sm" href="/artifacts?path={$thread.id}">The files of this thread</a>
		{if $context}<a class="btn btn-sm" href="/artifacts/files/{$thread.id}/CONTEXT.md">The agent's context</a>{/if}
	</p>

	<ol class="refinement-thread">
	{foreach $entries as $entry}
		{if $entry.folded_count}
		<li class="refinement-folded">
			<details>
				<summary>{$entry.folded_count} folded into the request below</summary>
				<ol>
				{foreach $entry.folded as $comment}
					<li class="refinement-comment" data-run-state="{$comment.state}">
						<div class="refinement-comment-head">
							<strong>{$comment.user}</strong>
							<span class="faint">{$comment.created_at}</span>
							{if $comment.note_id}<a class="badge" href="/notes#{$comment.note_id}">the run</a>{/if}
							{if $comment.correction}<span class="badge">rule</span>{/if}
						</div>
						<p>{$comment.body}</p>
					</li>
				{/foreach}
				</ol>
			</details>
		</li>
		{/if}
		<li class="refinement-comment" data-run-state="{$entry.comment.state}">
			<div class="refinement-comment-head">
				<strong>{$entry.comment.user}</strong>
				<span class="faint">{$entry.comment.created_at}</span>
				{if $entry.comment.note_id}<a class="badge" href="/notes#{$entry.comment.note_id}">the run</a>{/if}
				{if $entry.applied}<span class="badge signal">applied</span>{/if}
				{if $entry.comment.correction}<span class="badge">rule</span>{/if}
				{* Deleting the design deletes the thread: the rules correct it,
				   and without it they correct nothing. The artifacts stay. *}
				<form class="refinement-delete" method="post" action="{$base}/threads/{$thread.id}/comments/{$entry.comment.id}/delete"
					{if $entry.comment.id == $design}data-confirm="Delete the design, and the whole thread with it? Its rules and its runs go. The files stay in the artifacts, under {$thread.id}."{else}data-confirm="Delete this request? The next Apply all will not carry it. The files it produced stay."{/if}>
					<button class="btn-icon" type="submit" title="Delete" aria-label="Delete"><svg class="ic"><use href="#i-trash"/></svg></button>
				</form>
			</div>
			<p>{$entry.comment.body}</p>
		</li>
	{/foreach}
	</ol>

	{* One run over every rule at once, rather than one run per rule over
	   what the run before it left. *}
	<form class="form refinement-apply" method="post" action="{$base}/threads/{$thread.id}/apply">
		<p class="faint">Folds every request above into one and runs it over the files as they stand, so the result satisfies all of them together rather than one after another.</p>
		<button class="btn" type="submit"><svg class="ic"><use href="#i-collapse"/></svg>Apply all</button>
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
