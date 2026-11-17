<form class="form" data-api="{data.api}" data-method="{data.method}" data-next="{data.next}" data-sched>
{if $data.edit}	<div class="field">
		<span class="label">Target</span>
		<p>{if $data.target === "workflow"}<svg class="ic lead"><use href="#i-flow"/></svg>{if $data.workflowHref}<a href="{data.workflowHref}">{data.workflowName}</a>{else}{data.workflowName}{/if}{else}{if $data.targetHref}<a href="{data.targetHref}">{data.folder} / {data.job}</a>{else}{data.folder} / {data.job}{/if}{/if}</p>
		<small class="faint">What a schedule is for cannot change; delete it and add another.</small>
	</div>
{if $data.target === "workflow"}	<input type="hidden" name="workflow" value="{data.workflow}">
{if $data.pickUser}{eval $data_runas = $data.runAs}{include components/run-as-field.tpl}{/if}
{else}	<input type="hidden" name="folder" value="{data.folder}">
	<input type="hidden" name="job" value="{data.job}">
	<div class="field">
		<label for="sch-args">Args</label>
		<input type="text" id="sch-args" name="args" data-type="args" value="{data.args}" placeholder="df -h /" spellcheck="false" autocomplete="off">
		<small class="faint">For a job that takes them; the command of a host.</small>
	</div>
	<div class="field whole">
		<label class="check"><input type="checkbox" id="sch-triage" name="triage"{if $data.triage} checked{/if}> Triage failures with Claude</label>
		<small class="faint">When the run fails, Claude reads the run in its folder, fixes the cause when it can, replies on the run's page and posts a note.</small>
	</div>
{if $data.pickUser}{eval $data_runas = $data.runAs}{include components/run-as-field.tpl}{/if}
{/if}
{else}	<div class="field">
		<label for="sch-workflow">Workflow</label>
		<select id="sch-workflow" name="workflow" data-sched-workflow>
			<option value="">A job</option>
{foreach $data.workflows as $flow}			<option value="{flow.id}"{if $flow.id === $data.workflow} selected{/if}>{flow.name}</option>
{/foreach}		</select>
		<small class="faint">Pick a workflow to schedule its steps instead of one job.</small>
	</div>
	<div data-sched-job>
		<div class="field">
			<label for="sch-folder">Folder</label>
			<select id="sch-folder" name="folder" data-narrows="sch-job" required>
{foreach $data.folders as $folder}				<option value="{folder.id}"{if $folder.selected} selected{/if}>{folder.id}</option>
{/foreach}			</select>
		</div>
		<div class="field">
			<label for="sch-job">Job</label>
			<select id="sch-job" name="job" required>
{foreach $data.jobs as $job}				<option value="{job.id}" data-folder="{job.folder}"{if $job.selected} selected{/if}>{job.name}</option>
{/foreach}{eval $data_hostjobs = $data.hosts}{include components/host-jobs.tpl}			</select>
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
{eval $data_runas = $data.runAs}{include components/run-as-field.tpl}	</div>
{/if}
{include components/sched-builder.tpl}
	<div class="field">
		<label class="check"><input type="checkbox" name="enabled"{if $data.enabled} checked{/if}> Enabled</label>
	</div>
	<div class="form-foot">
		<a class="btn" href="{data.back}">Cancel</a>
		<button class="btn btn-primary" type="submit">{data.label}</button>
	</div>
</form>
