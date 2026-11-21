{if $data_status.working}<div class="conversation-status" data-status data-state="{data_status.state}" data-started="{data_status.started}">
	<p aria-live="polite"><i class="dot working"></i> <span data-status-text>{data_status.label}</span> <span class="faint" data-status-elapsed></span></p>
{if $data_status.agents}	<ul class="status-agents" data-status-agents>
{foreach $data_status.agents as $agent}		<li><details class="status-agent {agent.state}" data-agent="{agent.id}">
			<summary><i class="ic ic-users" aria-hidden="true"></i> <span>{agent.task}</span>{if $agent.kind} <small class="faint">{agent.kind}</small>{/if} <span class="badge {if $agent.state == 'working'}signal{elseif $agent.state == 'failed'}bad{else}ok{/if}" data-status-agent-state>{agent.state}</span></summary>
{if $agent.prompt}			<pre class="status-agent-prompt">{agent.prompt}</pre>
{/if}{if $agent.result}			<pre class="status-agent-result">{agent.result}</pre>
{/if}		</details></li>
{/foreach}	</ul>
{/if}{if $data_status.can_end}	<div class="actions">
		<button class="btn-icon" type="button" data-stop="{data_status.run}" title="Stop this run and let it settle" aria-label="Stop this run"><i class="ic ic-stop" aria-hidden="true"></i></button>
		<button class="btn-icon danger" type="button" data-kill="{data_status.run}" title="Kill this run and everything it started" aria-label="Kill this run and everything it started"><i class="ic ic-trash" aria-hidden="true"></i></button>
	</div>
{/if}</div>
{/if}