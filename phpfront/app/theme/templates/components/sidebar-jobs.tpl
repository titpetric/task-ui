{foreach $data_jobs.jobs as $job}
<div class="job{if $job.pinned} pinned{/if}{if $job.active} active{/if}" data-job-key="{job.key}">
	<i class="shebang{if $job.running} running{/if}" aria-hidden="true">#!</i>
	<a href="{base}{job.href}"{if $job.hint} title="{job.hint}"{/if}>{if $job.kind}<i>{job.kind}:</i>{/if}{job.label}</a>
{if $job.running}
	<button type="button" class="play" data-attach="{job.attach}" title="Attach to {job.name}, running" aria-label="Attach to {job.name}"><svg class="ic"><use href="#i-eye"/></svg></button>
{elseif $job.can}
	<button type="button" class="play" data-run data-folder="{job.folder}"{if $job.dir} data-dir="{job.dir}"{/if} data-job="{job.job}"{if $job.args} data-run-args="{job.args}"{/if} data-title="{job.folder} / {job.said}" title="Run {job.said}" aria-label="Run {job.said}"><svg class="ic"><use href="#i-play"/></svg></button>
{/if}
</div>
{/foreach}
