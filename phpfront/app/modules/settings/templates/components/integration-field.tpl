{if $data_field.kind === "who"}							<div class="field wf-wide">
								<span class="label">{data_field.label}</span>{eval $data_whofield = $data_field.who}{include components/who-field.tpl}{if $data_field.help}<small class="faint">{data_field.help}</small>{/if}
							</div>
{elseif $data_field.kind === "env"}							<div class="field wf-wide env-flag">
								<div class="field-head">
									<span class="label">{data_field.label}</span>{eval $data_switch = $data_field.switch}{include components/integration-switch.tpl}								</div>
								<small class="faint">{if $data_field.help}{data_field.help}{if $data_field.spaced} {/if}{/if}<code>{data_field.env}</code></small>
							</div>
{else}							<div class="field{if $data_field.wide} wf-wide{/if}">
{if $data_field.toggle}								<div class="field-head">
									<label for="{data_field.id}">{data_field.label}{if $data_field.badge} <span class="badge{if $data_field.tone} {data_field.tone}{/if}">{data_field.badge}</span>{/if}</label>{eval $data_switch = $data_field.switch}{include components/integration-switch.tpl}								</div>
{else}								<label for="{data_field.id}">{data_field.label}{if $data_field.badge} <span class="badge{if $data_field.tone} {data_field.tone}{/if}">{data_field.badge}</span>{/if}</label>
{/if}{if $data_field.kind === "secret"}								<input type="password" id="{data_field.id}" name="{data_field.name}" placeholder="{data_field.placeholder}" autocomplete="new-password" spellcheck="false">
{elseif $data_field.kind === "select"}								<select id="{data_field.id}" name="{data_field.name}">
{foreach $data_field.options as $option}									<option value="{option.value}"{if $option.selected} selected{/if}>{option.value}</option>
{/foreach}								</select>
{else}								<input type="text" id="{data_field.id}" name="{data_field.name}" value="{data_field.value}" placeholder="{data_field.placeholder}" autocomplete="off" spellcheck="false">
{/if}								<small class="faint">{if $data_field.help}{data_field.help}{if $data_field.spaced} {/if}{/if}{if $data_field.env}<code>{data_field.env}</code>{/if}</small>
							</div>
{/if}