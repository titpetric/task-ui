<form class="composer" data-send="{data_composer.mode}"{if $data_composer.url} data-send-url="{data_composer.url}"{/if}{if $data_composer.next} data-next="{data_composer.next}"{/if}{if $data_composer.pane} data-pane{/if}>
{if $data_composer.parent}	<input type="hidden" name="parent" value="{data_composer.parent}">
{/if}{if $data_composer.channel}	<input type="hidden" name="channel" value="{data_composer.channel}">
{/if}{if $data_composer.folder}	<input type="hidden" name="folder" value="{data_composer.folder}">
{/if}{if $data_composer.dir}	<input type="hidden" name="dir" value="{data_composer.dir}">
{/if}{if $data_composer.run}	<input type="hidden" name="run" value="{data_composer.run}">
{/if}{if $data_composer.area}	<input type="hidden" name="area" value="{data_composer.area}">
{/if}{if $data_composer.origin}	<input type="hidden" name="origin" value="{data_composer.origin}">
{/if}	<div class="composer-box">
		{eval $data_markdown = $data_composer.editor}{include components/editor.tpl}
		<div class="composer-foot">
{if $data_composer.modes}			<div class="composer-modes" role="group" aria-label="What to send">
{foreach $data_composer.modes as $mode}				<button type="button" class="btn btn-sm" data-composer-mode="{mode.key}" aria-pressed="{if $mode.key == $data_composer.mode}true{else}false{/if}">{mode.label}</button>
{/foreach}			</div>
{/if}{if $data_composer.agents}			<label class="composer-field" data-composer-agent>Agent
				<select name="agent" aria-label="Answers as">{foreach $data_composer.agents as $name}<option value="{name}"{if $name == $data_composer.agent} selected{/if}>{name}</option>{/foreach}</select>
			</label>
{/if}{if $data_composer.show_place}			<label class="composer-field" data-composer-place>In
				<input type="text" name="place" placeholder="folder/dir, empty is the root" spellcheck="false" autocomplete="off"{if $data_composer.places} list="composer-places"{/if}>
{if $data_composer.places}				<datalist id="composer-places">{foreach $data_composer.places as $place}<option value="{place}">{/foreach}</datalist>
{/if}			</label>
{/if}			<span class="composer-space"></span>
{if $data_composer.mic}			<button class="btn btn-icon" type="button" data-voice-reply data-parent="{data_composer.parent}" data-channel="{data_composer.channel}" title="Record a message" aria-label="Record a message"><i class="ic ic-mic" aria-hidden="true"></i></button>
{/if}			<button class="btn btn-primary" type="submit" title="Enter sends, Shift+Enter adds a line">Send</button>
		</div>
	</div>
</form>
