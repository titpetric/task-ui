<li class="said{if $data_message.mine} mine{/if}{if $data_message.agent} agent{/if}{if $data_message.working} working{/if}" id="note-{data_message.id}">
	<header>
		<span class="author">{eval $data_user = $data_message.author}{include components/user.tpl}</span>
		<span class="faint" title="{data_message.stamp}">{data_message.ago}</span>
		{if $data_message.agent}			<span class="badge agent" title="The agent's answer">agent</span>
{/if}{if $data_message.state}			<span class="badge note-state {data_message.state}">{data_message.state}</span>
{/if}{if $data_message.run}			<a class="faint mono" href="{base}/runs/{data_message.run}" title="The run this message comes from">run {data_message.short}</a>
{/if}	</header>
	<div class="said-body prose">{data_message.body|raw}</div>
{foreach $data_message.work as $run}{eval $data_work = $run}{include components/work.tpl}
{/foreach}{if $data_message.audio}	<audio class="said-audio" controls preload="none" src="{data_message.audio}"></audio>
{/if}</li>
