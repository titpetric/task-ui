{* The workflows list. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}

{if $data.prompt}{eval $data_prompt = $data.prompt}
{include components/prompt.tpl}
{/if}
<section class="block">
{eval $data_block = $data.block}
{include components/block-head.tpl}
{if $data.rows}	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Workflow</th><th>Last run</th><th>By</th><th></th></tr></thead>
		<tbody>
{foreach $data.rows as $row}			<tr>
				<td class="key"><a href="{base}/workflows/{row.id}">{row.name}</a>{if $row.description}<small class="faint wf-desc">{row.description|raw}</small>{/if}{foreach $row.schedules as $hint}<small class="faint wf-desc" title="{hint.title}"><svg class="ic lead"><use href="#i-clock"/></svg><code>{hint.expression}</code>{if !$hint.enabled} disabled{/if}</small>{/foreach}</td>
				<td class="nowrap" data-label="Last run">{if $row.running}<span class="state"><i class="dot running"></i>running</span>{elseif $row.last}<span class="state"><i class="dot {row.state}"></i>{row.state}</span> <span class="faint" title="{row.stamp}">{row.ago}</span>{else}<span class="faint">never run</span>{/if}</td>
				<td data-label="By">{eval $data_user = $row.user}{include components/user.tpl}{if $row.runAs} <span class="faint nowrap">· runs as {row.runAs}</span>{/if}</td>
				<td class="actions">{if $row.canRun}{if $row.running}
					<button type="button" class="btn btn-sm" data-workflow-stop="{row.running}"><svg class="ic"><use href="#i-stop"/></svg>Stop</button>{else}
					<button type="button" class="btn btn-primary btn-sm" data-workflow-run="{row.id}"{if $row.prompted} data-workflow-prompt="{row.name}"{/if}><svg class="ic"><use href="#i-play"/></svg>Run</button>{/if}{elseif $row.runDenied}
					<span class="faint run-denied">{row.runDenied}</span>{/if}{eval $data_menu = $row.menu}{include components/menu.tpl}
				</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
{/if}</section>
<section class="block">
{eval $data_block = $data.runsBlock}
{include components/block-head.tpl}
{if $data.runs}	<div class="table-wrap">
	<table class="table cards wf-runs">
		<thead><tr><th>Run</th><th>By</th><th>Started</th></tr></thead>
		<tbody>
{foreach $data.runs as $run}			<tr data-run-id="{run.id}">
				<td class="key">{eval $data_wffold = $run.fold}{include components/wffold.tpl}<a href="{base}/workflows/{run.workflowId}/runs/{run.id}" title="{run.title}"><i class="dot {run.state}"></i>{run.name}</a></td>
				<td data-label="By">{eval $data_user = $run.user}{include components/user.tpl}</td>
				<td class="nowrap" data-label="Started" title="{run.stamp}">{run.ago}</td>
			</tr>
{eval $data_wfanswer = $run.answer}
{include components/wfanswer.tpl}
{/foreach}		</tbody>
	</table>
	</div>
{/if}</section>
