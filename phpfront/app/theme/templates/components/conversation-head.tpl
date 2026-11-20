<div class="conversation-head">
	<div class="conversation-title">
		<h2>{data_head.title}</h2>
{if $data_head.meta}		<p class="faint">{data_head.meta}</p>
{/if}	</div>
	<div class="actions">
{if $data_head.can_pick}		<form class="agent-pick" data-agent-form data-thread="{data_head.id}">
			<label class="faint" for="thread-agent">Agent</label>
			<select id="thread-agent" name="agent">
{foreach $data_head.agents as $name}				<option value="{name}"{if $name == $data_head.agent} selected{/if}>{name}</option>
{/foreach}			</select>
			<button class="btn" type="submit">Save</button>
		</form>
{/if}		{eval $data_menu = $data_head.menu}{include components/menu.tpl}
	</div>
</div>
