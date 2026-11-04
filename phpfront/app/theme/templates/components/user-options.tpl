{* input scope: data_options. See user-options.md. *}
{foreach $data_options.users as $option}			<option value="{option.name}"{if $option.name == $data_options.selected} selected{/if} data-keys="{option.keys}">{option.name}{if !$data_options.short} ({option.role}){/if} · {option.said}</option>
{/foreach}