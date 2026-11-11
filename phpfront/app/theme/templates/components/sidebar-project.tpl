				<div class="project{if $data_project.active} active{/if}" data-project="{data_project.id}">
					<a class="section{if $data_project.here} active{/if}" href="{base}{data_project.href}" data-folder="{data_project.id}" title="{data_project.title}">
{eval $data_icon = $data_project}{include components/sidebar-icon.tpl}<span>{data_project.name}</span>
{if $data_project.bad}<small class="bad" title="{data_project.badTitle}">{data_project.badge}</small>{elseif $data_project.counted}<small>{data_project.count}</small>{/if}
					</a>
				</div>
