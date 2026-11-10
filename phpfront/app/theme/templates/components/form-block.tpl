{if $data_form.err}<div class="alert bad" role="alert">{data_form.err}</div>{else}<form class="note-form run-form" data-run-form{if $data_form.folder} data-folder="{data_form.folder}"{/if}{if $data_form.dir} data-dir="{data_form.dir}"{/if}{if $data_form.job} data-job="{data_form.job}"{/if}{if $data_form.prompt} data-prompt="{data_form.prompt}"{/if}>
{if $data_form.title}<p class="form-title">{data_form.title}</p>
{/if}{foreach $data_form.fields as $field}{if $field.component == "hidden"}<input type="hidden" name="{field.key}" value="{field.value}" data-key="{field.key}" data-component="hidden"{if $field.when} data-when="{field.when}"{/if}{if $field.required} data-required{/if}>
{else}<div class="field" data-key="{field.key}" data-component="{field.component}"{if $field.when} data-when="{field.when}"{/if}{if $field.required} data-required{/if}{if $field.hidden} hidden{/if}>
{if $field.component == "bool"}<label class="check"><input type="checkbox" name="{field.key}" data-type="toggle"{if $field.checked} checked{/if}> {field.label}</label>
{elseif $field.component == "multi"}<span class="label">{field.label}</span>
<div class="checks">{foreach $field.options as $option}<label class="check"><input type="checkbox" name="{field.key}" value="{option.value}" data-type="set"{if $option.selected} checked{/if}{if $option.next} data-next="{option.next}"{/if}> {option.label}</label>{/foreach}</div>
{elseif $field.component == "textarea"}<span class="label">{field.label}</span>
<textarea name="{field.key}" rows="4" spellcheck="false" aria-label="{field.label}"{if $field.placeholder} placeholder="{field.placeholder}"{/if}{if $field.required} required{/if}>{field.value}</textarea>
{elseif $field.component == "number"}<span class="label">{field.label}</span>
<input type="number" name="{field.key}" value="{field.value}" aria-label="{field.label}"{if $field.placeholder} placeholder="{field.placeholder}"{/if}{if $field.required} required{/if}>
{elseif $field.component == "text"}<span class="label">{field.label}</span>
<input type="text" name="{field.key}" value="{field.value}" aria-label="{field.label}"{if $field.placeholder} placeholder="{field.placeholder}"{/if}{if $field.required} required{/if}>
{else}<span class="label">{field.label}</span>
<select name="{field.key}" aria-label="{field.label}"{if $field.required} required{/if}>{if !$field.required}<option value=""></option>{/if}{foreach $field.options as $option}<option value="{option.value}"{if $option.next} data-next="{option.next}"{/if}{if $option.selected} selected{/if}>{option.label}</option>{/foreach}</select>
{/if}{if $field.help}<small class="faint">{field.help}</small>
{/if}</div>
{/if}{/foreach}<div class="form-foot">{if $data_form.job}<span class="faint form-runs">Runs {if $data_form.folder}{data_form.folder}/{/if}{data_form.job}</span>{/if}<span class="faint form-state"></span><button class="btn btn-primary" type="submit">{data_form.label}</button></div>
</form>{/if}
