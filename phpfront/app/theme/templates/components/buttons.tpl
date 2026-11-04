{* input scope: data_buttons. See buttons.md. *}
{foreach $data_buttons.items as $button}{if $button.bare}	<button type="button" class="btn-icon"{if $button.attrs} {button.attrs|raw}{/if}><svg class="ic"><use href="#i-{button.icon}"/></svg></button>
{elseif $button.faint}	<span class="faint run-denied">{button.faint}</span>
{elseif $button.href}	<a class="btn{if $data_buttons.small || $button.small} btn-sm{/if}{if $button.primary} btn-primary{/if}{if $button.danger} btn-danger{/if}" href="{button.href}"{if $button.download} download{/if}{if $button.attrs} {button.attrs|raw}{/if}>{if $button.icon}<svg class="ic"><use href="#i-{button.icon}"/></svg>{/if}{button.label}</a>
{else}	<button type="button" class="btn{if $data_buttons.small || $button.small} btn-sm{/if}{if $button.primary} btn-primary{/if}{if $button.danger} btn-danger{/if}"{if $button.disabled} disabled{if $button.note} title="{button.note}"{/if}{/if}{if $button.attrs} {button.attrs|raw}{/if}>{if $button.icon}<svg class="ic"><use href="#i-{button.icon}"/></svg>{/if}{button.label}</button>
{/if}{/foreach}
