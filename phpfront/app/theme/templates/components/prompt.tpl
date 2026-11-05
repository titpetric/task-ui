<details class="promptbox"{if $data_prompt.open} open{/if}{if $data_prompt.hidden} hidden{/if}>
	<summary><h3 class="section-title"><svg class="ic chev"><use href="#i-chev"/></svg>{data_prompt.title} <small>{data_prompt.note}</small></h3></summary>
{if !$data_prompt.watch}	<form class="note-form" data-api="{data_prompt.api}" data-refresh-notes>
{foreach $data_prompt.fields as $field}		<input type="hidden" name="{field.name}" value="{field.value}">
{/foreach}		<textarea name="prompt" rows="3" required spellcheck="false" placeholder="{data_prompt.placeholder}" aria-label="The prompt"></textarea>
		<div class="form-foot note"><span class="faint run-in">{data_prompt.lead}{if $data_prompt.options.users}<select name="user" aria-label="Runs as">{eval $data_options = $data_prompt.options}{include components/user-options.tpl}		</select>{else}you{/if}{data_prompt.tail}{if $data_prompt.picks}<select name="folder" aria-label="Run in">{foreach $data_prompt.picks as $pick}<option value="{pick.value}">{pick.text}</option>{/foreach}</select>{/if}{data_prompt.after}</span><button class="btn btn-primary" type="submit"><svg class="ic"><use href="#i-play"/></svg>Run prompt</button></div>
	</form>
{/if}	<div class="notes prompt-threads" data-notes data-dir="{data_prompt.place}"{if $data_prompt.area} data-area="{data_prompt.area}"{/if}>
{foreach $data_prompt.threads as $thread}{eval $data_note = $thread}{include components/note.tpl}{/foreach}	</div>
</details>
