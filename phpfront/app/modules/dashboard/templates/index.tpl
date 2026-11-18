{eval $data_head = $data.head}
{include components/page-head.tpl}
{if $data.note}<div class="alert bad" role="alert"><b>{data.note.pipelines} {data.note.says} errors</b> in {foreach $data.note.workspaces as $i => $ws}{if $i}, {/if}<a href="{ws.href}">{ws.name}</a>{/foreach}. {data.note.lists} them.</div>
{/if}
<div data-sse="runs" data-dashboard>
{if $data.threads}<section class="block">
	<h3>Working threads <small>Claude conversations in progress</small></h3>
	<div class="tiles" data-threads>
{foreach $data.threads as $thread}		<a class="tile thread-tile" href="{thread.href}" data-thread-tile="{thread.id}">
			<div class="tile-head">
				<i class="dot {thread.state}"></i>
				<b>{thread.name}</b>
				<span class="faint mono">{thread.place}</span>
				<span class="spacer"></span>
				{eval $data_avatar = $thread.who}{include components/avatar.tpl}
			</div>
			<div class="frame faint">Claude is working in the thread.</div>
		</a>
{/foreach}	</div>
</section>
{/if}
<section class="block">
	<h3>Running now <small>one run per job, up to a run per CPU; terminals take no turn</small></h3>
	<div class="tiles" data-tiles>
{foreach $data.runs as $run}		<a class="tile" href="{run.href}" data-attach="{run.id}" data-tile="{run.id}">
			<div class="tile-head">
				<i class="dot {run.state}"></i>
				<b>{run.name}</b>
				<span class="term-title" data-term-title{if $run.title} title="{run.title}"{/if}>{run.title}</span>
				<span class="spacer"></span>
				<span class="faint">{run.faint}</span>
				{eval $data_avatar = $run.who}{include components/avatar.tpl}
			</div>
			<div class="frame"{if $run.running} data-frame="{run.id}"{/if} role="img" aria-label="Terminal of {run.name}">{run.frame|raw}</div>
		</a>
{/foreach}	</div>
	<div class="alert" data-tiles-empty{if $data.runs} hidden{/if}>Nothing is running. Open a folder and press Run, or run a workflow, and its terminal appears here for everyone.</div>
</section>

</div>
