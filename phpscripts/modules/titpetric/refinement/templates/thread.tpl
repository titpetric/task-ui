<section class="refinement">
	<p class="refinement-back"><a href="{$base}">&larr; Every thread</a></p>

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
				<form class="refinement-delete" method="post" action="{$base}/threads/{$thread.id}/comments/{$entry.comment.id}/delete"
					{if $entry.comment.id == $design}data-confirm="Delete the design, and the whole thread with it? Its rules and its runs go. The files stay in the artifacts, under {$thread.id}."{else}data-confirm="Delete this request? The files it produced stay."{/if}>
					<button class="btn-icon" type="submit" title="Delete" aria-label="Delete"><svg class="ic"><use href="#i-trash"/></svg></button>
				</form>
			</div>
			<div class="refinement-body">{$entry.comment.html|raw}</div>
		</li>
	{/foreach}
	</ol>

	<form class="form refinement-apply" method="post" action="{$base}/threads/{$thread.id}/apply">
		<button class="btn" type="submit"><svg class="ic"><use href="#i-collapse"/></svg>Apply all</button>
	</form>

	<form class="form refinement-fork" method="post" action="{$base}/threads/{$thread.id}/fork">
		<label for="refinement-fork">Fork</label>
		<textarea id="refinement-fork" name="want" rows="2" placeholder="A deployment topology diagram: the hosts, what runs on each and the links between them." required></textarea>
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
