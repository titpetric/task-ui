{eval $data_head = $data.head}
{include components/page-head.tpl}

<section class="refinement">
	<form class="form refinement-start" method="post" action="{base}/refinement/threads">
		<p class="faint">Start a new thread: what to design. {data.help|raw}</p>
		{eval $data_markdown = $data.editor}{include components/editor.tpl}
		<div class="form-foot">
			<button class="btn btn-primary" type="submit">Start a thread</button>
		</div>
	</form>
{if $data.table.body}{eval $data_table = $data.table}
{include components/table.tpl}
{else}	<p class="faint">No threads yet.</p>
{/if}</section>
