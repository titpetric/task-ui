{if $data_project.top}
				<details class="project{if $data_project.active} active{/if}" data-project="{data_project.id}" open>
					<summary{if $data_project.here} class="active"{/if} title="{data_project.title}">
						<a class="section{if $data_project.here} active{/if}" href="{base}{data_project.href}" data-folder="{data_project.id}">
{eval $data_icon = $data_project}{include components/sidebar-icon.tpl}<span>{data_project.name}</span>
{if $data_project.bad}<small class="bad" title="{data_project.badTitle}">{data_project.badge}</small>{elseif $data_project.counted}<small>{data_project.count}</small>{/if}
						</a>
						<svg class="ic chev"><use href="#i-chev"/></svg>
					</summary>
{if $data_project.jobs}
					<div class="jobs top">
{eval $data_jobs = $data_project}{include components/sidebar-jobs.tpl}
					</div>
{/if}
{foreach $data_project.nested as $folder}{eval $data_folder = $folder}{include components/sidebar-folder.tpl}{/foreach}
				</details>
{else}
{foreach $data_project.nested as $folder}{eval $data_folder = $folder}{include components/sidebar-folder.tpl}{/foreach}
{/if}
