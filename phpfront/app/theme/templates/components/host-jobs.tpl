{if $data_hostjobs}			<optgroup label="Hosts">
{foreach $data_hostjobs as $job}				<option value="{job.id}" data-folder="{job.folder}" title="{job.title}"{if $job.selected} selected{/if}>{job.name}</option>
{/foreach}			</optgroup>
{/if}