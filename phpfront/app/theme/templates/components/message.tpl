<li class="{data_message.className}" id="note-{data_message.id}" data-run-id="{data_message.run}">
	<header>
		<span class="author">{eval $data_user = $data_message.author}{include components/user.tpl}</span>
		<span class="faint" title="{data_message.stamp}">{data_message.ago}</span>
		{if $data_message.agent}			<span class="badge agent" title="The agent's answer">agent</span>
{/if}{if $data_message.state}			<span class="badge note-state {data_message.state}">{data_message.state}</span>
{/if}{if $data_message.run}			<a class="faint mono" href="{base}/runs/{data_message.run}" title="The run this message comes from">run {data_message.short}</a>
{/if}	</header>
{if $data_message.body}	<div class="said-body prose">{data_message.body|raw}</div>
{/if}{if $data_message.working}{eval $data_working = $data_message.working_view}{include components/working.tpl}
{else}{foreach $data_message.work as $run}{eval $data_work = $run}{include components/work.tpl}
{/foreach}{/if}{if $data_message.audio}	<audio class="said-audio" controls preload="none" src="{data_message.audio}"></audio>
{/if}</li>
