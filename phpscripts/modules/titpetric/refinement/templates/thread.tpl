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

	{* One link, not a list. The thread's files are a folder of the
	   artifact tree and the artifacts view is what browses one, with the
	   preview, the history and the download the module would otherwise
	   be building a second time. *}
	<p class="refinement-files">
		<a class="btn btn-sm" href="/artifacts?path={$thread.id}">The files of this thread</a>
	</p>

	<ol class="refinement-thread">
	{foreach $comments as $comment}
		<li class="refinement-comment" data-run-state="{$comment.state}">
			<div class="refinement-comment-head">
				<strong>{$comment.user}</strong>
				<span class="faint">{$comment.created_at}</span>
				{if $comment.note_id}<a class="badge" href="/notes#{$comment.note_id}">the run</a>{/if}
				{if $comment.correction}<span class="badge">correction</span>{/if}
			</div>
			<p>{$comment.body}</p>
		</li>
	{/foreach}
	</ol>

	<form class="form refinement-comment-form" method="post" action="{$base}/threads/{$thread.id}/comments">
		<label for="refinement-comment">Refine it</label>
		<p class="faint" id="refinement-comment-help">One change per comment. Each one starts a run over the artifacts as they stand now.</p>
		<textarea id="refinement-comment" name="comment" rows="3" aria-describedby="refinement-comment-help" placeholder="use ULID as primary keys" required></textarea>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Send</button>
		</div>
	</form>
</section>
