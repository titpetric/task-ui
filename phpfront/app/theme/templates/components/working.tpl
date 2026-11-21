<div class="said-live" data-status data-state="{data_working.state}" data-started="{data_working.started}">
	<div class="{data_working.progressClass}" data-work-progress>{if $data_working.progressHtml}{data_working.progressHtml|raw}{/if}</div>
	<p class="said-state" aria-live="polite"><i class="dot working"></i> <span data-status-text>{data_working.label}</span> <span class="faint" data-status-elapsed></span>{if $data_working.canEnd}<span class="actions"><button class="btn-icon" type="button" data-stop="{data_working.run}" title="Stop this run and let it settle" aria-label="Stop this run"><i class="ic ic-stop" aria-hidden="true"></i></button><button class="btn-icon danger" type="button" data-kill="{data_working.run}" title="Kill this run and everything it started" aria-label="Kill this run and everything it started"><i class="ic ic-trash" aria-hidden="true"></i></button></span>{/if}</p>
{if $data_working.agents}	<ul class="status-agents" data-status-agents>
{foreach $data_working.agents as $agent}		<li><details class="{agent.className}" data-agent="{agent.id}">
			<summary><i class="ic ic-users" aria-hidden="true"></i> <span>{agent.task}</span>{if $agent.kind} <small class="faint">{agent.kind}</small>{/if} <span class="{agent.badge}" data-status-agent-state>{agent.state}</span></summary>
{if $agent.prompt}			<pre class="status-agent-prompt">{agent.prompt}</pre>
{/if}{if $agent.result}			<pre class="status-agent-result">{agent.result}</pre>
{/if}		</details></li>
{/foreach}	</ul>
{/if}	<p class="said-live-foot"><span class="work-stats" data-work-stats data-tools="{data_working.tools}">{data_working.stats}</span></p>
</div>
