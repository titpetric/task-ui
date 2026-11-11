{if !$data_menu.empty}<details class="more-menu">
	<summary class="btn-icon" aria-haspopup="menu" aria-expanded="false" title="{data_menu.label}" aria-label="{data_menu.label}"><svg class="ic"><use href="#i-more"/></svg></summary>
	<div class="more-list">
{if $data_menu.details}		<dl class="more-details">
{foreach $data_menu.details as $detail}			<dt>{detail.label}</dt><dd{if $detail.mono} class="mono"{/if}{if $detail.title} title="{detail.title}"{/if}>{detail.value}</dd>
{/foreach}		</dl>
{/if}		<div class="more-items" role="menu" aria-label="{data_menu.label}">
{foreach $data_menu.groups as $index => $group}{if $index}			<hr class="more-sep">
{/if}{foreach $group as $item}{if $item.href}			<a class="more-item{if $item.danger} danger{/if}" role="menuitem" href="{item.href}"{if $item.download} download{/if}{if $item.attrs} {item.attrs|raw}{/if}><svg class="ic"><use href="#i-{item.icon}"/></svg><span>{item.label}</span>{if $item.note}<small class="faint">{item.note}</small>{/if}</a>
{else}			<button type="button" class="more-item{if $item.danger} danger{/if}" role="menuitem"{if $item.disabled} disabled{/if}{if $item.attrs} {item.attrs|raw}{/if}><svg class="ic"><use href="#i-{item.icon}"/></svg><span>{item.label}</span>{if $item.note}<small class="faint">{item.note}</small>{/if}</button>
{/if}{/foreach}{/foreach}
		</div>
	</div>
</details>{/if}
