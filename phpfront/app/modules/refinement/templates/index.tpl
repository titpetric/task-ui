{* The refinement board. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}

<section class="refinement">
	<form class="form refinement-start" method="post" action="{base}/refinement/threads">
		<label for="refinement-want">Start a new thread: what to design</label>
		<p class="faint" id="refinement-want-help">{data.help|raw}</p>
		<textarea id="refinement-want" name="want" rows="5" aria-describedby="refinement-want-help" placeholder="{data.want}" required></textarea>
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Start a thread</button>
		</div>
	</form>
{if $data.table.body}{eval $data_table = $data.table}
{include components/table.tpl}
{else}	<p class="faint">No threads yet.</p>
{/if}</section>
