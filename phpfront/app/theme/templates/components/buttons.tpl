{foreach $data_buttons.items as $button}{if $button.bare}	<button type="button" class="btn-icon"{if $button.attrs} {button.attrs|raw}{/if}><i class="ic ic-{button.icon}" aria-hidden="true"></i></button>
{elseif $button.faint}	<span class="faint run-denied">{button.faint}</span>
{elseif $button.href}	<a class="{button.className}" href="{button.href}"{if $button.download} download{/if}{if $button.attrs} {button.attrs|raw}{/if}>{if $button.icon}<i class="ic ic-{button.icon}" aria-hidden="true"></i>{/if}{button.label}</a>
{else}	<button type="button" class="{button.className}"{if $button.disabled} disabled{if $button.note} title="{button.note}"{/if}{/if}{if $button.attrs} {button.attrs|raw}{/if}>{if $button.icon}<i class="ic ic-{button.icon}" aria-hidden="true"></i>{/if}{button.label}</button>
{/if}{/foreach}
