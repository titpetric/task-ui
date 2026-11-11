{eval $data_head = $data.head}
{include components/page-head.tpl}

<section class="refinement">
{if $data.visual}	<div class="refinement-output">{data.output|raw}</div>
{if $data.thinking}	<details class="refinement-thinking">
		<summary>show thinking</summary>
		<div class="refinement-report">{data.thinking|raw}</div>
	</details>
{/if}{elseif $data.thinking}	<div class="block refinement-report">
		<h3>The design <small>as the last run left it</small></h3>
		{data.thinking|raw}
	</div>
{else}	<p class="faint">The first run has not written its report yet. It appears here once the run ends.</p>
{/if}
	<ol class="refinement-thread">
{foreach $data.entries as $entry}{if $entry.folded_count}		<li class="refinement-folded">
			<details>
				<summary>{entry.folded_count} folded into the request below</summary>
				<ol>
{foreach $entry.folded as $comment}					<li class="refinement-comment">
						<div class="refinement-comment-head">
							<strong>{comment.user}</strong>
							<span class="faint">{comment.created_at}</span>
							{if $comment.run}<span class="state" data-run-id="{comment.run}"></span>{/if}{if $comment.note_id}<a class="badge" href="{base}/notes#{comment.note_id}">the run</a>{/if}
							{if $comment.correction}<span class="badge">rule</span>{/if}
						</div>
						<div class="refinement-body">{comment.html|raw}</div>
					</li>
{/foreach}				</ol>
			</details>
		</li>
{/if}		<li class="refinement-comment">
			<div class="refinement-comment-head">
				<strong>{entry.comment.user}</strong>
				<span class="faint">{entry.comment.created_at}</span>
				{if $entry.comment.run}<span class="state" data-run-id="{entry.comment.run}"></span>{/if}{if $entry.comment.note_id}<a class="badge" href="{base}/notes#{entry.comment.note_id}">the run</a>{/if}
				{if $entry.applied}<span class="badge signal">applied</span>{/if}
				{if $entry.comment.correction}<span class="badge">rule</span>{/if}
				<form class="refinement-delete" method="post" action="{base}/refinement/threads/{data.thread.id}/comments/{entry.comment.id}/delete"
					{if $entry.comment.id == $data.design}data-confirm="Delete the design, and the whole thread with it? Its rules and its runs go. The files stay in the artifacts, under {data.thread.id}."{else}data-confirm="Delete this request? The files it produced stay."{/if}>
					<button class="btn-icon" type="submit" title="Delete" aria-label="Delete"><svg class="ic"><use href="#i-trash"/></svg></button>
				</form>
			</div>
			<div class="refinement-body">{entry.comment.html|raw}</div>
		</li>
{/foreach}	</ol>

	<form class="form refinement-apply" method="post" action="{base}/refinement/threads/{data.thread.id}/apply">
		<button class="btn" type="submit"><svg class="ic"><use href="#i-collapse"/></svg>Apply all</button>
	</form>

	<form class="form refinement-fork" method="post" action="{base}/refinement/threads/{data.thread.id}/fork">
		<p class="faint">Fork: a thread of its own, from this design.</p>
		{eval $data_markdown = $data.fork}{include components/editor.tpl}
		<div class="form-foot">
			<button class="btn" type="submit"><svg class="ic"><use href="#i-flow"/></svg>Fork</button>
		</div>
	</form>

	<form class="form refinement-comment-form" method="post" action="{base}/refinement/threads/{data.thread.id}/comments">
		<p class="faint">Refine it. One change per comment, and each one starts a run over the artifacts as they stand now.</p>
		{eval $data_markdown = $data.comment}{include components/editor.tpl}
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Send</button>
		</div>
	</form>
</section>
