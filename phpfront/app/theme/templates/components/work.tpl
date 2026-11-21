{if $data_work.stats || $data_work.trace.states}<div class="work-run" data-run-id="{data_work.id}">
{if $data_work.trace.states}	<button type="button" class="work-trace" data-trace data-mode="duration" data-ms="{data_work.trace.ms}" data-count="{data_work.trace.count}" title="Switch the bar between duration, count and time">
		<span class="work-trace-mode" data-trace-mode>Duration</span>
		<span class="work-trace-bar">
{foreach $data_work.trace.states as $state}<span class="work-trace-group {state.key}" style="width:{state.share}%" data-share="{state.share}" data-count="{state.count}" data-label="{state.label}" data-tip="{state.label}: {state.share}%"></span>
{/foreach}{foreach $data_work.trace.segments as $segment}<span class="work-trace-event {segment.key}" data-ms="{segment.ms}" data-label="{segment.label}" data-tip="{segment.label}: {segment.span}"></span>
{/foreach}		</span>
	</button>
{/if}{if $data_work.stats}	<div class="work-stats">{data_work.stats}</div>
{/if}</div>{/if}
