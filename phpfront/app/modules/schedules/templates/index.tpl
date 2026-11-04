{* Scheduled: the form that adds a schedule of a job, and what runs on one.
   The scope is {data}; the chrome is {chrome}. *}
{eval $data_head = $data.head}
{include components/page-head.tpl}

{if $data.any}<section class="block" id="add-schedule">
<h3>Add schedule</h3>
<form class="form inline" data-api="/api/schedules" data-method="POST" data-next="reload" data-notice="Schedule added" data-sched>
	<div class="field">
		<label for="sch-folder">Folder</label>
		<select id="sch-folder" name="folder" data-narrows="sch-job" required>
{foreach $data.folders as $folder}			<option value="{folder.id}"{if $folder.selected} selected{/if}>{folder.id}</option>
{/foreach}		</select>
	</div>
	<div class="field">
		<label for="sch-job">Job</label>
		<select id="sch-job" name="job" required>
{foreach $data.jobs as $job}			<option value="{job.id}" data-folder="{job.folder}">{job.name}</option>
{/foreach}{eval $data_hostjobs = $data.hosts}{include components/host-jobs.tpl}		</select>
	</div>
	<div class="field">
		<label for="sch-args">Args</label>
		<input type="text" id="sch-args" name="args" data-type="args" placeholder="df -h /" spellcheck="false" autocomplete="off">
		<small class="faint">For a job that takes them; the command of a host.</small>
	</div>
	<div class="field whole">
		<label class="check"><input type="checkbox" id="sch-triage" name="triage"> Triage failures with Claude</label>
		<small class="faint">When the run fails, Claude reads the run in its folder, fixes the cause when it can, replies on the run's page and posts a note.</small>
	</div>
{eval $data_runas = $data.runAs}{include components/run-as-field.tpl}{include components/sched-builder.tpl}</form>
</section>
{/if}
<section class="block">
{eval $data_block = $data.block}{include components/block-head.tpl}
{if $data.schedules}	<div class="table-wrap fit">
	<table class="table fit-cards">
		<thead><tr><th>Folder</th><th>Job</th><th>Expression</th><th>Next run</th><th class="hide-narrow">Last run</th><th>Runs as</th><th>State</th><th></th></tr></thead>
		<tbody>
{foreach $data.schedules as $row}			<tr>
				<td data-label="Folder">{if $row.workflow}<svg class="ic lead"><use href="#i-flow"/></svg>workflow{else}<a href="{base}{row.folderHref}">{row.folder}</a>{/if}</td>
				<td class="key" data-label="Job">{if $row.jobHref !== ""}<a href="{base}{row.jobHref}">{row.job}</a>{else}{row.job}{/if}{if $row.args !== ""} <code>{row.args}</code>{/if}</td>
				<td data-label="Expression"><code class="nowrap">{row.expression}</code></td>
				<td data-label="Next run">{if $row.on}{foreach $row.next as $part}<span class="nowrap">{part}</span> {/foreach}{else}<span class="faint">-</span>{/if}</td>
				<td class="nowrap hide-narrow" data-label="Last run">{if $row.last !== ""}{row.last}{else}<span class="faint">never</span>{/if}</td>
				<td data-label="Runs as">{if $row.pick}<form class="run-as" data-api="{row.api}" data-method="PUT" data-next="reload" data-notice="Schedule saved">{eval $data_avatar = $row.user}{include components/avatar.tpl} <select name="user" data-submit aria-label="Runs as">
{eval $data_options = $row.options}{include components/user-options.tpl}		</select>{eval $data_menu = $row.user.menu}{include components/menu.tpl}</form>{else}{eval $data_user = $row.user}{include components/user.tpl}{/if}</td>
				<td data-label="State">{if $row.on}<span class="badge ok">enabled</span>{else}<span class="badge">disabled</span>{/if}{if $row.triage} <span class="badge signal" title="Claude triages a failed run">triage</span>{/if}</td>
				<td class="actions">{eval $data_menu = $row.menu}{include components/menu.tpl}</td>
			</tr>
{/foreach}		</tbody>
	</table>
	</div>
{/if}</section>
