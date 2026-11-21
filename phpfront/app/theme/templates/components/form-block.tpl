{if $data_form.err}<div class="alert bad" role="alert">{data_form.err}</div>{else}<form class="note-form run-form" data-run-form data-note="{data_form.note}" data-block="{data_form.index}"{if $data_form.folder} data-folder="{data_form.folder}"{/if}{if $data_form.dir} data-dir="{data_form.dir}"{/if}{if $data_form.job} data-job="{data_form.job}"{/if}{if $data_form.prompt} data-prompt="{data_form.prompt}"{/if}>
{if $data_form.title}<p class="form-title">{data_form.title}</p>
{/if}{if $data_form.subtitle}<p class="form-subtitle faint">{data_form.subtitle}</p>
{/if}{if $data_form.context}<div class="form-context prose">{data_form.context}</div>
{/if}{if $data_form.prerequisites}<ul class="form-prerequisites faint">{foreach $data_form.prerequisites as $item}<li>{item}</li>{/foreach}</ul>
{/if}{foreach $data_form.sections as $section}{if $section.title}<h3 class="form-section">{section.title}</h3>
{/if}{if $section.description}<p class="faint form-section-note">{section.description}</p>
{/if}{foreach $section.fields as $field}{if $field.component == "hidden"}<input type="hidden" name="{field.key}" value="{field.value}" data-key="{field.key}" data-component="hidden"{if $field.when} data-when="{field.when}"{/if}{if $field.required} data-required{/if}>
{else}<div class="field" data-key="{field.key}" data-component="{field.component}"{if $field.kind} data-kind="{field.kind}"{/if}{if $field.when} data-when="{field.when}"{/if}{if $field.required} data-required{/if}{if $field.hidden} hidden{/if}>
{if $field.component == "bool"}<label class="check"><input type="checkbox" name="{field.key}" data-type="toggle"{if $field.checked} checked{/if}> {field.label}</label>
{elseif $field.component == "multi"}<span class="label">{field.label}</span>
<div class="checks">{foreach $field.options as $option}<label class="check"><input type="checkbox" name="{field.key}" value="{option.value}" data-type="set"{if $option.selected} checked{/if}{if $option.next} data-next="{option.next}"{/if}> <span>{option.label}</span>{if $option.recommended} <span class="badge signal">recommended</span>{/if}{if $option.rationale}<small class="faint">{option.rationale}</small>{/if}</label>{/foreach}</div>
{elseif $field.component == "checklist"}<span class="label">{field.label}</span>
<div class="checks">{foreach $field.options as $option}<label class="check"><input type="checkbox" name="{field.key}" value="{option.value}" data-type="set"{if $option.selected} checked{/if}{if $option.next} data-next="{option.next}"{/if}> <span>{option.label}</span>{if $option.recommended} <span class="badge signal">recommended</span>{/if}{if $option.rationale}<small class="faint">{option.rationale}</small>{/if}</label>{/foreach}</div>
{elseif $field.kind == "qcm_free_text"}<span class="label">{field.label}</span>
<div class="checks">{foreach $field.options as $option}<label class="check"><input type="radio" name="{field.key}" value="{option.value}"{if $option.selected} checked{/if}{if $option.next} data-next="{option.next}"{/if}> <span>{option.label}</span>{if $option.recommended} <span class="badge signal">recommended</span>{/if}{if $option.rationale}<small class="faint">{option.rationale}</small>{/if}</label>{/foreach}</div>
{elseif $field.component == "textarea"}<span class="label">{field.label}</span>
<textarea name="{field.key}" rows="4" spellcheck="false" aria-label="{field.label}"{if $field.placeholder} placeholder="{field.placeholder}"{/if}{if $field.required} required{/if}>{field.value}</textarea>
{elseif $field.component == "number"}<span class="label">{field.label}</span>
<input type="number" name="{field.key}" value="{field.value}" aria-label="{field.label}"{if $field.placeholder} placeholder="{field.placeholder}"{/if}{if $field.required} required{/if}>
{elseif $field.component == "text"}<span class="label">{field.label}</span>
<input type="text" name="{field.key}" value="{field.value}" aria-label="{field.label}"{if $field.placeholder} placeholder="{field.placeholder}"{/if}{if $field.required} required{/if}>
{else}<span class="label">{field.label}</span>
<select name="{field.key}" aria-label="{field.label}"{if $field.required} required{/if}>{if !$field.required}<option value=""></option>{/if}{foreach $field.options as $option}<option value="{option.value}"{if $option.next} data-next="{option.next}"{/if}{if $option.selected} selected{/if}>{option.label}</option>{/foreach}</select>
{/if}{if $field.free_text}<input class="free-text" type="text" name="{field.key}.text" data-free-text maxlength="{field.max_free_text_bytes}" placeholder="{field.free_text_placeholder}" aria-label="{field.free_text_label}">
{/if}{if $field.help}<small class="faint">{field.help}</small>
{/if}</div>
{/if}{/foreach}{/foreach}<div class="form-note"><textarea name="general_note" rows="2" placeholder="Anything else? (optional)" aria-label="A note with the answers"></textarea></div>
<div class="form-foot">{if $data_form.job}<span class="faint form-runs">Runs {if $data_form.folder}{data_form.folder}/{/if}{data_form.job}</span>{/if}<span class="faint form-state"></span><button class="btn btn-primary" type="submit">{data_form.label}</button></div>
</form>{/if}
