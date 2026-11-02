{* input scope: data. See sidebar-menu.md. *}
			<nav class="menu">
{foreach $data as $item}
				<a href="{base}{item.href}" class="{if $item.active}active{/if}"><svg class="ic"><use href="#i-{item.icon}"/></svg><span>{item.title}</span>{if $item.pulse}<i class="dot" data-any-running></i>{/if}</a>
{/foreach}
			</nav>
