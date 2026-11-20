{if $data_status.working}<div class="conversation-status" data-status data-state="{data_status.state}" data-started="{data_status.started}">
	<p aria-live="polite"><i class="dot working"></i> <span data-status-text>{data_status.label}</span> <span class="faint" data-status-elapsed></span></p>
{if $data_status.can_end}	<div class="actions">
		<button class="btn btn-sm" type="button" data-stop="{data_status.run}" title="Stop this run and let it settle">Stop</button>
		<button class="btn btn-sm btn-danger" type="button" data-kill="{data_status.run}" title="Kill this run and everything it started">Kill</button>
	</div>
{/if}</div>
{/if}
