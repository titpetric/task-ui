{if $data_workspace.top}
<details class="workspace{if $data_workspace.active} active{/if}" data-project="{data_workspace.id}" open>
	<summary{if $data_workspace.here} class="active"{/if} title="{data_workspace.title}">
		<svg class="ic chev"><use href="#i-chev"/></svg>
		<a class="section{if $data_workspace.here} active{/if}" href="{base}{data_workspace.href}" data-folder="{data_workspace.id}">
			<span>{data_workspace.name}</span>
{if $data_workspace.bad}<small class="bad" title="{data_workspace.badTitle}">{data_workspace.badge}</small>{else}<small>{data_workspace.count}</small>{/if}
		</a>
	</summary>
{if $data_workspace.jobs}
	<div class="jobs top">
{eval $data_jobs = $data_workspace}{include components/sidebar-jobs.tpl}
	</div>
{/if}
{foreach $data_workspace.projects as $project}{eval $data_project = $project}{include components/sidebar-project.tpl}{/foreach}
</details>
{else}
{foreach $data_workspace.projects as $project}{eval $data_project = $project}{include components/sidebar-project.tpl}{/foreach}
{/if}
