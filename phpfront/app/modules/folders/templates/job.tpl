{* The page of one job: what it is, every run of it, and what runs it on a schedule. The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}
{if $data.running}<div class="alert signal">
	<span class="state"><i class="dot running"></i><b>Running</b></span>, started by <b>{data.running.who}</b> {data.running.ago}{if $data.running.watching}, {data.running.watching} watching{/if}.{if $data.running.kills} Attach to follow it, or kill it.
{else} Attach to follow it.
{/if}</div>
{/if}{if $data.denied}<div class="alert">You can see this job and its history, but you don't have permission to run it.</div>
{/if}{if $data.summary}<pre class="summary">{data.summary}</pre>
{/if}<section class="block">
	<h3>Job</h3>
	<div class="table-wrap">
	<table class="table table-kv">
		<tbody>
			<tr><th>Folder</th><td><a href="{data.job.folderHref}"><svg class="ic"><use href="#i-folder"/></svg> {data.job.folder}</a> <span class="dim">{data.job.title}</span></td></tr>
{if $data.job.dir}			<tr><th>Directory</th><td><a href="{data.job.dirHref}"><svg class="ic"><use href="#i-folder-open"/></svg> {data.job.dir}/</a> <span class="dim">the job runs here</span></td></tr>
{/if}			<tr><th>Source</th><td><svg class="ic"><use href="#i-file"/></svg> {data.job.source}{if $data.job.shows}{eval $data_source = $data.job.shows}{include components/show-source.tpl}{/if}</td></tr>
{if $data.job.includes}			<tr><th>Includes</th><td>{foreach $data.job.includes as $read}<div><svg class="ic"><use href="#i-file"/></svg> {read.file}{eval $data_source = $read.shows}{include components/show-source.tpl}</div>{/foreach}</td></tr>
{/if}			<tr><th>Command</th><td><code>{data.job.command}</code></td></tr>
{if $data.job.depends}			<tr><th>Depends on</th><td>{foreach $data.job.depends as $dep}{if $dep.href}<a href="{dep.href}">{dep.name}</a>{else}{dep.text}{/if} {/foreach}</td></tr>
{/if}			<tr><th>Steps</th><td class="steps">{foreach $data.job.steps as $step}<code>{step}</code>{else}<span class="faint">none listed</span>{/foreach}</td></tr>
{if $data.job.aliases}			<tr><th>Aliases</th><td>{foreach $data.job.aliases as $alias}<span class="badge">{alias}</span> {/foreach}</td></tr>
{/if}			<tr><th>Who can run</th><td>{data.job.who}</td></tr>
			<tr><th>Hours</th><td>{eval $data_hours = $data.job.hours}{include components/hours.tpl}</td></tr>
			<tr><th>Input</th><td>{if $data.job.interactive}the job takes keyboard input{else}output only{/if}</td></tr>
		</tbody>
	</table>
	</div>
</section>
<section class="block" data-live>
	<h3>History{if $data.count} <small>{data.count} runs</small>{/if}</h3>
{if $data.activity}{eval $data_chart = $data.activity}{include components/chart.tpl}
{/if}	<h4>Runs <small>newest first</small></h4>
	<div class="table-wrap">
	<table class="table cards">
		<thead><tr><th>Started</th><th>Runs as</th><th>Duration</th><th>Result</th><th></th></tr></thead>
		<tbody>
{foreach $data.runs as $run}			<tr>
				<td class="nowrap key" title="{run.ago}">{run.stamp}</td>
				<td data-label="By">{eval $data_who = $run.who}{include components/who.tpl}</td>
				<td class="nowrap" data-label="Took">{run.duration}</td>
				<td data-label="Result">{eval $data_state = $run.state}{include components/state.tpl}</td>
				<td class="actions">
					<a class="btn btn-sm" href="{run.href}" data-attach="{run.id}">{run.attach}</a>{if $run.pending} <span class="faint">summarizing</span>{/if}
{eval $data_menu = $run.menu}
{include components/menu.tpl}
				</td>
			</tr>
{if $run.summary}			<tr class="summary-row">
				<td colspan="5"><details class="summary-details">
					<summary>Summary{if $run.summary.by} <span class="faint">asked by {run.summary.by}</span>{/if}</summary>
					<div class="markdown-doc">{run.summary.body|raw}</div>
					<a class="btn btn-sm" href="{run.summary.href}">Report</a>
				</details></td>
			</tr>
{/if}{else}			<tr class="empty"><td colspan="5">No runs yet.{if $data.can} Press Run to start the first one.{/if}</td></tr>
{/foreach}		</tbody>
	</table>
	</div>
</section>
{if $data.schedules}<section class="block">
{eval $data_block = $data.schedules.block}{include components/block-head.tpl}
{eval $data_sched = $data.schedules.table}{include components/sched-table.tpl}
{if $data.can}	<form class="form inline" data-api="/api/schedules" data-method="POST" data-next="reload" data-notice="Schedule added" data-sched>
		<input type="hidden" name="folder" value="{data.schedules.folder}">
		<input type="hidden" name="job" value="{data.schedules.job}">
		<div class="field whole">
			<label class="check"><input type="checkbox" id="sch-triage" name="triage"> Triage failures with Claude</label>
		<small class="faint">When the run fails, Claude reads the run in its folder, fixes the cause when it can, replies on the run's page and posts a note.</small>
		</div>
{eval $data_runas = $data.schedules.runAs}{include components/run-as-field.tpl}{include components/sched-builder.tpl}	</form>
{/if}</section>
{/if}
