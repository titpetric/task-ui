{eval $data_head = $data.head}
{include components/page-head.tpl}
{if $data.note}<div class="alert bad" role="alert"><b>{data.note.pipelines} {data.note.says} errors</b> in {foreach $data.note.workspaces as $i => $ws}{if $i}, {/if}<a href="{ws.href}">{ws.name}</a>{/foreach}. {data.note.lists} them.</div>
{/if}
<div data-sse="runs" data-dashboard>
<section class="block">
	<h3>Processes <small>one row per run in progress; the task opens where the work is done</small></h3>
	<div class="table-wrap">
	<table class="table">
		<thead><tr><th>PID</th><th>User</th><th>Stat</th><th>Start</th><th>Time</th><th>Task</th><th>Command</th></tr></thead>
		<tbody data-processes>
{foreach $data.runs as $run}			<tr data-run-id="{run.id}">
				<td class="mono nowrap" title="{run.id}">{run.pid}</td>
				<td class="nowrap">{run.user}</td>
				<td class="nowrap"><span class="state"><i class="dot {run.state}"></i>{run.state}</span></td>
				<td class="mono nowrap">{run.start}</td>
				<td class="mono nowrap" data-took data-started="{run.started}">{run.time}</td>
				<td class="key">{if $run.href}<a href="{run.href}">{run.task}</a>{else}{run.task}{/if}</td>
				<td class="desc">{run.command}</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
	<div class="alert" data-processes-empty{if $data.runs} hidden{/if}>Nothing is running. Start a job or ask a prompt and its row appears here.</div>
</section>
</div>
