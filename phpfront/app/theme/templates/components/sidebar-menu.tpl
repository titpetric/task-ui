			<nav class="menu">
{foreach $data as $item}
				<a href="{item.href}" class="{if $item.active}active{/if}"><svg class="ic"><use href="#i-{item.icon}"/></svg><span>{item.title}</span>{if $item.pulse}<i class="dot{if $chrome.running} running{/if}" data-any-running></i>{/if}</a>
{if $item.flows && $chrome.workflows}{eval $data_workflows = $chrome.workflows}				<div class="flows">{include components/sidebar-workflows.tpl}				</div>
{/if}
{/foreach}
			</nav>
