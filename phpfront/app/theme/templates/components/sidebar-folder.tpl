<details class="folder{if $data_folder.active} active{/if}{if !$data_folder.jobs} leaf{/if}" data-folder="{data_folder.id}"{if $data_folder.open} open{/if}>
	<summary title="{data_folder.title}">
		<a href="{base}{data_folder.href}">{eval $data_icon = $data_folder}{include components/sidebar-icon.tpl}<span>{if $data_folder.dir}<i>{data_folder.dir}/</i>{/if}<b>{data_folder.name}</b></span></a>
		<i class="dot"></i>
{if $data_folder.bad}<small class="bad" title="{data_folder.badTitle}">{data_folder.badge}</small>{else}<small>{data_folder.count}</small>{/if}
{if $data_folder.jobs}		<svg class="ic chev"><use href="#i-chev"/></svg>
{/if}	</summary>
{if $data_folder.jobs}	<div class="jobs">
{eval $data_jobs = $data_folder}{include components/sidebar-jobs.tpl}
	</div>
{/if}</details>
