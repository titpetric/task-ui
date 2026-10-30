<section class="refinement">
	<form class="form refinement-start" method="post" action="{$base}/threads">
		<label for="refinement-want">What to design</label>
		<p class="faint" id="refinement-want-help">One paragraph, and the thread starts from it. Every comment after it is a Claude run over the artifacts as they stand.</p>
		<textarea id="refinement-want" name="want" rows="5" aria-describedby="refinement-want-help" placeholder="A database schema with strict separation of user, user_login_method and user_login_method_password." required></textarea>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Start a thread</button>
		</div>
	</form>

	{if $threads}
	<div class="table-wrap">
		<table class="table">
			<thead><tr><th>Thread</th><th class="num">Comments</th><th>Started</th><th></th></tr></thead>
			<tbody>
			{foreach $threads as $thread}
				<tr>
					<td class="key"><a href="{$base}/threads/{$thread.id}">{$thread.title}</a></td>
					<td class="num mono">{$thread.comments}</td>
					<td class="nowrap faint">{$thread.created_at}</td>
					<td class="actions"><a class="btn btn-sm" href="{$base}/threads/{$thread.id}">Open</a></td>
				</tr>
			{/foreach}
			</tbody>
		</table>
	</div>
	{else}
	<p class="faint">No threads yet.</p>
	{/if}
</section>
